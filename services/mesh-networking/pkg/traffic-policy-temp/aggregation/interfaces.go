package traffic_policy_aggregation

import (
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	zephyr_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
)

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_interfaces.go

type Aggregator interface {
	// Check whether any of this incoming policy's configuration directly conflicts with the policies in the given list.
	// This is agnostic of source/destination; instead, we just a look at the actual routing configuration in the list.
	FindMergeConflict(
		trafficPolicyToMerge *zephyr_networking_types.TrafficPolicySpec,
		policiesToMergeWith []*zephyr_networking_types.TrafficPolicySpec,
		meshServices []*zephyr_discovery.MeshService,
	) *zephyr_networking_types.TrafficPolicyStatus_ConflictError

	// return a list of pairs:
	//    - the service (note that it will have the previously-recorded traffic policy state in its status)
	//           (services that have no traffic policies applying to them *will* be reflected in this list- their ServiceWithRelevantPolicies struct will have an empty `TrafficPolicies` field)
	//    - the traffic policies in the given snapshot that are associated with the above service.
	//           (This list must be reconciled with the existing state in the service's status)
	GroupByMeshService(
		trafficPolicies []*zephyr_networking.TrafficPolicy,
		meshServiceToClusterName map[*zephyr_discovery.MeshService]string,
	) []*ServiceWithRelevantPolicies
}

type ServiceWithRelevantPolicies struct {
	MeshService     *zephyr_discovery.MeshService
	TrafficPolicies []*zephyr_networking.TrafficPolicy
}
