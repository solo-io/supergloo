package preprocess

import (
	"context"

	discovery_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	networking_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/mesh-projects/pkg/selector"
)

//go:generate mockgen -source ./interfaces.go -destination mocks/mock_interfaces.go

type TrafficPolicyPreprocessor interface {
	PreprocessTrafficPolicy(
		ctx context.Context,
		trafficPolicy *networking_v1alpha1.TrafficPolicy,
	) (map[selector.MeshServiceId][]*networking_v1alpha1.TrafficPolicy, error)

	PreprocessTrafficPoliciesForMeshService(
		ctx context.Context,
		meshService *discovery_v1alpha1.MeshService,
	) (map[selector.MeshServiceId][]*networking_v1alpha1.TrafficPolicy, error)
}

type TrafficPolicyMerger interface {
	MergeTrafficPoliciesForMeshServices(
		ctx context.Context,
		meshServices []*discovery_v1alpha1.MeshService,
	) (map[selector.MeshServiceId][]*networking_v1alpha1.TrafficPolicy, error)
}

type TrafficPolicyValidator interface {
	Validate(ctx context.Context, trafficPolicy *networking_v1alpha1.TrafficPolicy) error
}
