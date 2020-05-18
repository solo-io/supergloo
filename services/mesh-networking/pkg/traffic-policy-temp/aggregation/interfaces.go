package traffic_policy_aggregation

import (
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	zephyr_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	mesh_translation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/translation/meshes"
)

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_interfaces.go

// Build a collection of policies that apply to a traffic target (either a source or destination), while maintaining
// the last-known good state of those policies as recorded on the target in the case of a policy becoming invalid for some reason.
// This complexity is adopted in order to be defensive about not destructively mutating mesh config, which can disrupt traffic.
type PolicyCollector interface {
	// The `policiesToRecordOnService` slice will consist of a mixture of last-known good state plus newly-updated policies that pass validation.
	// If no update to the service status is required, that slice will be returned in the same order as it was recorded on the service.
	// If a change to the service status is required, that slice will be ordered by unchanged policies first, followed by any policies that were both updated and validated.
	// No finer-grained ordering is guaranteed beyond that. No guarantee is made about whether the references in `policiesToRecordOnService` are the
	// same (by reference equality) as the references that came in through the service.
	// Errors can occur due to invalid selectors on policies.
	CollectForService(
		meshService *zephyr_discovery.MeshService,
		mesh *zephyr_discovery.Mesh,
		translationValidator mesh_translation.TranslationValidator,
		allTrafficPolicies []*zephyr_networking.TrafficPolicy,
	) (*CollectionResult, error)
}

type CollectionResult struct {
	// the policies (a mix of updated and last-known good state) that should be recorded next on the service
	PoliciesToRecordOnService []*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy

	PolicyToConflictErrors   map[*zephyr_networking.TrafficPolicy][]*zephyr_networking_types.TrafficPolicyStatus_ConflictError
	PolicyToTranslatorErrors map[*zephyr_networking.TrafficPolicy][]*zephyr_networking_types.TrafficPolicyStatus_TranslatorError
}

// This interface serves to abstract away the details of merging new entries into various objects' statuses. Its behavior needs to
// ensure that we don't incur k8s api server calls when they are not necessary.
// The logic depends on there being an idempotent order coming out of `PolicyCollector` if no updates have happened.
// These methods can change the objects they are handed- they return true if the object was changed in memory and should be updated in the persistence layer
type InMemoryStatusUpdater interface {
	UpdateServicePolicies(
		meshService *zephyr_discovery.MeshService,
		newlyComputedMergeablePolicies []*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy,
	) (policyNeedsUpdating bool)
	UpdateConflictAndTranslatorErrors(
		policy *zephyr_networking.TrafficPolicy,
		newConflictErrors []*zephyr_networking_types.TrafficPolicyStatus_ConflictError,
		newTranslationErrors []*zephyr_networking_types.TrafficPolicyStatus_TranslatorError,
	) (policyNeedsUpdating bool)
}

type Aggregator interface {
	// Check whether any of this incoming policy's configuration directly conflicts with the policies in the given list.
	// This is agnostic of source/destination; instead, we just a look at the actual routing configuration in the list.
	FindMergeConflict(
		trafficPolicyToMerge *zephyr_networking_types.TrafficPolicySpec,
		policiesToMergeWith []*zephyr_networking_types.TrafficPolicySpec,
		meshService *zephyr_discovery.MeshService,
	) *zephyr_networking_types.TrafficPolicyStatus_ConflictError

	// return the policies tht have the given mesh service as a destination
	PoliciesForService(
		trafficPolicies []*zephyr_networking.TrafficPolicy,
		meshService *zephyr_discovery.MeshService,
	) ([]*zephyr_networking.TrafficPolicy, error)
}
