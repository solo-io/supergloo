package destinationrule

import (
	discoveryv1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/istio/plugins"
	istiov1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
)

// SimplePlugins only look at the input MeshService when updating the DestinationRule.
type Plugin interface {
	plugins.Plugin
	Process(service *discoveryv1alpha1.MeshService, output *istiov1alpha3.DestinationRule)
}
