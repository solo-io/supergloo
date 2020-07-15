package timeout

import (
	"github.com/gogo/protobuf/types"
	discoveryv1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/istio/meshservice/virtualservice"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/istio/plugins"
	istiov1alpha3spec "istio.io/api/networking/v1alpha3"
)

const (
	pluginName = "timeout"
)

func init() {
	plugins.Register(pluginConstructor)
}

func pluginConstructor(params plugins.Parameters) plugins.Plugin {
	return NewTimeoutPlugin()
}

// handles setting Timeout on a VirtualService
type timeoutPlugin struct {
}

var _ virtualservice.TrafficPolicyPlugin = &timeoutPlugin{}

func NewTimeoutPlugin() *timeoutPlugin {
	return &timeoutPlugin{
	}
}

func (p *timeoutPlugin) PluginName() string {
	return pluginName
}

func (p *timeoutPlugin) ProcessTrafficPolicy(
	appliedPolicy *discoveryv1alpha1.MeshServiceStatus_AppliedTrafficPolicy,
	_ *discoveryv1alpha1.MeshService,
	output *istiov1alpha3spec.HTTPRoute,
	registerField virtualservice.RegisterField,
) error {
	timeout, err := p.translateTimeout(appliedPolicy.Spec)
	if err != nil {
		return err
	}
	if timeout != nil {
		if err := registerField(&output.Timeout, timeout); err != nil {
			return err
		}
		output.Timeout = timeout
	}
	return nil
}

func (p *timeoutPlugin) translateTimeout(
	trafficPolicy *v1alpha1.TrafficPolicySpec,
) (*types.Duration, error) {
	return trafficPolicy.RequestTimeout, nil
}
