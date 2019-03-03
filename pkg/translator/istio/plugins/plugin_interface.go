package plugins

import (
	"context"

	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"

	"github.com/solo-io/supergloo/pkg/api/external/istio/networking/v1alpha3"
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
	ProcessRoute(params Params, in v1.RoutingRuleSpec, out *v1alpha3.HTTPRoute) error
}

type registry struct {
	plugins []Plugin
}

var globalRegistry = func(kc kubernetes.Interface) *registry {
	reg := &registry{}
	// plugins should be added here
	reg.plugins = append(reg.plugins)
	return reg
}

func Plugins(kc kubernetes.Interface) []Plugin {
	return globalRegistry(kc).plugins
}
