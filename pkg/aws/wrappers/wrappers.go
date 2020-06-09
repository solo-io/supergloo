package wrappers

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appmesh"
)

/*
	Nil-safe field accessors on Appmesh objects.
*/

type VirtualService struct {
	Data *appmesh.VirtualServiceData
}

func (v *VirtualService) Name() string {
	return aws.StringValue(v.Data.VirtualServiceName)
}

func (v *VirtualService) MeshName() string {
	return aws.StringValue(v.Data.MeshName)
}

func (v *VirtualService) VirtualRouterName() string {
	if v.Data.Spec.Provider != nil &&
		v.Data.Spec.Provider.VirtualRouter != nil &&
		v.Data.Spec.Provider.VirtualRouter.VirtualRouterName != nil {
		return aws.StringValue(v.Data.Spec.Provider.VirtualRouter.VirtualRouterName)
	}
	return ""
}

func (v *VirtualService) VirtualNodeName() string {
	if v.Data.Spec.Provider != nil &&
		v.Data.Spec.Provider.VirtualNode != nil &&
		v.Data.Spec.Provider.VirtualNode.VirtualNodeName != nil {
		return aws.StringValue(v.Data.Spec.Provider.VirtualNode.VirtualNodeName)
	}
	return ""
}

type VirtualNode struct {
	Data *appmesh.VirtualNodeData
}

func (v *VirtualNode) Name() string {
	return aws.StringValue(v.Data.VirtualNodeName)
}

func (v *VirtualNode) MeshName() string {
	return aws.StringValue(v.Data.MeshName)
}

func (v *VirtualNode) HostName() string {
	if v.Data.Spec.ServiceDiscovery != nil &&
		v.Data.Spec.ServiceDiscovery.Dns != nil {
		return aws.StringValue(v.Data.Spec.ServiceDiscovery.Dns.Hostname)
	}
	return ""
}

type VirtualRouter struct {
	Data *appmesh.VirtualRouterData
}

func (v *VirtualRouter) Name() string {
	return aws.StringValue(v.Data.VirtualRouterName)
}

func (v *VirtualRouter) MeshName() string {
	return aws.StringValue(v.Data.MeshName)
}

type Route struct {
	Data *appmesh.RouteData
}

func (r *Route) Name() string {
	return aws.StringValue(r.Data.RouteName)
}

func (r *Route) MeshName() string {
	return aws.StringValue(r.Data.MeshName)
}

func (r *Route) VirtualRouterName() string {
	return aws.StringValue(r.Data.VirtualRouterName)
}

func (r *Route) Priority() int {
	return int(aws.Int64Value(r.Data.Spec.Priority))
}

func (r *Route) Prefix() string {
	if r.Data.Spec.HttpRoute != nil &&
		r.Data.Spec.HttpRoute.Match != nil {
		return aws.StringValue(r.Data.Spec.HttpRoute.Match.Prefix)
	}
	return ""
}

func (r *Route) Scheme() string {
	if r.Data.Spec.HttpRoute != nil &&
		r.Data.Spec.HttpRoute.Match != nil {
		return aws.StringValue(r.Data.Spec.HttpRoute.Match.Scheme)
	}
	return ""
}

func (r *Route) Method() string {
	if r.Data.Spec.HttpRoute != nil &&
		r.Data.Spec.HttpRoute.Match != nil {
		return aws.StringValue(r.Data.Spec.HttpRoute.Match.Method)
	}
	return ""
}

func (r *Route) WeightedTargets() []*appmesh.WeightedTarget {
	if r.Data.Spec.HttpRoute != nil &&
		r.Data.Spec.HttpRoute.Action != nil {
		return r.Data.Spec.HttpRoute.Action.WeightedTargets
	}
	return nil
}
