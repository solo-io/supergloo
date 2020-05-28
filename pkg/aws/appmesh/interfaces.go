package appmesh

import (
	appmesh2 "github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/aws/aws-sdk-go/service/sts"
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

	// For a given MeshService, return a Route that targets all its backing workloads with equal weight.
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

/*
	Provide methods that ensure the existence of the given Appmesh resource.
	"Ensure" means create if the resource doesn't exist as identified by its name.
	If it does exist, update it if its spec doesn't match the provided resource spec. Otherwise do nothing.
*/
type AppmeshClient interface {
	EnsureVirtualService(virtualServiceData *appmesh2.VirtualServiceData) error
	EnsureVirtualRouter(virtualRouter *appmesh2.VirtualRouterData) error
	EnsureRoute(route *appmesh2.RouteData) error
	EnsureVirtualNode(virtualNode *appmesh2.VirtualNodeData) error

	DeleteAllDefaultRoutes(meshName string) error
}

type STSClient interface {
	// Retrieves caller identity metadata by making a request to AWS STS (Secure Token Service).
	GetCallerIdentity() (*sts.GetCallerIdentityOutput, error)
}
