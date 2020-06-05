package traffic_policy_validation

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/reconciliation"
)

type trafficPolicyReaderStatusUpdated interface {
	zephyr_networking.TrafficPolicyReader
	zephyr_networking.TrafficPolicyStatusWriter
}

func NewValidationReconciler(
	trafficPolicyClient trafficPolicyReaderStatusUpdated,
	meshServiceClient zephyr_discovery.MeshServiceReader,
	validator Validator,
) reconciliation.Reconciler {
	return &validationLoop{
		trafficPolicyClient: trafficPolicyClient,
		validator:           validator,
		meshServiceClient:   meshServiceClient,
	}
}

type validationLoop struct {
	trafficPolicyClient trafficPolicyReaderStatusUpdated
	meshServiceClient   zephyr_discovery.MeshServiceReader
	validator           Validator
}

func (*validationLoop) GetName() string {
	return "traffic-policy-validation"
}

func (v *validationLoop) Reconcile(ctx context.Context) error {
	logger := contextutils.LoggerFrom(ctx)
	trafficPolicies, err := v.trafficPolicyClient.ListTrafficPolicy(ctx)
	if err != nil {
		return err
	}

	meshServiceList, err := v.meshServiceClient.ListMeshService(ctx)
	if err != nil {
		return err
	}

	var meshServices []*zephyr_discovery.MeshService
	for _, ms := range meshServiceList.Items {
		meshService := ms
		meshServices = append(meshServices, &meshService)
	}

	for _, trafficPolicy := range trafficPolicies.Items {
		newValidationStatus, validationErr := v.validator.ValidateTrafficPolicy(&trafficPolicy, meshServices)
		if validationErr == nil {
			logger.Debugf("Traffic policy %s.%s passed validation", trafficPolicy.GetName(), trafficPolicy.GetNamespace())
		} else {
			logger.Infof("Traffic policy %s.%s failed validation for reason: %+v", trafficPolicy.GetName(), trafficPolicy.GetNamespace(), validationErr)
		}

		if !trafficPolicy.Status.GetValidationStatus().Equal(newValidationStatus) ||
			trafficPolicy.Status.ObservedGeneration != trafficPolicy.Generation {
			trafficPolicy.Status.ObservedGeneration = trafficPolicy.Generation
			trafficPolicy.Status.ValidationStatus = newValidationStatus

			err := v.trafficPolicyClient.UpdateTrafficPolicyStatus(ctx, &trafficPolicy)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
