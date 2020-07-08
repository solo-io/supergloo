package plugin

import (
	discoveryv1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/snapshot/input"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/istio/destinationrule/plugin/mtls"
	istiov1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
)

// Note: Translator plugins should be added here.
// To prevent regressions, the order of plugins defind here should not matter
func makePlugins(in input.Snapshot) []Plugin {
	return []Plugin{
		mtls.NewMtlsPlugin(),
	}
}

// the plugin Factory initializes Translator plugins on each reconcile
type Factory interface {
	// return a set of plugins built from the given snapshot.
	MakePlugins(in input.Snapshot) []Plugin
}

type factory struct{}

func NewFactory() *factory {
	return &factory{}
}

func (f *factory) MakePlugins(in input.Snapshot) []Plugin {
	return makePlugins(in)
}

// Plugins modify the output DestinationRule corresponding to the input MeshService.
type Plugin interface {
	// unique identifier for plugin
	PluginName() string
}

// SimplePlugins only look at the input MeshService when updating the DestinationRule.
type SimplePlugin interface {
	Plugin
	Process(service *discoveryv1alpha1.MeshService, output *istiov1alpha3.DestinationRule)
}

// TrafficPolicyPlugins modify the DestinationRule based on a TrafficPolicy which applies to the MeshService.
type TrafficPolicyPlugin interface {
	Plugin
	ProcessTrafficPolicy(trafficPolicySpec *v1alpha1.TrafficPolicySpec, service *discoveryv1alpha1.MeshService, output *istiov1alpha3.DestinationRule) error
}

// AccessPolicyPlugins modify the DestinationRule based on an AccessPolicy which applies to the MeshService.
type AccessPolicyPlugin interface {
	Plugin
	ProcessAccessPolicy(accessPolicySpec *v1alpha1.AccessPolicySpec, service *discoveryv1alpha1.MeshService, output *istiov1alpha3.DestinationRule) error
}
