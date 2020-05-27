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
		routeA.Spec.HttpRoute.RetryPolicy != routeB.Spec.HttpRoute.RetryPolicy ||
		routeA.Spec.HttpRoute.Match.Scheme != routeB.Spec.HttpRoute.Match.Scheme ||
		routeA.Spec.HttpRoute.Match.Prefix != routeB.Spec.HttpRoute.Match.Prefix ||
		routeA.Spec.HttpRoute.Match.Method != routeB.Spec.HttpRoute.Match.Method ||
		routeA.Spec.HttpRoute.Match.Headers != nil ||
		routeB.Spec.HttpRoute.Match.Headers != nil {
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
	if virtualNodeA.MeshName != virtualNodeB.MeshName ||
		virtualNodeA.VirtualNodeName != virtualNodeB.VirtualNodeName {
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
	return virtualServiceA.MeshName == virtualServiceB.MeshName &&
		virtualServiceA.VirtualServiceName == virtualServiceB.VirtualServiceName &&
		virtualServiceA.Spec.Provider.VirtualNode.VirtualNodeName == virtualServiceB.Spec.Provider.VirtualNode.VirtualNodeName &&
		virtualServiceA.Spec.Provider.VirtualRouter.VirtualRouterName == virtualServiceB.Spec.Provider.VirtualRouter.VirtualRouterName
}

func (a *appmeshMatcher) AreVirtualRoutersEqual(
	virtualRouterA *appmesh2.VirtualRouterData,
	virtualRouterB *appmesh2.VirtualRouterData,
) bool {
	if virtualRouterA.MeshName != virtualRouterB.MeshName || virtualRouterA.VirtualRouterName != virtualRouterB.VirtualRouterName {
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
