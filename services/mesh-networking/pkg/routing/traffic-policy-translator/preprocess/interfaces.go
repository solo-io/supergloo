package preprocess

import (
	"context"

	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	discovery_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	networking_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/mesh-projects/services/mesh-networking/pkg/routing/traffic-policy-translator/keys"
)

//go:generate mockgen -source ./interfaces.go -destination mocks/mock_interfaces.go

type MeshServiceSelector interface {
	// fetch all MeshServices that match the given selector
	GetMatchingMeshServices(
		ctx context.Context,
		selector *core_types.Selector,
	) ([]*discovery_v1alpha1.MeshService, error)

	// fetch the MeshService backing a k8s Service by the Service's name, namespace, cluster name
	// return error if no MeshService found, or multiple
	GetBackingMeshService(
		ctx context.Context,
		kubeServiceName string,
		kubeServiceNamespace string,
		kubeServiceCluster string,
	) (*discovery_v1alpha1.MeshService, error)
}

type TrafficPolicyPreprocessor interface {
	PreprocessTrafficPolicy(
		ctx context.Context,
		trafficPolicy *networking_v1alpha1.TrafficPolicy,
	) (map[keys.MeshServiceMultiClusterKey][]*networking_v1alpha1.TrafficPolicy, error)

	PreprocessTrafficPoliciesForMeshService(
		ctx context.Context,
		meshService *discovery_v1alpha1.MeshService,
	) (map[keys.MeshServiceMultiClusterKey][]*networking_v1alpha1.TrafficPolicy, error)
}

type TrafficPolicyMerger interface {
	MergeTrafficPoliciesForMeshServices(
		ctx context.Context,
		meshServices []*discovery_v1alpha1.MeshService,
	) (map[keys.MeshServiceMultiClusterKey][]*networking_v1alpha1.TrafficPolicy, error)
}

type TrafficPolicyValidator interface {
	Validate(ctx context.Context, trafficPolicy *networking_v1alpha1.TrafficPolicy) error
}
