package traffic_policy_validation

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
)

type trafficPolicyReaderStatusUpdater interface {
	smh_networking.TrafficPolicyReader
	smh_networking.TrafficPolicyStatusWriter
}

func NewValidationProcessor(
	validator Validator,
) ValidationProcessor {
	return &validationProcessor{
		validator: validator,
	}
}

type validationProcessor struct {
	validator Validator
}

func (*validationProcessor) GetName() string {
	return "traffic-policy-validation"
}

func (v *validationProcessor) Process(ctx context.Context, allTrafficPolicies []*smh_networking.TrafficPolicy, meshServices []*smh_discovery.MeshService) []*smh_networking.TrafficPolicy {
	logger := contextutils.LoggerFrom(ctx)
	var updatedPolicies []*smh_networking.TrafficPolicy

	for _, trafficPolicy := range allTrafficPolicies {
		newValidationStatus, validationErr := v.validator.ValidateTrafficPolicy(trafficPolicy, meshServices)
		if validationErr == nil {
			logger.Debugf("Traffic policy %s.%s passed validation", trafficPolicy.GetName(), trafficPolicy.GetNamespace())
		} else {
			logger.Infof("Traffic policy %s.%s failed validation for reason: %+v", trafficPolicy.GetName(), trafficPolicy.GetNamespace(), validationErr)
		}
		if !trafficPolicy.Status.GetValidationStatus().Equal(newValidationStatus) ||
			trafficPolicy.Status.ObservedGeneration != trafficPolicy.Generation {
			trafficPolicy.Status.ObservedGeneration = trafficPolicy.Generation
			trafficPolicy.Status.ValidationStatus = newValidationStatus
			updatedPolicies = append(updatedPolicies, trafficPolicy)
		}
	}

	return updatedPolicies
}
