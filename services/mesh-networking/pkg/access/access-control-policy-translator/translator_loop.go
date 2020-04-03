package acp_translator

import (
	"context"
	"fmt"

	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	discovery_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	discovery_controller "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1/controller"
	"github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	networking_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	networking_controller "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1/controller"
	"github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/mesh-projects/pkg/clients"
	zephyr_discovery "github.com/solo-io/mesh-projects/pkg/clients/zephyr/discovery"
	zephyr_networking "github.com/solo-io/mesh-projects/pkg/clients/zephyr/networking"
	"github.com/solo-io/mesh-projects/pkg/logging"
	"github.com/solo-io/mesh-projects/services/mesh-networking/pkg/multicluster/selector"
)

func NewAcpTranslatorLoop(
	acpController networking_controller.AccessControlPolicyController,
	meshServiceController discovery_controller.MeshServiceController,
	meshClient zephyr_discovery.MeshClient,
	accessControlPolicyClient zephyr_networking.AccessControlPolicyClient,
	meshServiceSelector selector.MeshServiceSelector,
	meshTranslators []AcpMeshTranslator,
) AcpTranslatorLoop {
	return &translatorLoop{
		acpController:             acpController,
		meshServiceController:     meshServiceController,
		meshClient:                meshClient,
		accessControlPolicyClient: accessControlPolicyClient,
		meshServiceSelector:       meshServiceSelector,
		meshTranslators:           meshTranslators,
	}
}

type translatorLoop struct {
	acpController             networking_controller.AccessControlPolicyController
	meshServiceController     discovery_controller.MeshServiceController
	meshClient                zephyr_discovery.MeshClient
	accessControlPolicyClient zephyr_networking.AccessControlPolicyClient
	meshServiceSelector       selector.MeshServiceSelector
	meshTranslators           []AcpMeshTranslator
}

func (t *translatorLoop) Start(ctx context.Context) error {
	err := t.acpController.AddEventHandler(ctx, &networking_controller.AccessControlPolicyEventHandlerFuncs{
		OnCreate: func(acp *v1alpha1.AccessControlPolicy) error {
			logger := logging.BuildEventLogger(ctx, logging.CreateEvent, acp)
			logger.Debugf("Handling event: %+v", acp)
			translatorErrors, err := t.translateAccessControlPolicy(ctx, acp)
			t.setStatus(err, translatorErrors, acp)
			err = t.accessControlPolicyClient.UpdateStatus(ctx, acp)
			if err != nil {
				logger.Errorw("Error while handling AccessControlPolicy create event", err)
			}
			return nil
		},
		OnUpdate: func(_, acp *v1alpha1.AccessControlPolicy) error {
			logger := logging.BuildEventLogger(ctx, logging.UpdateEvent, acp)
			logger.Debugf("Handling event: %+v", acp)
			translatorErrors, err := t.translateAccessControlPolicy(ctx, acp)
			t.setStatus(err, translatorErrors, acp)
			err = t.accessControlPolicyClient.UpdateStatus(ctx, acp)
			if err != nil {
				logger.Errorw("Error while handling AccessControlPolicy update event", err)
			}
			return nil
		},
		OnDelete: func(policy *v1alpha1.AccessControlPolicy) error {
			logger := logging.BuildEventLogger(ctx, logging.DeleteEvent, policy)
			logger.Debugf("Ignoring event: %+v", policy)
			return nil
		},
		OnGeneric: func(policy *v1alpha1.AccessControlPolicy) error {
			logger := logging.BuildEventLogger(ctx, logging.GenericEvent, policy)
			logger.Debugf("Ignoring event: %+v", policy)
			return nil
		},
	})
	if err != nil {
		return err
	}
	return t.meshServiceController.AddEventHandler(ctx, &discovery_controller.MeshServiceEventHandlerFuncs{
		OnCreate: func(meshService *discovery_v1alpha1.MeshService) error {
			logger := logging.BuildEventLogger(ctx, logging.CreateEvent, meshService)
			logger.Debugf("Handling event: %+v", meshService)
			translatorErrorsForACPs, err := t.translateACPsForMeshService(ctx, meshService)
			// Update status for each ACP that was processed for MeshService
			for _, translatorErrWithACP := range translatorErrorsForACPs {
				t.setStatus(err, translatorErrWithACP.translatorErrors, translatorErrWithACP.accessControlPolicy)
				err = t.accessControlPolicyClient.UpdateStatus(ctx, translatorErrWithACP.accessControlPolicy)
				if err != nil {
					logger.Errorw("Error while handling MeshService create event", err)
				}
			}
			return nil
		},
		OnUpdate: func(_, meshService *discovery_v1alpha1.MeshService) error {
			logger := logging.BuildEventLogger(ctx, logging.UpdateEvent, meshService)
			logger.Debugf("Handling event: %+v", meshService)
			translatorErrorsForACPs, err := t.translateACPsForMeshService(ctx, meshService)
			// Update status for each ACP that was processed for MeshService
			for _, translatorErrWithACP := range translatorErrorsForACPs {
				t.setStatus(err, translatorErrWithACP.translatorErrors, translatorErrWithACP.accessControlPolicy)
				err = t.accessControlPolicyClient.UpdateStatus(ctx, translatorErrWithACP.accessControlPolicy)
				if err != nil {
					logger.Errorw("Error while handling MeshService create event", err)
				}
			}
			return nil
		},
		OnDelete: func(meshService *discovery_v1alpha1.MeshService) error {
			logger := logging.BuildEventLogger(ctx, logging.DeleteEvent, meshService)
			logger.Debugf("Ignoring event: %+v", meshService)
			return nil
		},
		OnGeneric: func(meshService *discovery_v1alpha1.MeshService) error {
			logger := logging.BuildEventLogger(ctx, logging.GenericEvent, meshService)
			logger.Debugf("Ignoring event: %+v", meshService)
			return nil
		},
	})
}

