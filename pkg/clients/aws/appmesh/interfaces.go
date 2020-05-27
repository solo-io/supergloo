package appmesh

import (
	"github.com/aws/aws-sdk-go/service/appmesh"
)

//go:generate mockgen -source ./interfaces.go -destination mocks/mock_interfaces.go

type AppmeshClient interface {
	EnsureVirtualService(virtualServiceData *appmesh.VirtualServiceData) error
	EnsureVirtualRouter(virtualRouter *appmesh.VirtualRouterData) error
	EnsureRoute(route *appmesh.RouteData) error
	EnsureVirtualNode(virtualNode *appmesh.VirtualNodeData) error
}
