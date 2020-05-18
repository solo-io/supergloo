package traffic_policy_aggregation

import (
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
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
	) (
		policiesToRecordOnService []*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy,
		policyToConflictErrors map[*zephyr_networking.TrafficPolicy][]*zephyr_networking_types.TrafficPolicyStatus_ConflictError,
		policyToTranslatorErrors map[*zephyr_networking.TrafficPolicy][]*zephyr_networking_types.TrafficPolicyStatus_TranslatorError,
		err error,
	)
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

	// return a list of pairs:
	//    - the service (note that it will have the previously-recorded traffic policy state in its status)
	//           (services that have no traffic policies applying to them *will* be reflected in this list- their ServiceWithRelevantPolicies struct will have an empty `TrafficPolicies` field)
	//    - the traffic policies in the given snapshot that are associated with the above service.
	//           (This list must be reconciled with the existing state in the service's status)
	GroupByMeshService(
		trafficPolicies []*zephyr_networking.TrafficPolicy,
		meshServices []*zephyr_discovery.MeshService,
	) (result []*ServiceWithRelevantPolicies, err error)
}

type ServiceWithRelevantPolicies struct {
	MeshService     *zephyr_discovery.MeshService
	TrafficPolicies []*zephyr_networking.TrafficPolicy
}

type MeshServiceInfo struct {
	ClusterName string
	Mesh        *zephyr_discovery.Mesh
	MeshType    zephyr_core_types.MeshType
}
