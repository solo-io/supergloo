package headermanipulation

import (
	discoveryv1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/istio/meshservice/virtualservice/plugins"
	istiov1alpha3spec "istio.io/api/networking/v1alpha3"
)

const (
	pluginName = "header-manipulation"
)

func init() {
	plugins.Register(pluginConstructor)
}

func pluginConstructor(_ plugins.Parameters) plugins.Plugin {
	return NewHeaderManipulationPlugin()
}

// Handles setting Headers on a VirtualService.
type headerManipulationPlugin struct{}

func NewHeaderManipulationPlugin() *headerManipulationPlugin {
	return &headerManipulationPlugin{}
}

func (h *headerManipulationPlugin) PluginName() string {
	return pluginName
}

func (h *headerManipulationPlugin) ProcessTrafficPolicy(
	appliedPolicy *discoveryv1alpha1.MeshServiceStatus_AppliedTrafficPolicy,
	_ *discoveryv1alpha1.MeshService,
	output *istiov1alpha3spec.HTTPRoute,
	registerField plugins.RegisterField,
) error {
	headers := h.translateHeaderManipulation(appliedPolicy.Spec)
	if headers != nil {
		if err := registerField(&output.Headers, headers); err != nil {
			return err
		}
		output.Headers = headers
	}
	return nil
}

func (h *headerManipulationPlugin) translateHeaderManipulation(
	trafficPolicy *v1alpha1.TrafficPolicySpec,
) *istiov1alpha3spec.Headers {
	headerManipulation := trafficPolicy.GetHeaderManipulation()
	if headerManipulation == nil {
		return nil
	}
	return &istiov1alpha3spec.Headers{
		Request: &istiov1alpha3spec.Headers_HeaderOperations{
			Add:    headerManipulation.GetAppendRequestHeaders(),
			Remove: headerManipulation.GetRemoveRequestHeaders(),
		},
		Response: &istiov1alpha3spec.Headers_HeaderOperations{
			Add:    headerManipulation.GetAppendResponseHeaders(),
			Remove: headerManipulation.GetRemoveResponseHeaders(),
		},
	}
}
