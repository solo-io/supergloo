package selector

import (
	"context"

	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	discovery_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
)

//go:generate mockgen -source ./interfaces.go -destination mocks/mock_interfaces.go

// Methods for fetching applicable MeshServices for Selector
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
