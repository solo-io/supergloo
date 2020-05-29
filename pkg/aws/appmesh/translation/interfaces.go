package translation

import (
	"context"

	"github.com/aws/aws-sdk-go/service/appmesh"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/collections/sets"
)

//go:generate mockgen -source ./interfaces.go -destination mocks/mock_interfaces.go

type AppmeshTranslationReconciler interface {
	Reconcile(ctx context.Context,
		mesh *zephyr_discovery.Mesh,
		virtualMesh *zephyr_networking.VirtualMesh,
	) error
}

type AppmeshTranslator interface {
	// For a given MeshWorkload, return a VirtualNode with the given upstream services declared as VirtualService backends.
	BuildVirtualNode(
		appmeshName *string,
		meshWorkload *zephyr_discovery.MeshWorkload,
		meshService *zephyr_discovery.MeshService,
		upstreamServices []*zephyr_discovery.MeshService,
	) *appmesh.VirtualNodeData

	// For a given MeshService, return a Route that targets all its backing workloads with equal weight.
	BuildDefaultRoute(
		appmeshName *string,
		meshService *zephyr_discovery.MeshService,
		meshWorkloads []*zephyr_discovery.MeshWorkload,
	) (*appmesh.RouteData, error)

	// For a given MeshService, return a VirtualService.
	BuildVirtualService(
		appmeshName *string,
		meshService *zephyr_discovery.MeshService,
	) *appmesh.VirtualServiceData

	// For a given MeshService, return a VirtualRouter.
	BuildVirtualRouter(
		appmeshName *string,
		meshService *zephyr_discovery.MeshService,
	) *appmesh.VirtualRouterData
}

type AppmeshAccessControlDao interface {
	// Return two maps which associate workloads to backing services and vice versa.
	GetAllServiceWorkloadPairsForMesh(
		ctx context.Context,
		mesh *zephyr_discovery.Mesh,
	) (map[*zephyr_discovery.MeshService][]*zephyr_discovery.MeshWorkload,
		map[*zephyr_discovery.MeshWorkload][]*zephyr_discovery.MeshService,
		error)

	GetWorkloadsToAllUpstreamServices(
		ctx context.Context,
		mesh *zephyr_discovery.Mesh,
	) (map[*zephyr_discovery.MeshWorkload]sets.MeshServiceSet, error)

	GetServicesWithACP(
		ctx context.Context,
		mesh *zephyr_discovery.Mesh,
	) (sets.MeshServiceSet, error)

	GetWorkloadsToUpstreamServicesWithACP(
		ctx context.Context,
		mesh *zephyr_discovery.Mesh,
	) (sets.MeshWorkloadSet, map[*zephyr_discovery.MeshWorkload]sets.MeshServiceSet, error)

	EnsureVirtualService(
		mesh *zephyr_discovery.Mesh,
		virtualServiceData *appmesh.VirtualServiceData,
	) error

	EnsureVirtualRouter(
		mesh *zephyr_discovery.Mesh,
		virtualRouter *appmesh.VirtualRouterData,
	) error

	EnsureRoute(
		mesh *zephyr_discovery.Mesh,
		route *appmesh.RouteData,
	) error

	EnsureVirtualNode(
		mesh *zephyr_discovery.Mesh,
		virtualNode *appmesh.VirtualNodeData,
	) error

	ReconcileVirtualServices(
		ctx context.Context,
		mesh *zephyr_discovery.Mesh,
		virtualServices []*appmesh.VirtualServiceData,
	) error

	ReconcileVirtualRouters(
		ctx context.Context,
		mesh *zephyr_discovery.Mesh,
		virtualRouters []*appmesh.VirtualRouterData,
	) error

	ReconcileRoutes(
		ctx context.Context,
		mesh *zephyr_discovery.Mesh,
		routes []*appmesh.RouteData,
	) error

	ReconcileVirtualNodes(
		ctx context.Context,
		mesh *zephyr_discovery.Mesh,
		virtualNodes []*appmesh.VirtualNodeData,
	) error
}
