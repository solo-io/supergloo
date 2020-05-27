package appmesh

import (
	"context"

	"github.com/aws/aws-sdk-go/service/appmesh"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
)

//go:generate mockgen -source ./interfaces.go -destination mocks/mock_interfaces.go

type AppmeshAccessControlDao interface {
	// Return two maps which associate workloads to services and vice versa.
	GetServicesAndWorkloadsForMesh(
		ctx context.Context,
		mesh *zephyr_discovery.Mesh,
	) (map[*zephyr_discovery.MeshService][]*zephyr_discovery.MeshWorkload,
		map[*zephyr_discovery.MeshWorkload][]*zephyr_discovery.MeshService,
		error)

	EnsureVirtualService(virtualServiceData *appmesh.VirtualServiceData) error

	EnsureVirtualRouter(virtualRouter *appmesh.VirtualRouterData) error

	EnsureRoute(route *appmesh.RouteData) error

	EnsureVirtualNode(virtualNode *appmesh.VirtualNodeData) error
}
