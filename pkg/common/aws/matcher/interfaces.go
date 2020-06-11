package matcher

import "github.com/aws/aws-sdk-go/service/appmesh"

//go:generate mockgen -source ./interfaces.go -destination mocks/mock_interfaces.go

type AppmeshMatcher interface {
	AreRoutesEqual(
		routeA *appmesh.RouteData,
		routeB *appmesh.RouteData,
	) bool

	AreVirtualNodesEqual(
		virtualNodeA *appmesh.VirtualNodeData,
		virtualNodeB *appmesh.VirtualNodeData,
	) bool

	AreVirtualServicesEqual(
		virtualServiceA *appmesh.VirtualServiceData,
		virtualServiceB *appmesh.VirtualServiceData,
	) bool

	AreVirtualRoutersEqual(
		virtualRouterA *appmesh.VirtualRouterData,
		virtualRouterB *appmesh.VirtualRouterData,
	) bool
}
