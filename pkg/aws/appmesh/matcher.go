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
	rA := Route{Data: routeA}
	rB := Route{Data: routeB}
	if rA.Name() != rB.Name() ||
		rA.MeshName() != rB.MeshName() ||
		rA.VirtualRouterName() != rB.VirtualRouterName() ||
		rA.Priority() != rB.Priority() ||
		routeA.Spec.GrpcRoute != routeB.Spec.GrpcRoute ||
		routeA.Spec.Http2Route != routeB.Spec.Http2Route ||
		routeA.Spec.TcpRoute != routeB.Spec.TcpRoute ||
		rA.Scheme() != rB.Scheme() ||
		rA.Prefix() != rB.Prefix() ||
		rA.Method() != rB.Method() ||
		// TODO: write a matcher for Header and RetryPolicy objects.
		// Until then, always perform update if Headers exist to ensure declared headers.
		routeA.Spec.HttpRoute.Match.Headers != nil ||
		routeB.Spec.HttpRoute.Match.Headers != nil ||
		routeA.Spec.HttpRoute.RetryPolicy != nil ||
		routeB.Spec.HttpRoute.RetryPolicy != nil {
		return false
	}
	weightedTargetsA := map[string]int64{}
	for _, weightedTargetA := range rA.WeightedTargets() {
		weightedTargetsA[aws2.StringValue(weightedTargetA.VirtualNode)] = aws2.Int64Value(weightedTargetA.Weight)
	}
	for _, weightedTargetB := range rB.WeightedTargets() {
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
	vnA := VirtualNode{Data: virtualNodeA}
	vnB := VirtualNode{Data: virtualNodeB}
	if vnA.Name() != vnB.Name() || vnA.MeshName() != vnB.MeshName() || vnA.HostName() != vnB.HostName() {
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
	vsA := VirtualService{Data: virtualServiceA}
	vsB := VirtualService{Data: virtualServiceB}
	return vsA.Name() == vsB.Name() &&
		vsA.MeshName() == vsB.MeshName() &&
		vsA.VirtualNodeName() == vsB.VirtualNodeName() &&
		vsA.VirtualRouterName() == vsB.VirtualRouterName()
}

func (a *appmeshMatcher) AreVirtualRoutersEqual(
	virtualRouterA *appmesh2.VirtualRouterData,
	virtualRouterB *appmesh2.VirtualRouterData,
) bool {
	vrA := VirtualRouter{Data: virtualRouterA}
	vrB := VirtualRouter{Data: virtualRouterB}
	if vrA.Name() != vrB.Name() || vrA.MeshName() != vrB.MeshName() {
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
