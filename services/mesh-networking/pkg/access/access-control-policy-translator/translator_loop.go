package acp_translator

import (
	"context"
	"fmt"

	"github.com/solo-io/go-utils/contextutils"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_controller "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/controller"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	zephyr_networking_controller "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/controller"
	zephyr_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/clients"
	"github.com/solo-io/service-mesh-hub/pkg/logging"
	"github.com/solo-io/service-mesh-hub/pkg/selector"
	"go.uber.org/zap"
)

func NewAcpTranslatorLoop(
	acpController zephyr_networking_controller.AccessControlPolicyEventWatcher,
	MeshServiceEventWatcher zephyr_discovery_controller.MeshServiceEventWatcher,
	meshClient zephyr_discovery.MeshClient,
	accessControlPolicyClient zephyr_networking.AccessControlPolicyClient,
	resourceSelector selector.ResourceSelector,
	meshTranslators []AcpMeshTranslator,
) AcpTranslatorLoop {
	return &translatorLoop{
		acpController:             acpController,
		MeshServiceEventWatcher:   MeshServiceEventWatcher,
		meshClient:                meshClient,
		accessControlPolicyClient: accessControlPolicyClient,
		resourceSelector:          resourceSelector,
		meshTranslators:           meshTranslators,
	}
}

type translatorLoop struct {
	acpController             zephyr_networking_controller.AccessControlPolicyEventWatcher
	MeshServiceEventWatcher   zephyr_discovery_controller.MeshServiceEventWatcher
	meshClient                zephyr_discovery.MeshClient
	accessControlPolicyClient zephyr_networking.AccessControlPolicyClient
	meshTranslators           []AcpMeshTranslator
	resourceSelector          selector.ResourceSelector
}

func (t *translatorLoop) Start(ctx context.Context) error {
	err := t.acpController.AddEventHandler(ctx, &zephyr_networking_controller.AccessControlPolicyEventHandlerFuncs{
		OnCreate: func(obj *zephyr_networking.AccessControlPolicy) error {
			logger := logging.BuildEventLogger(ctx, logging.CreateEvent, obj)
			logger.Debugw("event handler enter",
				zap.Any("spec", obj.Spec),
				zap.Any("status", obj.Status),
			)
			translatorErrors, err := t.translateAccessControlPolicy(ctx, obj)
			t.setStatus(err, translatorErrors, obj)
			err = t.accessControlPolicyClient.UpdateAccessControlPolicyStatus(ctx, obj)
			if err != nil {
				logger.Errorw("Error while handling AccessControlPolicy create event", err)
			}
			return nil
		},
		OnUpdate: func(old, new *zephyr_networking.AccessControlPolicy) error {
			logger := logging.BuildEventLogger(ctx, logging.UpdateEvent, new)
			logger.Debugw("event handler enter",
				zap.Any("old_spec", old.Spec),
				zap.Any("old_status", old.Status),
				zap.Any("new_spec", new.Spec),
				zap.Any("new_status", new.Status),
			)
			translatorErrors, err := t.translateAccessControlPolicy(ctx, new)
			t.setStatus(err, translatorErrors, new)
			err = t.accessControlPolicyClient.UpdateAccessControlPolicyStatus(ctx, new)
			if err != nil {
				logger.Errorw("Error while handling AccessControlPolicy update event", err)
			}
			return nil
		},
		OnDelete: func(obj *zephyr_networking.AccessControlPolicy) error {
			logger := logging.BuildEventLogger(ctx, logging.DeleteEvent, obj)
			logger.Debugw("ignoring event",
				zap.Any("spec", obj.Spec),
				zap.Any("status", obj.Status),
			)
			return nil
		},
		OnGeneric: func(obj *zephyr_networking.AccessControlPolicy) error {
			logger := logging.BuildEventLogger(ctx, logging.GenericEvent, obj)
			logger.Debugw("ignoring event",
				zap.Any("spec", obj.Spec),
				zap.Any("status", obj.Status),
			)
			return nil
		},
	})
	if err != nil {
		return err
	}
	return t.MeshServiceEventWatcher.AddEventHandler(ctx, &zephyr_discovery_controller.MeshServiceEventHandlerFuncs{
		OnCreate: func(obj *zephyr_discovery.MeshService) error {
			logger := logging.BuildEventLogger(ctx, logging.CreateEvent, obj)
			logger.Debugw("event handler enter",
				zap.Any("spec", obj.Spec),
				zap.Any("status", obj.Status),
			)
			translatorErrorsForACPs, err := t.translateACPsForMeshService(ctx, obj)
			// Update status for each ACP that was processed for MeshService
			for _, translatorErrWithACP := range translatorErrorsForACPs {
				t.setStatus(err, translatorErrWithACP.translatorErrors, translatorErrWithACP.accessControlPolicy)
				err = t.accessControlPolicyClient.UpdateAccessControlPolicyStatus(ctx, translatorErrWithACP.accessControlPolicy)
				if err != nil {
					logger.Errorw("Error while handling MeshService create event", err)
				}
			}
			return nil
		},
		OnUpdate: func(old, new *zephyr_discovery.MeshService) error {
			logger := logging.BuildEventLogger(ctx, logging.UpdateEvent, new)
			logger.Debugw("event handler enter",
				zap.Any("old_spec", old.Spec),
				zap.Any("old_status", old.Status),
				zap.Any("new_spec", new.Spec),
				zap.Any("new_status", new.Status),
			)
			translatorErrorsForACPs, err := t.translateACPsForMeshService(ctx, new)
			// Update status for each ACP that was processed for MeshService
			for _, translatorErrWithACP := range translatorErrorsForACPs {
				t.setStatus(err, translatorErrWithACP.translatorErrors, translatorErrWithACP.accessControlPolicy)
				err = t.accessControlPolicyClient.UpdateAccessControlPolicyStatus(ctx, translatorErrWithACP.accessControlPolicy)
				if err != nil {
					logger.Errorw("Error while handling MeshService create event", err)
				}
			}
			return nil
		},
		OnDelete: func(obj *zephyr_discovery.MeshService) error {
			logger := logging.BuildEventLogger(ctx, logging.DeleteEvent, obj)
			logger.Debugw("ignoring event",
				zap.Any("spec", obj.Spec),
				zap.Any("status", obj.Status),
			)
			return nil
		},
		OnGeneric: func(obj *zephyr_discovery.MeshService) error {
			logger := logging.BuildEventLogger(ctx, logging.GenericEvent, obj)
			logger.Debugw("ignoring event",
				zap.Any("spec", obj.Spec),
				zap.Any("status", obj.Status),
			)
			return nil
		},
	})
}

