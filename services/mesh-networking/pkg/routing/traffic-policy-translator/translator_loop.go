package traffic_policy_translator

import (
	"context"
	"fmt"

	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/contextutils"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_controller "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/controller"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	zephyr_networking_controller "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/controller"
	zephyr_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/clients"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/container-runtime"
	"github.com/solo-io/service-mesh-hub/pkg/kube/selection"
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
	TrafficPolicyEventWatcher zephyr_networking_controller.TrafficPolicyEventWatcher,
	MeshServiceEventWatcher zephyr_discovery_controller.MeshServiceEventWatcher,
) TrafficPolicyTranslatorLoop {
	return &trafficPolicyTranslatorLoop{
		preprocessor:              preprocessor,
		meshTranslators:           meshTranslators,
		meshClient:                meshClient,
		meshServiceClient:         meshServiceClient,
		trafficPolicyClient:       trafficPolicyClient,
		TrafficPolicyEventWatcher: TrafficPolicyEventWatcher,
		MeshServiceEventWatcher:   MeshServiceEventWatcher,
	}
}

type trafficPolicyTranslatorLoop struct {
	preprocessor              preprocess.TrafficPolicyPreprocessor
	meshTranslators           []TrafficPolicyMeshTranslator
	meshClient                zephyr_discovery.MeshClient
	meshServiceClient         zephyr_discovery.MeshServiceClient
	trafficPolicyClient       zephyr_networking.TrafficPolicyClient
	TrafficPolicyEventWatcher zephyr_networking_controller.TrafficPolicyEventWatcher
	MeshServiceEventWatcher   zephyr_discovery_controller.MeshServiceEventWatcher
}

