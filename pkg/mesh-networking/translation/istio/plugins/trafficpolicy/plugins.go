package trafficpolicy

import (
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	"github.com/solo-io/smh/pkg/mesh-networking/plugins"
	"istio.io/api/networking/v1alpha3"
)

// TrafficPolicyPlugins modify the DestinationRule based on a TrafficPolicy which applies to the MeshService.
type DestinationRuleDecorator interface {
	plugins.Plugin

	DecorateDestinationRule(
		appliedPolicy *v1alpha1.MeshServiceStatus_AppliedTrafficPolicy,
		service *v1alpha1.MeshService,
		output *v1alpha3.DestinationRule,
		registerField plugins.RegisterField,
	) error
}

// TrafficPolicyPlugins modify the VirtualService based on a TrafficPolicy which applies to the MeshService.
type VirtualServiceDecorator interface {
	plugins.Plugin

	DecorateVirtualService(
		appliedPolicy *v1alpha1.MeshServiceStatus_AppliedTrafficPolicy,
		service *v1alpha1.MeshService,
		output *v1alpha3.HTTPRoute,
		registerField plugins.RegisterField,
	) error
}
