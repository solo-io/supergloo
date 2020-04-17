package preprocess

import (
	"context"

	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/selector"
)

//go:generate mockgen -source ./interfaces.go -destination mocks/mock_interfaces.go

type TrafficPolicyPreprocessor interface {
	PreprocessTrafficPolicy(
		ctx context.Context,
		trafficPolicy *zephyr_networking.TrafficPolicy,
	) (map[selector.MeshServiceId][]*zephyr_networking.TrafficPolicy, error)

	PreprocessTrafficPoliciesForMeshService(
		ctx context.Context,
		meshService *zephyr_discovery.MeshService,
	) (map[selector.MeshServiceId][]*zephyr_networking.TrafficPolicy, error)
}

type TrafficPolicyMerger interface {
	MergeTrafficPoliciesForMeshServices(
		ctx context.Context,
		meshServices []*zephyr_discovery.MeshService,
	) (map[selector.MeshServiceId][]*zephyr_networking.TrafficPolicy, error)
}

type TrafficPolicyValidator interface {
	Validate(ctx context.Context, trafficPolicy *zephyr_networking.TrafficPolicy) error
}
