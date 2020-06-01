package appmesh

import (
	"strings"

	aws2 "github.com/aws/aws-sdk-go/aws"
	appmesh2 "github.com/aws/aws-sdk-go/service/appmesh"
	"k8s.io/apimachinery/pkg/util/sets"
)

type appmeshMatcher struct {
}

func NewAppmeshMatcher() AppmeshMatcher {
	return &appmeshMatcher{}
}

func (a *appmeshMatcher) AreRoutesEqual(routeA *appmesh2.RouteData, routeB *appmesh2.RouteData) bool {
	if aws2.StringValue(routeA.RouteName) != aws2.StringValue(routeB.RouteName) ||
		aws2.StringValue(routeA.MeshName) != aws2.StringValue(routeB.MeshName) ||
		aws2.StringValue(routeA.VirtualRouterName) != aws2.StringValue(routeB.VirtualRouterName) ||
		aws2.Int64Value(routeA.Spec.Priority) != aws2.Int64Value(routeB.Spec.Priority) ||
		routeA.Spec.GrpcRoute != routeB.Spec.GrpcRoute ||
		routeA.Spec.Http2Route != routeB.Spec.Http2Route ||
		routeA.Spec.TcpRoute != routeB.Spec.TcpRoute ||
		aws2.StringValue(routeA.Spec.HttpRoute.Match.Scheme) != aws2.StringValue(routeB.Spec.HttpRoute.Match.Scheme) ||
		aws2.StringValue(routeA.Spec.HttpRoute.Match.Prefix) != aws2.StringValue(routeB.Spec.HttpRoute.Match.Prefix) ||
		aws2.StringValue(routeA.Spec.HttpRoute.Match.Method) != aws2.StringValue(routeB.Spec.HttpRoute.Match.Method) ||
		// TODO: write a matcher for Header and RetryPolicy objects.
		// Until then, always perform update if Headers exist to ensure declared headers.
		routeA.Spec.HttpRoute.Match.Headers != nil ||
		routeB.Spec.HttpRoute.Match.Headers != nil ||
		routeA.Spec.HttpRoute.RetryPolicy != nil ||
		routeB.Spec.HttpRoute.RetryPolicy != nil {
		return false
	}
	weightedTargetsA := map[string]int64{}
	for _, weightedTargetA := range routeA.Spec.HttpRoute.Action.WeightedTargets {
		weightedTargetsA[aws2.StringValue(weightedTargetA.VirtualNode)] = aws2.Int64Value(weightedTargetA.Weight)
	}
	for _, weightedTargetB := range routeB.Spec.HttpRoute.Action.WeightedTargets {
		virtualNodeNameB := aws2.StringValue(weightedTargetB.VirtualNode)
		weightA, ok := weightedTargetsA[virtualNodeNameB]
		if !ok || weightA != aws2.Int64Value(weightedTargetB.Weight) {
			return false
		}
		delete(weightedTargetsA, virtualNodeNameB)
	}
	return len(weightedTargetsA) == 0
}

func (a *appmeshMatcher) AreVirtualNodesEqual(

	virtualNodeA *appmesh2.VirtualNodeData,
	virtualNodeB *appmesh2.VirtualNodeData,
) bool {
	if aws2.StringValue(virtualNodeA.MeshName) != aws2.StringValue(virtualNodeB.MeshName) ||
		aws2.StringValue(virtualNodeA.VirtualNodeName) != aws2.StringValue(virtualNodeB.VirtualNodeName) {
		return false
	}
	// TODO check for listener TLS + Healthcheck?
	var portMappingsA []*appmesh2.PortMapping
	var portMappingsB []*appmesh2.PortMapping
	for _, listener := range virtualNodeA.Spec.Listeners {
		portMappingsA = append(portMappingsA, listener.PortMapping)
	}
	for _, listener := range virtualNodeB.Spec.Listeners {
		portMappingsB = append(portMappingsB, listener.PortMapping)
	}
	if !a.arePortMappingsEqual(portMappingsA, portMappingsB) {
		return false
	}
	virtualServiceBackendsA := sets.NewString()
	virtualServiceBackendsB := sets.NewString()
	for _, backend := range virtualNodeA.Spec.Backends {
		virtualServiceBackendsA.Insert(aws2.StringValue(backend.VirtualService.VirtualServiceName))
	}
	for _, backend := range virtualNodeB.Spec.Backends {
		virtualServiceBackendsB.Insert(aws2.StringValue(backend.VirtualService.VirtualServiceName))
	}
	return virtualServiceBackendsA.Equal(virtualServiceBackendsB)
}

func (a *appmeshMatcher) AreVirtualServicesEqual(
	virtualServiceA *appmesh2.VirtualServiceData,
	virtualServiceB *appmesh2.VirtualServiceData,
) bool {
	if aws2.StringValue(virtualServiceA.MeshName) != aws2.StringValue(virtualServiceB.MeshName) ||
		aws2.StringValue(virtualServiceA.VirtualServiceName) != aws2.StringValue(virtualServiceB.VirtualServiceName) {
		return false
	}
	if virtualServiceA.Spec.Provider.VirtualNode != nil && virtualServiceB.Spec.Provider.VirtualNode != nil {
		return aws2.StringValue(virtualServiceA.Spec.Provider.VirtualNode.VirtualNodeName) ==
			aws2.StringValue(virtualServiceB.Spec.Provider.VirtualNode.VirtualNodeName)
	} else if virtualServiceA.Spec.Provider.VirtualRouter != nil && virtualServiceB.Spec.Provider.VirtualRouter != nil {
		return aws2.StringValue(virtualServiceA.Spec.Provider.VirtualRouter.VirtualRouterName) ==
			aws2.StringValue(virtualServiceB.Spec.Provider.VirtualRouter.VirtualRouterName)
	} else {
		// VirtualServices have different types of Providers
		return false
	}
}

func (a *appmeshMatcher) AreVirtualRoutersEqual(
	virtualRouterA *appmesh2.VirtualRouterData,
	virtualRouterB *appmesh2.VirtualRouterData,
) bool {
	if aws2.StringValue(virtualRouterA.MeshName) != aws2.StringValue(virtualRouterB.MeshName) ||
		aws2.StringValue(virtualRouterA.VirtualRouterName) != aws2.StringValue(virtualRouterB.VirtualRouterName) {
		return false
	}
	var portMappingsA []*appmesh2.PortMapping
	var portMappingsB []*appmesh2.PortMapping
	for _, listener := range virtualRouterA.Spec.Listeners {
		portMappingsA = append(portMappingsA, listener.PortMapping)
	}
	for _, listener := range virtualRouterB.Spec.Listeners {
		portMappingsB = append(portMappingsB, listener.PortMapping)
	}
	return a.arePortMappingsEqual(portMappingsA, portMappingsB)
}

func (a *appmeshMatcher) arePortMappingsEqual(
	portMappingsA []*appmesh2.PortMapping,
	portMappingsB []*appmesh2.PortMapping,
) bool {
	portsA := map[int64]string{}
	for _, port := range portMappingsA {
		portsA[aws2.Int64Value(port.Port)] = strings.ToLower(aws2.StringValue(port.Protocol))
	}
	for _, portB := range portMappingsB {
		protocolA, ok := portsA[aws2.Int64Value(portB.Port)]
		if !ok || strings.ToLower(protocolA) != strings.ToLower(aws2.StringValue(portB.Protocol)) {
			return false
		}
		delete(portsA, aws2.Int64Value(portB.Port))
	}
	return len(portsA) == 0
}
