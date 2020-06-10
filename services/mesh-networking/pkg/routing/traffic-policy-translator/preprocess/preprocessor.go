package preprocess

import (
	"context"

	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/selection"
)

type trafficPolicyPreprocessor struct {
	resourceSelector selection.ResourceSelector
	merger           TrafficPolicyMerger
	validator        TrafficPolicyValidator
}

func NewTrafficPolicyPreprocessor(
	resourceSelector selection.ResourceSelector,
	merger TrafficPolicyMerger,
	validator TrafficPolicyValidator,
) TrafficPolicyPreprocessor {
	return &trafficPolicyPreprocessor{
		resourceSelector: resourceSelector,
		merger:           merger,
		validator:        validator,
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
	trafficPolicy *smh_networking.TrafficPolicy,
) (map[selection.MeshServiceId][]*smh_networking.TrafficPolicy, error) {
	err := d.validator.Validate(ctx, trafficPolicy)
	if err != nil {
		return nil, err
	}
	meshServices, err := d.resourceSelector.GetAllMeshServicesByServiceSelector(
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
	meshService *smh_discovery.MeshService,
) (map[selection.MeshServiceId][]*smh_networking.TrafficPolicy, error) {
	return d.merger.MergeTrafficPoliciesForMeshServices(ctx, []*smh_discovery.MeshService{meshService})
}
