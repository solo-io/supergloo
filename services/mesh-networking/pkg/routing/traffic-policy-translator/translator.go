package traffic_policy_translator

import (
	"context"
	"fmt"

	"github.com/rotisserie/eris"
	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	discovery_controller "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1/controller"
	networking_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	networking_controller "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1/controller"
	"github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/mesh-projects/pkg/clients/zephyr/discovery"
	zephyr_networking "github.com/solo-io/mesh-projects/pkg/clients/zephyr/networking"
	"github.com/solo-io/mesh-projects/pkg/logging"
	"github.com/solo-io/mesh-projects/services/mesh-networking/pkg/routing/traffic-policy-translator/errors"
	"github.com/solo-io/mesh-projects/services/mesh-networking/pkg/routing/traffic-policy-translator/keys"
	"github.com/solo-io/mesh-projects/services/mesh-networking/pkg/routing/traffic-policy-translator/preprocess"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewTrafficPolicyTranslator(
	ctx context.Context,
	preprocessor preprocess.TrafficPolicyPreprocessor,
	meshTranslators []TrafficPolicyMeshTranslator,
	meshClient zephyr_discovery.MeshClient,
	meshServiceClient zephyr_discovery.MeshServiceClient,
	trafficPolicyClient zephyr_networking.TrafficPolicyClient,
	trafficPolicyController networking_controller.TrafficPolicyController,
	meshServiceController discovery_controller.MeshServiceController,
) TrafficPolicyTranslator {
	return &trafficPolicyTranslator{
		ctx:                     ctx,
		preprocessor:            preprocessor,
		meshTranslators:         meshTranslators,
		meshClient:              meshClient,
		meshServiceClient:       meshServiceClient,
		trafficPolicyClient:     trafficPolicyClient,
		trafficPolicyController: trafficPolicyController,
		meshServiceController:   meshServiceController,
	}
}

type trafficPolicyTranslator struct {
	ctx                     context.Context
	preprocessor            preprocess.TrafficPolicyPreprocessor
	meshTranslators         []TrafficPolicyMeshTranslator
	meshClient              zephyr_discovery.MeshClient
	meshServiceClient       zephyr_discovery.MeshServiceClient
	trafficPolicyClient     zephyr_networking.TrafficPolicyClient
	trafficPolicyController networking_controller.TrafficPolicyController
	meshServiceController   discovery_controller.MeshServiceController
}

func (t *trafficPolicyTranslator) Start(ctx context.Context) error {
	err := t.trafficPolicyController.AddEventHandler(ctx, &networking_controller.TrafficPolicyEventHandlerFuncs{
		OnCreate: func(trafficPolicy *networking_v1alpha1.TrafficPolicy) error {
			logger := logging.BuildEventLogger(t.ctx, logging.CreateEvent, trafficPolicy)
			logger.Debugf("Handling event: %+v", trafficPolicy)
			translatorErrors, err := t.upsertPolicyResourcesForTrafficPolicy(trafficPolicy)
			t.setStatus(err, translatorErrors, trafficPolicy)
			err = t.trafficPolicyClient.UpdateStatus(t.ctx, trafficPolicy)
			if err != nil {
				logger.Error("Error while handling TrafficPolicy create event", err)
			}
			return nil
		},

		OnUpdate: func(_, new *networking_v1alpha1.TrafficPolicy) error {
			logger := logging.BuildEventLogger(t.ctx, logging.UpdateEvent, new)
			logger.Debugf("Handling event: %+v", new)
			translatorErrors, err := t.upsertPolicyResourcesForTrafficPolicy(new)
			t.setStatus(err, translatorErrors, new)
			err = t.trafficPolicyClient.UpdateStatus(t.ctx, new)
			if err != nil {
				logger.Error("Error while handling TrafficPolicy update event", err)
			}
			return nil
		},

		OnDelete: func(trafficPolicy *networking_v1alpha1.TrafficPolicy) error {
			logger := logging.BuildEventLogger(t.ctx, logging.DeleteEvent, trafficPolicy)
			logger.Warn("Ignoring event: %+v", trafficPolicy)
			return nil
		},

		OnGeneric: func(trafficPolicy *networking_v1alpha1.TrafficPolicy) error {
			logging.BuildEventLogger(t.ctx, logging.GenericEvent, trafficPolicy).
				Warn("Ignoring event: %+v", trafficPolicy)
			return nil
		},
	})

	if err != nil {
		return err
	}

	err = t.meshServiceController.AddEventHandler(ctx, &discovery_controller.MeshServiceEventHandlerFuncs{
		OnCreate: func(meshService *v1alpha1.MeshService) error {
			logger := logging.BuildEventLogger(ctx, logging.CreateEvent, meshService)
			logger.Debugf("Handling event: %+v", meshService)
			err := t.upsertPolicyResourcesForMeshService(meshService)
			if err != nil {
				logger.Error("Error while handling MeshService create event", err)
			}
			return nil
		},

		OnUpdate: func(_, new *v1alpha1.MeshService) error {
			logger := logging.BuildEventLogger(ctx, logging.UpdateEvent, new)
			logger.Debugf("Handling event: %+v", new)
			err := t.upsertPolicyResourcesForMeshService(new)
			if err != nil {
				logger.Error("Error while handling MeshService update event", err)
			}
			return nil
		},

		OnDelete: func(meshService *v1alpha1.MeshService) error {
			logger := logging.BuildEventLogger(ctx, logging.DeleteEvent, meshService)
			logger.Warn("Ignoring event: %+v", meshService)
			return nil
		},

		OnGeneric: func(meshService *v1alpha1.MeshService) error {
			logging.BuildEventLogger(ctx, logging.GenericEvent, meshService).
				Warn("Ignoring event: %+v", meshService)
			return nil
		},
	})

	if err != nil {
		return err
	}
	return nil
}

// Compute and upsert all Mesh-specific configuration needed to reflect TrafficPolicy
func (t *trafficPolicyTranslator) upsertPolicyResourcesForTrafficPolicy(
	trafficPolicy *networking_v1alpha1.TrafficPolicy,
) ([]*types.TrafficPolicyStatus_TranslatorError, error) {
	mergedTrafficPoliciesByMeshService, err := t.preprocessor.PreprocessTrafficPolicy(t.ctx, trafficPolicy)
	if err != nil {
		return nil, err
	}
	return t.translateMergedTrafficPolicies(mergedTrafficPoliciesByMeshService)
}

// Compute and upsert all Mesh-specific configuration needed to reflect TrafficPolicies for the given MeshService
func (t *trafficPolicyTranslator) upsertPolicyResourcesForMeshService(
	meshService *v1alpha1.MeshService,
) error {
	mergedTrafficPoliciesByMeshService, err := t.preprocessor.PreprocessTrafficPoliciesForMeshService(t.ctx, meshService)
	if err != nil {
		return err
	}
	// ignore Mesh specific statuses because there's no triggering TrafficPolicy whose status we can update
	_, err = t.translateMergedTrafficPolicies(mergedTrafficPoliciesByMeshService)
	return err
}

func (t *trafficPolicyTranslator) translateMergedTrafficPolicies(
	mergedTrafficPoliciesByMeshService map[keys.MeshServiceMultiClusterKey][]*networking_v1alpha1.TrafficPolicy,
) ([]*types.TrafficPolicyStatus_TranslatorError, error) {
	var meshTypeStatuses []*types.TrafficPolicyStatus_TranslatorError
	for meshServiceKey, mergedTrafficPolicies := range mergedTrafficPoliciesByMeshService {
		meshServiceObjectKey := client.ObjectKey{Name: meshServiceKey.Name, Namespace: meshServiceKey.Namespace}
		meshService, err := t.meshServiceClient.Get(t.ctx, meshServiceObjectKey)
		if err != nil {
			return nil, err
		}
		mesh, err := t.meshClient.Get(t.ctx, client.ObjectKey{
			Name:      meshService.Spec.GetMesh().GetName(),
			Namespace: meshService.Spec.GetMesh().GetNamespace()})
		if err != nil {
			return nil, err
		}
		for _, meshTranslator := range t.meshTranslators {
			translatorError := meshTranslator.TranslateTrafficPolicy(t.ctx, meshService, mesh, mergedTrafficPolicies)
			if translatorError != nil {
				meshTypeStatuses = append(meshTypeStatuses, translatorError)
			}
		}
	}
	return meshTypeStatuses, nil
}

// err represents errors during processing prior to translation
// translatorErrors represent errors during translation to mesh-specific config
func (t *trafficPolicyTranslator) setStatus(
	err error,
	translatorErrors []*types.TrafficPolicyStatus_TranslatorError,
	trafficPolicy *networking_v1alpha1.TrafficPolicy) {
	if err != nil {
		if eris.Is(err, errors.TrafficPolicyConflictError) {
			trafficPolicy.Status.ComputedStatus = &core_types.ComputedStatus{
				Status:  core_types.ComputedStatus_CONFLICT,
				Message: "TrafficPolicy conflicts with existing set of TrafficPolicies",
			}
		} else {
			trafficPolicy.Status.ComputedStatus = &core_types.ComputedStatus{
				Status:  core_types.ComputedStatus_PROCESSING_ERROR,
				Message: fmt.Sprintf("Error while processing TrafficPolicy: %s", err.Error()),
			}
		}
	} else if translatorErrors != nil {
		trafficPolicy.Status.ComputedStatus = &core_types.ComputedStatus{
			Status:  core_types.ComputedStatus_PROCESSING_ERROR,
			Message: fmt.Sprintf("Error while translating TrafficPolicy, check Status.TranslatorErrors for details"),
		}
		trafficPolicy.Status.TranslatorErrors = translatorErrors
	} else {
		trafficPolicy.Status.ComputedStatus = &core_types.ComputedStatus{
			Status: core_types.ComputedStatus_ACCEPTED,
		}
	}
}
