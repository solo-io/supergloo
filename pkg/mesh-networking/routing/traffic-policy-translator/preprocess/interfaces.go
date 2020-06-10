package preprocess

import (
	"context"

	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/selection"
)

//go:generate mockgen -source ./interfaces.go -destination mocks/mock_interfaces.go

type TrafficPolicyPreprocessor interface {
	PreprocessTrafficPolicy(
		ctx context.Context,
		trafficPolicy *smh_networking.TrafficPolicy,
	) (map[selection.MeshServiceId][]*smh_networking.TrafficPolicy, error)

	PreprocessTrafficPoliciesForMeshService(
		ctx context.Context,
		meshService *smh_discovery.MeshService,
	) (map[selection.MeshServiceId][]*smh_networking.TrafficPolicy, error)
}

type TrafficPolicyMerger interface {
	MergeTrafficPoliciesForMeshServices(
		ctx context.Context,
		meshServices []*smh_discovery.MeshService,
	) (map[selection.MeshServiceId][]*smh_networking.TrafficPolicy, error)
}

type TrafficPolicyValidator interface {
	Validate(ctx context.Context, trafficPolicy *smh_networking.TrafficPolicy) error
}
