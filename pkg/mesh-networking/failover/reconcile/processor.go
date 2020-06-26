package reconcile

import (
	"context"

	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/failover"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/failover/translation"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/failover/validation"
)

//go:generate mockgen -source ./processor.go -destination ./mocks/mock_processor.go

type FailoverServiceProcessor interface {
	Process(ctx context.Context, inputSnapshot failover.InputSnapshot) failover.OutputSnapshot
}

type failoverServiceProcessor struct {
	validator   validation.FailoverServiceValidator
	translators []translation.FailoverServiceTranslator
}

func NewFailoverServiceProcessor(
	validator validation.FailoverServiceValidator,
	translators []translation.FailoverServiceTranslator,
) FailoverServiceProcessor {
	return &failoverServiceProcessor{
		validator:   validator,
		translators: translators,
	}
}

/*
	Processing consists of the following sequence of steps:
	1. Validate the FailoverServices and update the validation status.
	2. For the valid FailoverServices, translate them to mesh-specific configuration, update translation status.
	Return an OutputSnapshot containing FailoverServices with updated statuses and translated resources.

	TODO(harveyxia) The FailoverService processor must also be invoked for the following CRD events:
	1. TrafficPolicy
	2. MeshService
	3. VirtualMesh
	4. Mesh
*/
func (f *failoverServiceProcessor) Process(ctx context.Context, inputSnapshot failover.InputSnapshot) failover.OutputSnapshot {
	outputSnapshot := failover.OutputSnapshot{}
	// Validate will set the validation status and observed generation on the FailoverService status.
	f.validator.Validate(inputSnapshot)
	for _, failoverService := range inputSnapshot.FailoverServices {
		if f.readyToProcess(failoverService) {
			prioritizedMeshServices, err := f.collectMeshServicesForFailoverService(failoverService, inputSnapshot.MeshServices)
			if err != nil {
				failoverService.Status = f.computeProcessingErrorStatus(err)
				continue
			}
			failoverServiceWithUpdatedStatus, meshOutputs := f.processSingle(ctx, failoverService, prioritizedMeshServices)
			// Accumulate outputs for each FailoverService
			outputSnapshot.MeshOutputs = outputSnapshot.MeshOutputs.Append(meshOutputs)
			failoverService = failoverServiceWithUpdatedStatus
		}
		outputSnapshot.FailoverServices = append(outputSnapshot.FailoverServices, failoverService)
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
			return nil, validation.ServiceNotFound(serviceRef)
		}
		prioritizedMeshServices = append(prioritizedMeshServices, matchingMeshService)
	}
	return prioritizedMeshServices, nil
}

// Process a single FailoverService and return MeshOutput containing computed translated resources and
// the FailoverService with updated status.
func (f *failoverServiceProcessor) processSingle(
	ctx context.Context,
	failoverService *smh_networking.FailoverService,
	prioritizedMeshServices []*smh_discovery.MeshService,
) (*smh_networking.FailoverService, failover.MeshOutputs) {
	var translatorErrs []*types.FailoverServiceStatus_TranslatorError
	var outputs failover.MeshOutputs
	for _, translator := range f.translators {
		output, translatorErr := translator.Translate(ctx, failoverService, prioritizedMeshServices)
		if translatorErr != nil {
			translatorErrs = append(translatorErrs, translatorErr)
			continue
		}
		// Accumulate mesh specific output resources.
		outputs = outputs.Append(output)
	}
	// Set status on FailoverService and add to OutputSnapshot
	failoverService.Status.TranslationStatus = f.computeTranslationStatus(translatorErrs)
	failoverService.Status.TranslatorErrors = translatorErrs
	return failoverService, outputs
}

func (f *failoverServiceProcessor) computeProcessingErrorStatus(err error) types.FailoverServiceStatus {
	return types.FailoverServiceStatus{
		ValidationStatus: &smh_core_types.Status{
			State:   smh_core_types.Status_INVALID,
			Message: err.Error(),
		},
	}
}

func (f *failoverServiceProcessor) computeTranslationStatus(
	translatorErrs []*types.FailoverServiceStatus_TranslatorError,
) *smh_core_types.Status {
	var status *smh_core_types.Status
	if len(translatorErrs) == 0 {
		status = &smh_core_types.Status{
			State: smh_core_types.Status_ACCEPTED,
		}
	} else {
		status = &smh_core_types.Status{
			State: smh_core_types.Status_PROCESSING_ERROR,
		}
	}
	return status
}
