package plugin

import (
	discoveryv1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/snapshot/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	"github.com/solo-io/smh/pkg/mesh-networking/translator/istio/virtualservice/plugin/faultinjection"
	"github.com/solo-io/smh/pkg/mesh-networking/translator/istio/virtualservice/plugin/mirror"
	"github.com/solo-io/smh/pkg/mesh-networking/translator/utils/hostutils"
	istiov1alpha3spec "istio.io/api/networking/v1alpha3"
	istiov1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
)

// Note: Translator plugins should be added here.
// To prevent regressions, the order of plugins defind here should not matter
func makePlugins(
	clusterDomains hostutils.ClusterDomainRegistry,
	in input.Snapshot,
) []Plugin {
	return []Plugin{
		faultinjection.NewFaultInjectionPlugin(),
		mirror.NewMirrorPlugin(clusterDomains, in.MeshServices()),
	}
}

// the plugin Factory initializes Translator plugins on each reconcile
type Factory interface {
	// return a set of plugins built from the given snapshot.
	MakePlugins(
		clusterDomains hostutils.ClusterDomainRegistry,
		in input.Snapshot,
	) []Plugin
}

type factory struct{}

func (f *factory) MakePlugins(
	clusterDomains hostutils.ClusterDomainRegistry,
	in input.Snapshot,
) []Plugin {
	return makePlugins(
		clusterDomains,
		in,
	)
}

// Plugins modify the output VirtualService corresponding to the input MeshService.
type Plugin interface {
	// unique identifier for plugin
	PluginName() string
}

// SimplePlugins only look at the input MeshService when updating the VirtualService.
type SimplePlugin interface {
	Plugin
	Process(service *discoveryv1alpha1.MeshService, output *istiov1alpha3.VirtualService)
}

// TrafficPolicyPlugins modify the VirtualService based on a TrafficPolicy which applies to the MeshService.
type TrafficPolicyPlugin interface {
	Plugin
	ProcessTrafficPolicy(trafficPolicySpec *v1alpha1.TrafficPolicySpec, service *discoveryv1alpha1.MeshService, output *istiov1alpha3spec.HTTPRoute) error
}

// AccessPolicyPlugins modify the VirtualService based on an AccessPolicy which applies to the MeshService.
type AccessPolicyPlugin interface {
	Plugin
	ProcessAccessPolicy(accessPolicySpec *v1alpha1.AccessPolicySpec, service *discoveryv1alpha1.MeshService, output *istiov1alpha3.VirtualService) error
}
