package failover

import (
	"context"

	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/types"
)

type FailoverServiceProcessor interface {
	Process(ctx context.Context, inputSnapshot InputSnapshot) OutputSnapshot
}

type failoverServiceProcessor struct {
	validator   FailoverServiceValidator
	translators []FailoverServiceTranslator
}

/*
	Processing consists of the following sequence of steps:
	1. Validate the FailoverServices and update the validation status.
	2. For the valid FailoverServices, translate them to mesh-specific configuration, update translation status.
	Return an OutputSnapshot containing FailoverServices with updated statuses and translated resources.
*/
func (f *failoverServiceProcessor) Process(ctx context.Context, inputSnapshot InputSnapshot) OutputSnapshot {
	outputSnapshot := OutputSnapshot{}
	// Validate will set the validation status and observed generation on the FailoverService status.
	f.validator.Validate(inputSnapshot)
	for _, failoverService := range inputSnapshot.FailoverServices {
		if !f.readyToProcess(failoverService) {
			continue
		}
		prioritizedMeshServices, err := f.collectMeshServicesForFailoverService(failoverService, inputSnapshot.MeshServices)
		if err != nil {
			failoverService.Status = f.computeProcessingErrorStatus(err)
			continue
		}
		outputSnapshotForFailoverService := f.processSingle(ctx, failoverService, prioritizedMeshServices)
		// Accumulate outputs for each FailoverService
		outputSnapshot.append(outputSnapshotForFailoverService)
	}
	return outputSnapshot
}

// Return true if validation status is accepted and observed generation matches generation.
func (f *failoverServiceProcessor) readyToProcess(failoverService *smh_networking.FailoverService) bool {
	return failoverService.Status.GetValidationStatus().GetState() == smh_core_types.Status_ACCEPTED &&
		failoverService.Status.GetObservedGeneration() == failoverService.GetGeneration()
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

// Process a single FailoverService and return OutputSnapshot containing computed translated resources and
// the FailoverService with updated status.
func (f *failoverServiceProcessor) processSingle(
	ctx context.Context,
	failoverService *smh_networking.FailoverService,
	prioritizedMeshServices []*smh_discovery.MeshService,
) OutputSnapshot {
	var translatorErrs []*types.FailoverServiceStatus_TranslatorError
	var outputSnapshot OutputSnapshot
	for _, translator := range f.translators {
		output, translatorErr := translator.Translate(ctx, failoverService, prioritizedMeshServices)
		// Accumulate mesh specific output resources.
		outputSnapshot.MeshOutputs.append(output)
		if translatorErr != nil {
			translatorErrs = append(translatorErrs, translatorErr)
		}
	}
	// Set status on FailoverService and add to OutputSnapshot
	failoverService.Status = f.computeTranslatorErrorStatus(translatorErrs)
	outputSnapshot.FailoverServices = []*smh_networking.FailoverService{failoverService}
	return outputSnapshot
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
