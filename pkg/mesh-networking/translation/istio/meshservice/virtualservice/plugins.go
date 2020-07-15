package virtualservice

import (
	discoveryv1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/istio/plugins"
	istiov1alpha3spec "istio.io/api/networking/v1alpha3"
)

type RegisterField func(fieldPtr, val interface{}) error

// TrafficPolicyPlugins modify the VirtualService based on a TrafficPolicy which applies to the MeshService.
type TrafficPolicyPlugin interface {
	plugins.Plugin

	ProcessTrafficPolicy(
		appliedPolicy *discoveryv1alpha1.MeshServiceStatus_AppliedTrafficPolicy,
		service *discoveryv1alpha1.MeshService,
		output *istiov1alpha3spec.HTTPRoute,
		registerField RegisterField,
	) error
}