func (t *trafficPolicyTranslatorLoop) Start(ctx context.Context) error {
	err := t.TrafficPolicyEventWatcher.AddEventHandler(ctx, &zephyr_networking_controller.TrafficPolicyEventHandlerFuncs{
		OnCreate: func(trafficPolicy *zephyr_networking.TrafficPolicy) error {
			logger := container_runtime.BuildEventLogger(ctx, container_runtime.CreateEvent, trafficPolicy)
			logger.Debugw("event handler enter",
				zap.Any("spec", trafficPolicy.Spec),
				zap.Any("status", trafficPolicy.Status),
			)
			translatorErrors, err := t.upsertPolicyResourcesForTrafficPolicy(ctx, trafficPolicy)
			t.setStatus(err, translatorErrors, trafficPolicy)
			err = t.trafficPolicyClient.UpdateTrafficPolicyStatus(ctx, trafficPolicy)
			if err != nil {
				logger.Errorw("Error while handling TrafficPolicy create event", err)
			}
			return nil
		},

		OnUpdate: func(old, new *zephyr_networking.TrafficPolicy) error {
			logger := container_runtime.BuildEventLogger(ctx, container_runtime.UpdateEvent, new)
			logger.Debugw("event handler enter",
				zap.Any("old_spec", old.Spec),
				zap.Any("old_status", old.Status),
				zap.Any("new_spec", new.Spec),
				zap.Any("new_status", new.Status),
			)
			translatorErrors, err := t.upsertPolicyResourcesForTrafficPolicy(ctx, new)
			t.setStatus(err, translatorErrors, new)
			err = t.trafficPolicyClient.UpdateTrafficPolicyStatus(ctx, new)
			if err != nil {
				logger.Errorw("Error while handling TrafficPolicy update event", err)
			}
			return nil
		},

		OnDelete: func(trafficPolicy *zephyr_networking.TrafficPolicy) error {
			logger := container_runtime.BuildEventLogger(ctx, container_runtime.DeleteEvent, trafficPolicy)
			logger.Debugf("Ignoring event for traffic policy: %s.%s", trafficPolicy.Name, trafficPolicy.Namespace)
			return nil
		},

		OnGeneric: func(trafficPolicy *zephyr_networking.TrafficPolicy) error {
			container_runtime.BuildEventLogger(ctx, container_runtime.GenericEvent, trafficPolicy).
				Debugf("Ignoring event for traffic policy: %s.%s", trafficPolicy.Name, trafficPolicy.Namespace)
			return nil
		},
	})

	if err != nil {
		return err
	}

	err = t.MeshServiceEventWatcher.AddEventHandler(ctx, &zephyr_discovery_controller.MeshServiceEventHandlerFuncs{
		OnCreate: func(meshService *zephyr_discovery.MeshService) error {
			logger := container_runtime.BuildEventLogger(ctx, container_runtime.CreateEvent, meshService)
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

		OnUpdate: func(old, new *zephyr_discovery.MeshService) error {
			logger := container_runtime.BuildEventLogger(ctx, container_runtime.UpdateEvent, new)
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

		OnDelete: func(meshService *zephyr_discovery.MeshService) error {
			logger := container_runtime.BuildEventLogger(ctx, container_runtime.DeleteEvent, meshService)
			logger.Debugw("Ignoring event",
				zap.Any("spec", meshService.Spec),
				zap.Any("status", meshService.Status),
			)
			return nil
		},

		OnGeneric: func(meshService *zephyr_discovery.MeshService) error {
			logger := container_runtime.BuildEventLogger(ctx, container_runtime.DeleteEvent, meshService)
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
	trafficPolicy *zephyr_networking.TrafficPolicy,
) ([]*zephyr_networking_types.TrafficPolicyStatus_TranslatorError, error) {
	mergedTrafficPoliciesByMeshService, err := t.preprocessor.PreprocessTrafficPolicy(ctx, trafficPolicy)
	if err != nil {
		return nil, err
	}
	return t.translateMergedTrafficPolicies(ctx, mergedTrafficPoliciesByMeshService)
}

// Compute and upsert all Mesh-specific configuration needed to reflect TrafficPolicies for the given MeshService
func (t *trafficPolicyTranslatorLoop) upsertPolicyResourcesForMeshService(
	ctx context.Context,
	meshService *zephyr_discovery.MeshService,
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
	mergedTrafficPoliciesByMeshService map[selection.MeshServiceId][]*zephyr_networking.TrafficPolicy,
) ([]*zephyr_networking_types.TrafficPolicyStatus_TranslatorError, error) {
	var meshTypeStatuses []*zephyr_networking_types.TrafficPolicyStatus_TranslatorError
	for meshServiceKey, mergedTrafficPolicies := range mergedTrafficPoliciesByMeshService {
		meshServiceObjectKey := client.ObjectKey{Name: meshServiceKey.Name, Namespace: meshServiceKey.Namespace}
		meshService, err := t.meshServiceClient.GetMeshService(ctx, meshServiceObjectKey)
		if err != nil {
			return nil, err
		}
		mesh, err := t.meshClient.GetMesh(ctx, clients.ResourceRefToObjectKey(meshService.Spec.GetMesh()))
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
	translatorErrors []*zephyr_networking_types.TrafficPolicyStatus_TranslatorError,
	trafficPolicy *zephyr_networking.TrafficPolicy) {
	if err != nil {
		// clear out any previous translator errors
		trafficPolicy.Status.TranslatorErrors = nil
		if eris.Is(err, errors.TrafficPolicyConflictError) {
			trafficPolicy.Status.TranslationStatus = &zephyr_core_types.Status{
				State:   zephyr_core_types.Status_CONFLICT,
				Message: "TrafficPolicy conflicts with existing set of TrafficPolicies",
			}
		} else {
			trafficPolicy.Status.TranslationStatus = &zephyr_core_types.Status{
				State:   zephyr_core_types.Status_PROCESSING_ERROR,
				Message: fmt.Sprintf("Error while processing TrafficPolicy: %s", err.Error()),
			}
		}
	} else if translatorErrors != nil {
		trafficPolicy.Status.TranslationStatus = &zephyr_core_types.Status{
			State:   zephyr_core_types.Status_PROCESSING_ERROR,
			Message: fmt.Sprintf("Error while translating TrafficPolicy, check Status.TranslatorErrors for details"),
		}
		trafficPolicy.Status.TranslatorErrors = translatorErrors
	} else {
		trafficPolicy.Status.TranslationStatus = &zephyr_core_types.Status{
			State: zephyr_core_types.Status_ACCEPTED,
		}
		// clear out any previous translator errors
		trafficPolicy.Status.TranslatorErrors = nil
	}
}
