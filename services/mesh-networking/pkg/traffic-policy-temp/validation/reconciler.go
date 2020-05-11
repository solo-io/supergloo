package traffic_policy_validation

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/reconciliation"
)

func NewValidationReconciler(
	trafficPolicyClient zephyr_networking.TrafficPolicyClient,
	meshServiceClient zephyr_discovery.MeshServiceClient,
	validator Validator,
) reconciliation.Reconciler {
	return &validationLoop{
		trafficPolicyClient: trafficPolicyClient,
		validator:           validator,
		meshServiceClient:   meshServiceClient,
	}
}

type validationLoop struct {
	trafficPolicyClient zephyr_networking.TrafficPolicyClient
	meshServiceClient   zephyr_discovery.MeshServiceClient
	validator           Validator
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

		if !trafficPolicy.Status.GetValidationStatus().Equal(newValidationStatus) {
			trafficPolicy.Status.ValidationStatus = newValidationStatus

			// also zero-out the conflict errors, since the state has changed and we don't know what it may conflict with now
			trafficPolicy.Status.ConflictErrors = nil

			err := v.trafficPolicyClient.UpdateTrafficPolicyStatus(ctx, &trafficPolicy)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
