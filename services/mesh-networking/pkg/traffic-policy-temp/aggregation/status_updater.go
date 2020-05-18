package traffic_policy_aggregation

import (
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	zephyr_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
)

func NewInMemoryStatusUpdater() InMemoryStatusUpdater {
	return &inMemoryStatusUpdater{}
}

type inMemoryStatusUpdater struct{}

func (*inMemoryStatusUpdater) UpdateServicePolicies(
	meshService *zephyr_discovery.MeshService,
	newlyComputedMergeablePolicies []*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy,
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

func (*inMemoryStatusUpdater) UpdateConflictAndTranslatorErrors(
	policy *zephyr_networking.TrafficPolicy,
	newConflictErrors []*zephyr_networking_types.TrafficPolicyStatus_ConflictError,
	newTranslationErrors []*zephyr_networking_types.TrafficPolicyStatus_TranslatorError,
) (policyNeedsUpdating bool) {
	if len(newConflictErrors) != len(policy.Status.GetConflictErrors()) {
		policyNeedsUpdating = true
	} else {
		for newErrorIndex, newError := range newConflictErrors {
			policyNeedsUpdating = policyNeedsUpdating || !policy.Status.GetConflictErrors()[newErrorIndex].Equal(newError)
		}
	}

	if len(newTranslationErrors) != len(policy.Status.TranslatorErrors) {
		policyNeedsUpdating = true
	} else {
		for newErrorIndex, newError := range newTranslationErrors {
			policyNeedsUpdating = policyNeedsUpdating || !policy.Status.GetTranslatorErrors()[newErrorIndex].Equal(newError)
		}
	}

	if policyNeedsUpdating {
		policy.Status.ConflictErrors = newConflictErrors
		policy.Status.TranslatorErrors = newTranslationErrors
	}

	return policyNeedsUpdating
}