// Translate AccessControlPolicy to AuthorizationPolicy for all targeted k8s Services
func (t *translatorLoop) translateAccessControlPolicy(
	ctx context.Context,
	acp *zephyr_networking.AccessControlPolicy,
) ([]*zephyr_networking_types.AccessControlPolicyStatus_TranslatorError, error) {
	targetServices, err := t.getTargetServices(ctx, acp)
	if err != nil {
		return nil, err
	}
	var translatorErrors []*zephyr_networking_types.AccessControlPolicyStatus_TranslatorError
	for _, meshTranslator := range t.meshTranslators {

		translatorError := meshTranslator.Translate(
			contextutils.WithLogger(ctx, meshTranslator.Name()),
			targetServices,
			acp)

		if translatorError != nil {
			translatorErrors = append(translatorErrors, translatorError)
		}
	}
	return translatorErrors, nil
}

// For all AccessControlPolicies that apply to MeshService, reprocess (i.e. translate) that AccessControlPolicy
// for that MeshService to reflect any changes to the underlying k8s Service.
func (t *translatorLoop) translateACPsForMeshService(
	ctx context.Context,
	meshService *zephyr_discovery.MeshService,
) ([]translatorErrorForACP, error) {
	mesh, err := t.meshClient.GetMesh(ctx, clients.ResourceRefToObjectKey(meshService.Spec.GetMesh()))
	if err != nil {
		return nil, err
	}
	targetService := TargetService{
		MeshService: meshService,
		Mesh:        mesh,
	}
	acps, err := t.getApplicableAccessControlPolicies(ctx, meshService)
	if err != nil {
		return nil, err
	}
	var translatorErrorsForACPs []translatorErrorForACP
	for _, acp := range acps {
		var translatorErrors []*zephyr_networking_types.AccessControlPolicyStatus_TranslatorError
		for _, meshTranslator := range t.meshTranslators {
			translatorError := meshTranslator.Translate(
				contextutils.WithLogger(ctx, meshTranslator.Name()),
				[]TargetService{targetService},
				acp,
			)
			if translatorError != nil {
				translatorErrors = append(translatorErrors, translatorError)
			}
		}
		translatorErrorsForACPs = append(translatorErrorsForACPs, translatorErrorForACP{
			accessControlPolicy: acp,
			translatorErrors:    translatorErrors,
		})
	}
	return translatorErrorsForACPs, nil
}

