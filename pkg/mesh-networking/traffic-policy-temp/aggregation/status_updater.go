package traffic_policy_aggregation

import (
	"github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	smh_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/types"
)

func NewInMemoryStatusMutator() InMemoryStatusMutator {
	return &inMemoryStatusMutator{}
}

type inMemoryStatusMutator struct{}

func (*inMemoryStatusMutator) MutateServicePolicies(
	meshService *smh_discovery.MeshService,
	newlyComputedMergeablePolicies []*smh_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy,
) (serviceNeedsUpdating bool) {
	if len(newlyComputedMergeablePolicies) != len(meshService.Status.ValidatedTrafficPolicies) {
		serviceNeedsUpdating = true
	} else {
		for i, newlyComputedPolicy := range newlyComputedMergeablePolicies {
			serviceNeedsUpdating = serviceNeedsUpdating || !meshService.Status.ValidatedTrafficPolicies[i].Equal(newlyComputedPolicy)
		}
	}

	if serviceNeedsUpdating {
		meshService.Status.ValidatedTrafficPolicies = newlyComputedMergeablePolicies
	}

	return serviceNeedsUpdating
}

func (*inMemoryStatusMutator) MutateTrafficPolicyTranslationStatus(
	policy *smh_networking.TrafficPolicy,
	newConflictErrors []*smh_networking_types.TrafficPolicyStatus_ConflictError,
	newTranslationErrors []*smh_networking_types.TrafficPolicyStatus_TranslatorError,
) {
	var translationStatus *types.Status
	if len(newConflictErrors) == 0 && len(newTranslationErrors) == 0 {
		translationStatus = &types.Status{
			State: types.Status_ACCEPTED,
		}
	} else {
		translationStatus = &types.Status{
			State: types.Status_INVALID,
		}
	}
	policy.Status.TranslationStatus = translationStatus
	policy.Status.ConflictErrors = newConflictErrors
	policy.Status.TranslatorErrors = newTranslationErrors
}
