package clients

import (
	"context"

	appmesh2 "github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/aws/aws-sdk-go/service/sts"
)

//go:generate mockgen -source ./interfaces.go -destination mocks/mock_interfaces.go

/*
	Provide methods that ensure the existence of the given Appmesh resource.
	"Ensure" means create if the resource doesn't exist as identified by its name.
	If it does exist, update it if its spec doesn't match the provided resource spec. Otherwise do nothing.
*/
type AppmeshClient interface {
	ListMeshes(input *appmesh2.ListMeshesInput) (*appmesh2.ListMeshesOutput, error)
	ListTagsForResource(*appmesh2.ListTagsForResourceInput) (*appmesh2.ListTagsForResourceOutput, error)

	EnsureVirtualService(virtualServiceData *appmesh2.VirtualServiceData) error
	EnsureVirtualRouter(virtualRouter *appmesh2.VirtualRouterData) error
	EnsureRoute(route *appmesh2.RouteData) error
	EnsureVirtualNode(virtualNode *appmesh2.VirtualNodeData) error

	ReconcileVirtualRoutersAndRoutesAndVirtualServices(
		ctx context.Context,
		meshName *string,
		virtualRouters []*appmesh2.VirtualRouterData,
		routes []*appmesh2.RouteData,
		virtualServices []*appmesh2.VirtualServiceData,
	) error

	ReconcileVirtualNodes(
		ctx context.Context,
		meshName *string,
		virtualNodes []*appmesh2.VirtualNodeData,
	) error
}

type STSClient interface {
	// Retrieves caller identity metadata by making a request to AWS STS (Secure Token Service).
	GetCallerIdentity() (*sts.GetCallerIdentityOutput, error)
}
