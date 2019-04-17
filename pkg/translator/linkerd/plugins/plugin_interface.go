package plugins

import (
	"context"

	"github.com/linkerd/linkerd2/controller/gen/apis/serviceprofile/v1alpha1"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"k8s.io/client-go/kubernetes"
)

type InitParams struct {
	Ctx context.Context
}

type Plugin interface {
	// init on first loop, flush cache
	Init(params InitParams) error
}

type Params struct {
	Ctx       context.Context
	Upstreams gloov1.UpstreamList
}

type RoutingPlugin interface {
	Plugin
	ProcessRoutes(params Params, in v1.RoutingRuleSpec, out []*v1alpha1.RouteSpec) error
}

type ServiceProfilePlugin interface {
	Plugin
	ProcessServiceProfile(params Params, in v1.RoutingRuleSpec, out *v1alpha1.ServiceProfileSpec) error
}

type registry struct {
	plugins []Plugin
}

var globalRegistry = func(kc kubernetes.Interface) *registry {
	reg := &registry{}
	// plugins should be added here
	reg.plugins = append(reg.plugins,
		NewRetriesPlugin(),
		//NewLinkerdTimeoutsPlugin(),
	)
	return reg
}

func Plugins(kc kubernetes.Interface) []Plugin {
	return globalRegistry(kc).plugins
}