// Translate AccessControlPolicy to AuthorizationPolicy for all targeted k8s Services
func (t *translatorLoop) translateAccessControlPolicy(
	ctx context.Context,
	acp *v1alpha1.AccessControlPolicy,
) ([]*types.AccessControlPolicyStatus_TranslatorError, error) {
	targetServices, err := t.getTargetServices(ctx, acp)
	if err != nil {
		return nil, err
	}
	var translatorErrors []*types.AccessControlPolicyStatus_TranslatorError
	for _, meshTranslator := range t.meshTranslators {
		translatorError := meshTranslator.Translate(ctx, targetServices, acp)
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
	meshService *discovery_v1alpha1.MeshService,
) ([]translatorErrorForACP, error) {
	mesh, err := t.meshClient.Get(ctx, clients.ResourceRefToObjectKey(meshService.Spec.GetMesh()))
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
		var translatorErrors []*types.AccessControlPolicyStatus_TranslatorError
		for _, meshTranslator := range t.meshTranslators {
			translatorError := meshTranslator.Translate(ctx, []TargetService{targetService}, acp)
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
func (t *translatorLoop) getTargetServices(ctx context.Context, acp *networking_v1alpha1.AccessControlPolicy) ([]TargetService, error) {
	meshServices, err := t.meshServiceSelector.GetMatchingMeshServices(ctx, acp.Spec.GetDestinationSelector())
	if err != nil {
		return nil, err
	}
	var targetServices []TargetService
	for _, meshService := range meshServices {
		mesh, err := t.meshClient.Get(ctx, clients.ResourceRefToObjectKey(meshService.Spec.GetMesh()))
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
	meshService *discovery_v1alpha1.MeshService,
) ([]*networking_v1alpha1.AccessControlPolicy, error) {
	var applicableACPs []*networking_v1alpha1.AccessControlPolicy
	acpList, err := t.accessControlPolicyClient.List(ctx)
	if err != nil {
		return nil, err
	}
	meshServiceKey, err := selector.BuildIdForMeshService(ctx, t.meshClient, meshService)
	if err != nil {
		return nil, err
	}
	for _, acp := range acpList.Items {
		acp := acp
		meshServicesForACP, err := t.meshServiceSelector.GetMatchingMeshServices(ctx, acp.Spec.GetDestinationSelector())
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
	translatorErrors []*types.AccessControlPolicyStatus_TranslatorError,
	acp *networking_v1alpha1.AccessControlPolicy) {
	if err != nil {
		acp.Status.TranslationStatus = &core_types.Status{
			State:   core_types.Status_PROCESSING_ERROR,
			Message: fmt.Sprintf("Error while processing TrafficPolicy: %s", err.Error()),
		}
		// Clear any prior TranslatorErrors
		acp.Status.TranslatorErrors = nil
	} else if translatorErrors != nil {
		acp.Status.TranslationStatus = &core_types.Status{
			State:   core_types.Status_PROCESSING_ERROR,
			Message: fmt.Sprintf("Error while translating TrafficPolicy, check Status.TranslatorErrors for details"),
		}
		acp.Status.TranslatorErrors = translatorErrors
	} else {
		acp.Status.TranslationStatus = &core_types.Status{
			State: core_types.Status_ACCEPTED,
		}
		// Clear any prior TranslatorErrors
		acp.Status.TranslatorErrors = nil
	}
}

// Represents an AccessControlPolicy target, consisting of a MeshService and its associated Mesh
type TargetService struct {
	MeshService *discovery_v1alpha1.MeshService
	Mesh        *discovery_v1alpha1.Mesh
}

// Struct that pairs TranslatorErrors with the AccessControlPolicy they apply to
type translatorErrorForACP struct {
	accessControlPolicy *networking_v1alpha1.AccessControlPolicy
	translatorErrors    []*types.AccessControlPolicyStatus_TranslatorError
}
