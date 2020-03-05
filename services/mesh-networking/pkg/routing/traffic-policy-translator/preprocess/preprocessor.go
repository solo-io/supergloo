package preprocess

import (
	"context"

	discovery_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	networking_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/mesh-projects/services/mesh-networking/pkg/routing/traffic-policy-translator/keys"
)

type trafficPolicyPreprocessor struct {
	meshServiceSelector MeshServiceSelector
	merger              TrafficPolicyMerger
}

func NewTrafficPolicyPreprocessor(
	meshServiceSelector MeshServiceSelector,
	merger TrafficPolicyMerger,
) TrafficPolicyPreprocessor {
	return &trafficPolicyPreprocessor{
		meshServiceSelector: meshServiceSelector,
		merger:              merger,
	}
}

/*
	Given a TrafficPolicy, do the following:
	1. Fetch all destination MeshServices
	2. For each MeshService, do the following:
		a. Fetch existing TrafficPolicies that apply to it
		b. Sort the TrafficPolicies by creation time ascending
		c. Merge the TrafficPolicies
			- If conflict encountered on any TrafficPolicy, return conflict error
*/
func (d *trafficPolicyPreprocessor) PreprocessTrafficPolicy(
	ctx context.Context,
	trafficPolicy *networking_v1alpha1.TrafficPolicy,
) (map[keys.MeshServiceMultiClusterKey][]*networking_v1alpha1.TrafficPolicy, error) {
	meshServices, err := d.meshServiceSelector.GetMatchingMeshServices(
		ctx,
		trafficPolicy.Spec.GetDestinationSelector(),
	)
	if err != nil {
		return nil, err
	}
	return d.merger.MergeTrafficPoliciesForMeshServices(ctx, meshServices)
}

/*
	Given a MeshService, do the following:
	1. Fetch existing TrafficPolicies that apply to it
	2. Sort the TrafficPolicies by creation time ascending
	3. Merge the TrafficPolicies
		- If conflict encountered on any TrafficPolicy, do not apply any of its rules and update its status to CONFLICT
*/
func (d *trafficPolicyPreprocessor) PreprocessTrafficPoliciesForMeshService(
	ctx context.Context,
	meshService *discovery_v1alpha1.MeshService,
) (map[keys.MeshServiceMultiClusterKey][]*networking_v1alpha1.TrafficPolicy, error) {
	return d.merger.MergeTrafficPoliciesForMeshServices(ctx, []*discovery_v1alpha1.MeshService{meshService})
}
