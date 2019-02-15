package plugins

import (
	"github.com/solo-io/supergloo/pkg/api/external/istio/networking/v1alpha3"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

type TimeoutPlugin struct{}

func NewMltsPlugin() *TimeoutPlugin {
	return &TimeoutPlugin{}
}

func (p *TimeoutPlugin) Init(params InitParams) error {
	return nil
}

func (p *TimeoutPlugin) ProcessDestinationRule(params Params, in v1.RoutingRuleSpec, out *v1alpha3.DestinationRule) error {
	return nil
}

func (p *TimeoutPlugin) ProcessVirtualService(params Params, in v1.RoutingRuleSpec, out *v1alpha3.VirtualService) error {

}
