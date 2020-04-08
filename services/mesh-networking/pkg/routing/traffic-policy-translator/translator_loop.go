package traffic_policy_translator

import (
	"context"
	"fmt"

	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/contextutils"
	core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	discovery_controller "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/controller"
	networking_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	networking_controller "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/controller"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/clients"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/clients/zephyr/discovery"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/clients/zephyr/networking"
	"github.com/solo-io/service-mesh-hub/pkg/logging"
	"github.com/solo-io/service-mesh-hub/pkg/selector"
	"github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/routing/traffic-policy-translator/errors"
	"github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/routing/traffic-policy-translator/preprocess"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewTrafficPolicyTranslatorLoop(
	preprocessor preprocess.TrafficPolicyPreprocessor,
	meshTranslators []TrafficPolicyMeshTranslator,
	meshClient zephyr_discovery.MeshClient,
	meshServiceClient zephyr_discovery.MeshServiceClient,
	trafficPolicyClient zephyr_networking.TrafficPolicyClient,
	trafficPolicyController networking_controller.TrafficPolicyController,
	meshServiceController discovery_controller.MeshServiceController,
) TrafficPolicyTranslatorLoop {
	return &trafficPolicyTranslatorLoop{
		preprocessor:            preprocessor,
		meshTranslators:         meshTranslators,
		meshClient:              meshClient,
		meshServiceClient:       meshServiceClient,
		trafficPolicyClient:     trafficPolicyClient,
		trafficPolicyController: trafficPolicyController,
		meshServiceController:   meshServiceController,
	}
}

type trafficPolicyTranslatorLoop struct {
	preprocessor            preprocess.TrafficPolicyPreprocessor
	meshTranslators         []TrafficPolicyMeshTranslator
	meshClient              zephyr_discovery.MeshClient
	meshServiceClient       zephyr_discovery.MeshServiceClient
	trafficPolicyClient     zephyr_networking.TrafficPolicyClient
	trafficPolicyController networking_controller.TrafficPolicyController
	meshServiceController   discovery_controller.MeshServiceController
}

func (t *trafficPolicyTranslatorLoop) Start(ctx context.Context) error {
	err := t.trafficPolicyController.AddEventHandler(ctx, &networking_controller.TrafficPolicyEventHandlerFuncs{
		OnCreate: func(trafficPolicy *networking_v1alpha1.TrafficPolicy) error {
			logger := logging.BuildEventLogger(ctx, logging.CreateEvent, trafficPolicy)
			logger.Debugw("event handler enter",
				zap.Any("spec", trafficPolicy.Spec),
				zap.Any("status", trafficPolicy.Status),
			)
			translatorErrors, err := t.upsertPolicyResourcesForTrafficPolicy(ctx, trafficPolicy)
			t.setStatus(err, translatorErrors, trafficPolicy)
			err = t.trafficPolicyClient.UpdateStatus(ctx, trafficPolicy)
			if err != nil {
				logger.Errorw("Error while handling TrafficPolicy create event", err)
			}
			return nil
		},

		OnUpdate: func(old, new *networking_v1alpha1.TrafficPolicy) error {
			logger := logging.BuildEventLogger(ctx, logging.UpdateEvent, new)
			logger.Debugw("event handler enter",
				zap.Any("old_spec", old.Spec),
				zap.Any("old_status", old.Status),
				zap.Any("new_spec", new.Spec),
				zap.Any("new_status", new.Status),
			)
			translatorErrors, err := t.upsertPolicyResourcesForTrafficPolicy(ctx, new)
			t.setStatus(err, translatorErrors, new)
			err = t.trafficPolicyClient.UpdateStatus(ctx, new)
			if err != nil {
				logger.Errorw("Error while handling TrafficPolicy update event", err)
			}
			return nil
		},

		OnDelete: func(trafficPolicy *networking_v1alpha1.TrafficPolicy) error {
			logger := logging.BuildEventLogger(ctx, logging.DeleteEvent, trafficPolicy)
			logger.Debugf("Ignoring event for traffic policy: %s.%s", trafficPolicy.Name, trafficPolicy.Namespace)
			return nil
		},

		OnGeneric: func(trafficPolicy *networking_v1alpha1.TrafficPolicy) error {
			logging.BuildEventLogger(ctx, logging.GenericEvent, trafficPolicy).
				Debugf("Ignoring event for traffic policy: %s.%s", trafficPolicy.Name, trafficPolicy.Namespace)
			return nil
		},
	})

	if err != nil {
		return err
	}

	err = t.meshServiceController.AddEventHandler(ctx, &discovery_controller.MeshServiceEventHandlerFuncs{
		OnCreate: func(meshService *v1alpha1.MeshService) error {
			logger := logging.BuildEventLogger(ctx, logging.CreateEvent, meshService)
			logger.Debugw("event handler enter",
				zap.Any("spec", meshService.Spec),
				zap.Any("status", meshService.Status),
			)
			err := t.upsertPolicyResourcesForMeshService(ctx, meshService)
			if err != nil {
				logger.Errorw("Error while handling MeshService create event", err)
			}
			return nil
		},

		OnUpdate: func(old, new *v1alpha1.MeshService) error {
			logger := logging.BuildEventLogger(ctx, logging.UpdateEvent, new)
			logger.Debugw("event handler enter",
				zap.Any("old_spec", old.Spec),
				zap.Any("old_status", old.Status),
				zap.Any("new_spec", new.Spec),
				zap.Any("new_status", new.Status),
			)
			err := t.upsertPolicyResourcesForMeshService(ctx, new)
			if err != nil {
				logger.Errorw("Error while handling MeshService update event", err)
			}
			return nil
		},

		OnDelete: func(meshService *v1alpha1.MeshService) error {
			logger := logging.BuildEventLogger(ctx, logging.DeleteEvent, meshService)
			logger.Debugw("Ignoring event",
				zap.Any("spec", meshService.Spec),
				zap.Any("status", meshService.Status),
			)
			return nil
		},

		OnGeneric: func(meshService *v1alpha1.MeshService) error {
			logger := logging.BuildEventLogger(ctx, logging.DeleteEvent, meshService)
			logger.Debugw("Ignoring event",
				zap.Any("spec", meshService.Spec),
				zap.Any("status", meshService.Status),
			)
			return nil
		},
	})

	if err != nil {
		return err
	}
	return nil
}

// Compute and upsert all Mesh-specific configuration needed to reflect TrafficPolicy
func (t *trafficPolicyTranslatorLoop) upsertPolicyResourcesForTrafficPolicy(
	ctx context.Context,
	trafficPolicy *networking_v1alpha1.TrafficPolicy,
) ([]*types.TrafficPolicyStatus_TranslatorError, error) {
	mergedTrafficPoliciesByMeshService, err := t.preprocessor.PreprocessTrafficPolicy(ctx, trafficPolicy)
	if err != nil {
		return nil, err
	}
	return t.translateMergedTrafficPolicies(ctx, mergedTrafficPoliciesByMeshService)
}

// Compute and upsert all Mesh-specific configuration needed to reflect TrafficPolicies for the given MeshService
func (t *trafficPolicyTranslatorLoop) upsertPolicyResourcesForMeshService(
	ctx context.Context,
	meshService *v1alpha1.MeshService,
) error {
	mergedTrafficPoliciesByMeshService, err := t.preprocessor.PreprocessTrafficPoliciesForMeshService(ctx, meshService)
	if err != nil {
		return err
	}
	// ignore Mesh specific statuses because there's no triggering TrafficPolicy whose status we can update
	translatorErr, err := t.translateMergedTrafficPolicies(ctx, mergedTrafficPoliciesByMeshService)
	if translatorErr != nil {
		contextutils.LoggerFrom(ctx).Warnw("translator error", zap.Any("errors", translatorErr))
	}
	return err
}

func (t *trafficPolicyTranslatorLoop) translateMergedTrafficPolicies(
	ctx context.Context,
	mergedTrafficPoliciesByMeshService map[selector.MeshServiceId][]*networking_v1alpha1.TrafficPolicy,
) ([]*types.TrafficPolicyStatus_TranslatorError, error) {
	var meshTypeStatuses []*types.TrafficPolicyStatus_TranslatorError
	for meshServiceKey, mergedTrafficPolicies := range mergedTrafficPoliciesByMeshService {
		meshServiceObjectKey := client.ObjectKey{Name: meshServiceKey.Name, Namespace: meshServiceKey.Namespace}
		meshService, err := t.meshServiceClient.Get(ctx, meshServiceObjectKey)
		if err != nil {
			return nil, err
		}
		mesh, err := t.meshClient.Get(ctx, clients.ResourceRefToObjectKey(meshService.Spec.GetMesh()))
		if err != nil {
			return nil, err
		}
		for _, meshTranslator := range t.meshTranslators {
			translatorError := meshTranslator.TranslateTrafficPolicy(
				contextutils.WithLogger(ctx, meshTranslator.Name()),
				meshService,
				mesh,
				mergedTrafficPolicies,
			)
			if translatorError != nil {
				meshTypeStatuses = append(meshTypeStatuses, translatorError)
			}
		}
	}
	return meshTypeStatuses, nil
}

// err represents errors during processing prior to translation
// translatorErrors represent errors during translation to mesh-specific config
func (t *trafficPolicyTranslatorLoop) setStatus(
	err error,
	translatorErrors []*types.TrafficPolicyStatus_TranslatorError,
	trafficPolicy *networking_v1alpha1.TrafficPolicy) {
	if err != nil {
		// clear out any previous translator errors
		trafficPolicy.Status.TranslatorErrors = nil
		if eris.Is(err, errors.TrafficPolicyConflictError) {
			trafficPolicy.Status.TranslationStatus = &core_types.Status{
				State:   core_types.Status_CONFLICT,
				Message: "TrafficPolicy conflicts with existing set of TrafficPolicies",
			}
		} else {
			trafficPolicy.Status.TranslationStatus = &core_types.Status{
				State:   core_types.Status_PROCESSING_ERROR,
				Message: fmt.Sprintf("Error while processing TrafficPolicy: %s", err.Error()),
			}
		}
	} else if translatorErrors != nil {
		trafficPolicy.Status.TranslationStatus = &core_types.Status{
			State:   core_types.Status_PROCESSING_ERROR,
			Message: fmt.Sprintf("Error while translating TrafficPolicy, check Status.TranslatorErrors for details"),
		}
		trafficPolicy.Status.TranslatorErrors = translatorErrors
	} else {
		trafficPolicy.Status.TranslationStatus = &core_types.Status{
			State: core_types.Status_ACCEPTED,
		}
		// clear out any previous translator errors
		trafficPolicy.Status.TranslatorErrors = nil
	}
}
