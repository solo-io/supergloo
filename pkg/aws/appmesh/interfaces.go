package appmesh

import (
	appmesh2 "github.com/aws/aws-sdk-go/service/appmesh"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
)

//go:generate mockgen -source ./interfaces.go -destination mocks/mock_interfaces.go

type AppmeshMatcher interface {
	AreRoutesEqual(
		routeA *appmesh2.RouteData,
		routeB *appmesh2.RouteData,
	) bool

	AreVirtualNodesEqual(
		virtualNodeA *appmesh2.VirtualNodeData,
		virtualNodeB *appmesh2.VirtualNodeData,
	) bool

	AreVirtualServicesEqual(
		virtualServiceA *appmesh2.VirtualServiceData,
		virtualServiceB *appmesh2.VirtualServiceData,
	) bool

	AreVirtualRoutersEqual(
		virtualRouterA *appmesh2.VirtualRouterData,
		virtualRouterB *appmesh2.VirtualRouterData,
	) bool
}

type AppmeshTranslator interface {
	// For a given MeshWorkload, return a VirtualNode with the given upstream services declared as VirtualService backends.
	BuildDefaultVirtualNode(
		appmeshName *string,
		meshWorkload *zephyr_discovery.MeshWorkload,
		meshService *zephyr_discovery.MeshService,
		upstreamServices []*zephyr_discovery.MeshService,
	) *appmesh2.VirtualNodeData

	// For a given MeshService, return a Route that targets the given MeshWorkloads with equal weight.
	BuildDefaultRoute(
		appmeshName *string,
		meshService *zephyr_discovery.MeshService,
		meshWorkloads []*zephyr_discovery.MeshWorkload,
	) (*appmesh2.RouteData, error)

	// For a given MeshService, return a VirtualService.
	BuildVirtualService(
		appmeshName *string,
		meshService *zephyr_discovery.MeshService,
	) *appmesh2.VirtualServiceData

	// For a given MeshService, return a VirtualRouter.
	BuildVirtualRouter(
		appmeshName *string,
		meshService *zephyr_discovery.MeshService,
	) *appmesh2.VirtualRouterData
}
