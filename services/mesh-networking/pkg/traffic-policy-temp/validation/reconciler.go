package traffic_policy_validation

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
)

type ValidationProcessor interface {
	Process(ctx context.Context, allTrafficPolicies []*zephyr_networking.TrafficPolicy, meshServices []*zephyr_discovery.MeshService) []*zephyr_networking.TrafficPolicy
}
type trafficPolicyReaderStatusUpdater interface {
	zephyr_networking.TrafficPolicyReader
	zephyr_networking.TrafficPolicyStatusWriter
}

func NewValidationProcessor(
	trafficPolicyClient trafficPolicyReaderStatusUpdater,
	meshServiceReader zephyr_discovery.MeshServiceReader,
	validator Validator,
) ValidationProcessor {
	return &validationLoop{
		trafficPolicyClient: trafficPolicyClient,
		validator:           validator,
		meshServiceReader:   meshServiceReader,
	}
}

type validationLoop struct {
	trafficPolicyClient trafficPolicyReaderStatusUpdater
	meshServiceReader   zephyr_discovery.MeshServiceReader
	validator           Validator
}

func (*validationLoop) GetName() string {
	return "traffic-policy-validation"
}

func (v *validationLoop) Process(ctx context.Context, allTrafficPolicies []*zephyr_networking.TrafficPolicy, meshServices []*zephyr_discovery.MeshService) []*zephyr_networking.TrafficPolicy {
	logger := contextutils.LoggerFrom(ctx)
	var updatedPolicies []*zephyr_networking.TrafficPolicy

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
