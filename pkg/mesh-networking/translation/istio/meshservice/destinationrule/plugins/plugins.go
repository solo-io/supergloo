package plugins

import (
	discoveryv1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	"github.com/solo-io/smh/pkg/mesh-networking/plugins"
	istiov1alpha3spec "istio.io/api/networking/v1alpha3"
)

/*
The definitive list of plugins used by the translator is contained in ../registered_plugins.go

Adding new plugins requires adding an import statement to that file.
*/

// TrafficPolicyPlugins modify the DestinationRule based on a TrafficPolicy which applies to the MeshService.
type TrafficPolicyPlugin interface {
	plugins.Plugin

	ProcessTrafficPolicy(
		appliedPolicy *discoveryv1alpha1.MeshServiceStatus_AppliedTrafficPolicy,
		service *discoveryv1alpha1.MeshService,
		output *istiov1alpha3spec.DestinationRule,
		registerField plugins.RegisterField,
	) error
}
