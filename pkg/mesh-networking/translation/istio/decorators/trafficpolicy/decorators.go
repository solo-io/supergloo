package trafficpolicy

import (
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/decorators"
	"istio.io/api/networking/v1alpha3"
)

/*
	Interface definitions for decorators which take TrafficPolicy as an input and
	decorate a given output resource.
*/

// TrafficPolicyDecorators modify the DestinationRule based on a TrafficPolicy which applies to the MeshService.
type DestinationRuleDecorator interface {
	decorators.Decorator

	DecorateDestinationRule(
		appliedPolicy *v1alpha1.MeshServiceStatus_AppliedTrafficPolicy,
		service *v1alpha1.MeshService,
		output *v1alpha3.DestinationRule,
		registerField decorators.RegisterField,
	) error
}

// TrafficPolicyDecorators modify the VirtualService based on a TrafficPolicy which applies to the MeshService.
type VirtualServiceDecorator interface {
	decorators.Decorator

	DecorateVirtualService(
		appliedPolicy *v1alpha1.MeshServiceStatus_AppliedTrafficPolicy,
		service *v1alpha1.MeshService,
		output *v1alpha3.HTTPRoute,
		registerField decorators.RegisterField,
	) error
}
