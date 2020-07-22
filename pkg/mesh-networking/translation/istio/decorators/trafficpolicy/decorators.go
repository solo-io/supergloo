package trafficpolicy

import (
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/decorators"
	networkingv1alpha3spec "istio.io/api/networking/v1alpha3"
)

/*
	Interface definitions for decorators which take TrafficPolicy as an input and
	decorate a given output resource.
*/

// DestinationRuleDecorators modify the DestinationRule based on a TrafficPolicy which applies to the MeshService.
type DestinationRuleDecorator interface {
	decorators.Decorator

	ApplyToDestinationRule(
		appliedPolicy *v1alpha2.MeshServiceStatus_AppliedTrafficPolicy,
		service *v1alpha2.MeshService,
		output *networkingv1alpha3spec.DestinationRule,
		registerField decorators.RegisterField,
	) error
}

// AggregatingDestinationRuleDecorators modify the DestinationRule based on the entire list of TrafficPolicies which apply to the MeshService.
type AggregatingDestinationRuleDecorator interface {
	decorators.Decorator

	ApplyAllToDestinationRule(
		allAppliedPolicies []*v1alpha2.MeshServiceStatus_AppliedTrafficPolicy,
		output *networkingv1alpha3spec.DestinationRule,
		registerField decorators.RegisterField,
	) error
}

// TrafficPolicyDecorators modify the VirtualService based on a TrafficPolicy which applies to the MeshService.
type VirtualServiceDecorator interface {
	decorators.Decorator

	ApplyToVirtualService(
		appliedPolicy *v1alpha2.MeshServiceStatus_AppliedTrafficPolicy,
		service *v1alpha2.MeshService,
		output *networkingv1alpha3spec.HTTPRoute,
		registerField decorators.RegisterField,
	) error
}
