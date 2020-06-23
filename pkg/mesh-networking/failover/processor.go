package failover

import (
	"context"

	"github.com/rotisserie/eris"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/types"
)

var (
	ServiceNotFound = func(serviceRef *smh_core_types.ResourceRef) error {
		return eris.Errorf("Declared service %s.%s.%s not found in SMH discovery resources.",
			serviceRef.GetName(),
			serviceRef.GetNamespace(),
			serviceRef.GetCluster())
	}
)

type FailoverServiceProcessor interface {
	Process(ctx context.Context, inputSnapshot InputSnapshot) OutputSnapshot
}

type failoverServiceProcessor struct {
	translators []FailoverServiceTranslator
}

func (f *failoverServiceProcessor) Process(ctx context.Context, inputSnapshot InputSnapshot) OutputSnapshot {
	outputSnapshot := OutputSnapshot{}
	for _, failoverService := range inputSnapshot.FailoverServices {
		prioritizedMeshServices, err := f.collectMeshServicesForFailoverService(failoverService, inputSnapshot.MeshServices)
		if err != nil {
			failoverService.Status = f.computeProcessingErrorStatus(err)
			continue
		}
		translatorErrs := f.processFailoverService(ctx, failoverService, prioritizedMeshServices)
		failoverService.Status = f.computeTranslatorErrorStatus(translatorErrs)
		outputSnapshot.FailoverServices = append(outputSnapshot.FailoverServices, failoverService)
	}
	return outputSnapshot
}

// Collect, in priority order as declared in the FailoverService, the relevant MeshServices.
// If a MeshService cannot be found, return an error
func (f *failoverServiceProcessor) collectMeshServicesForFailoverService(
	failoverService *smh_networking.FailoverService,
	allMeshServices []*smh_discovery.MeshService,
) ([]*smh_discovery.MeshService, error) {
	var prioritizedMeshServices []*smh_discovery.MeshService
	for _, serviceRef := range failoverService.Spec.GetServices() {
		var matchingMeshService *smh_discovery.MeshService
		for _, meshService := range allMeshServices {
			kubeServiceRef := meshService.Spec.GetKubeService().GetRef()
			if serviceRef.GetName() != kubeServiceRef.GetName() ||
				serviceRef.GetNamespace() != kubeServiceRef.GetNamespace() ||
				serviceRef.GetCluster() != kubeServiceRef.GetCluster() {
				continue
			}
			matchingMeshService = meshService
		}
		if matchingMeshService == nil {
			return nil, ServiceNotFound(serviceRef)
		}
		prioritizedMeshServices = append(prioritizedMeshServices, matchingMeshService)
	}
	return prioritizedMeshServices, nil
}

func (f *failoverServiceProcessor) processFailoverService(
	ctx context.Context,
	failoverService *smh_networking.FailoverService,
	prioritizedMeshServices []*smh_discovery.MeshService,
) []*types.FailoverServiceStatus_TranslatorError {
	var translatorErrs []*types.FailoverServiceStatus_TranslatorError
	for _, translator := range f.translators {
		translatorErr := translator.Translate(ctx, failoverService, prioritizedMeshServices)
		if translatorErr != nil {
			translatorErrs = append(translatorErrs, translatorErr)
		}
	}
	return translatorErrs
}

// TODO this shouldn't be possible, how to handle?
func (f *failoverServiceProcessor) computeProcessingErrorStatus(err error) types.FailoverServiceStatus {
	return types.FailoverServiceStatus{
		ValidationStatus: &smh_core_types.Status{
			State:   smh_core_types.Status_INVALID,
			Message: err.Error(),
		},
	}
}

func (f *failoverServiceProcessor) computeTranslatorErrorStatus(
	translatorErrs []*types.FailoverServiceStatus_TranslatorError,
) types.FailoverServiceStatus {
	var status types.FailoverServiceStatus
	if len(translatorErrs) == 0 {
		return types.FailoverServiceStatus{
			TranslationStatus: &smh_core_types.Status{
				State: smh_core_types.Status_ACCEPTED,
			},
		}
	}
	status = types.FailoverServiceStatus{
		TranslationStatus: &smh_core_types.Status{
			State: smh_core_types.Status_PROCESSING_ERROR,
		},
	}
	for _, translatorErr := range translatorErrs {
		status.TranslatorErrors = append(status.TranslatorErrors, translatorErr)
	}
	return status
}