// Get all destination services' MeshService and backing Mesh selected by the AccessControlPolicy
func (t *translatorLoop) getTargetServices(ctx context.Context, acp *zephyr_networking.AccessControlPolicy) ([]TargetService, error) {
	meshServices, err := t.resourceSelector.GetMeshServicesByServiceSelector(ctx, acp.Spec.GetDestinationSelector())
	if err != nil {
		return nil, err
	}
	var targetServices []TargetService
	for _, meshService := range meshServices {
		mesh, err := t.meshClient.GetMesh(ctx, clients.ResourceRefToObjectKey(meshService.Spec.GetMesh()))
		if err != nil {
			return nil, err
		}
		targetServices = append(targetServices, TargetService{
			MeshService: meshService,
			Mesh:        mesh,
		})
	}
	return targetServices, nil
}

// Get all AccessControlPolicies that are applicable to the given MeshService
func (t *translatorLoop) getApplicableAccessControlPolicies(
	ctx context.Context,
	meshService *zephyr_discovery.MeshService,
) ([]*zephyr_networking.AccessControlPolicy, error) {
	var applicableACPs []*zephyr_networking.AccessControlPolicy
	acpList, err := t.accessControlPolicyClient.ListAccessControlPolicy(ctx)
	if err != nil {
		return nil, err
	}
	meshServiceKey, err := selector.BuildIdForMeshService(ctx, t.meshClient, meshService)
	if err != nil {
		return nil, err
	}
	for _, acp := range acpList.Items {
		acp := acp
		meshServicesForACP, err := t.resourceSelector.GetMeshServicesByServiceSelector(ctx, acp.Spec.GetDestinationSelector())
		if err != nil {
			return nil, err
		}
		for _, meshServiceForACP := range meshServicesForACP {
			meshServiceForACPKey, err := selector.BuildIdForMeshService(ctx, t.meshClient, meshServiceForACP)
			if err != nil {
				return nil, err
			}
			// MeshService equality is defined as equality on name, namespace, clusterName
			if meshServiceKey.Equals(meshServiceForACPKey) {
				applicableACPs = append(applicableACPs, &acp)
			}
		}
	}
	return applicableACPs, nil
}

// err represents errors during processing prior to translation
// translatorErrors represent errors during translation to mesh-specific config
func (t *translatorLoop) setStatus(
	err error,
	translatorErrors []*zephyr_networking_types.AccessControlPolicyStatus_TranslatorError,
	acp *zephyr_networking.AccessControlPolicy) {
	if err != nil {
		acp.Status.TranslationStatus = &zephyr_core_types.Status{
			State:   zephyr_core_types.Status_PROCESSING_ERROR,
			Message: fmt.Sprintf("Error while processing TrafficPolicy: %s", err.Error()),
		}
		// Clear any prior TranslatorErrors
		acp.Status.TranslatorErrors = nil
	} else if translatorErrors != nil {
		acp.Status.TranslationStatus = &zephyr_core_types.Status{
			State:   zephyr_core_types.Status_PROCESSING_ERROR,
			Message: fmt.Sprintf("Error while translating TrafficPolicy, check Status.TranslatorErrors for details"),
		}
		acp.Status.TranslatorErrors = translatorErrors
	} else {
		acp.Status.TranslationStatus = &zephyr_core_types.Status{
			State: zephyr_core_types.Status_ACCEPTED,
		}
		// Clear any prior TranslatorErrors
		acp.Status.TranslatorErrors = nil
	}
}

// Represents an AccessControlPolicy target, consisting of a MeshService and its associated Mesh
type TargetService struct {
	MeshService *zephyr_discovery.MeshService
	Mesh        *zephyr_discovery.Mesh
}

// Struct that pairs TranslatorErrors with the AccessControlPolicy they apply to
type translatorErrorForACP struct {
	accessControlPolicy *zephyr_networking.AccessControlPolicy
	translatorErrors    []*zephyr_networking_types.AccessControlPolicyStatus_TranslatorError
}
