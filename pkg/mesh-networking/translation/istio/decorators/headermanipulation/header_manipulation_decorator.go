package headermanipulation

import (
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	v1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/trafficpolicyutils"
	networkingv1alpha3spec "istio.io/api/networking/v1alpha3"
)

const (
	decoratorName = "header-manipulation"
)

func init() {
	decorators.Register(decoratorConstructor)
}

func decoratorConstructor(_ decorators.Parameters) decorators.Decorator {
	return NewHeaderManipulationDecorator()
}

// Handles setting Headers on a VirtualService.
type headerManipulationDecorator struct{}

var _ decorators.TrafficPolicyVirtualServiceDecorator = &headerManipulationDecorator{}

func NewHeaderManipulationDecorator() *headerManipulationDecorator {
	return &headerManipulationDecorator{}
}

func (d *headerManipulationDecorator) DecoratorName() string {
	return decoratorName
}

func (d *headerManipulationDecorator) ApplyTrafficPolicyToVirtualService(
	appliedPolicy *v1.AppliedTrafficPolicy,
	_ *discoveryv1.Destination,
	_ *discoveryv1.MeshInstallation,
	output *networkingv1alpha3spec.HTTPRoute,
	registerField decorators.RegisterField,
) error {
	headers := d.translateHeaderManipulation(appliedPolicy.Spec)
	if headers != nil {
		if err := registerField(&output.Headers, headers); err != nil {
			return err
		}
		output.Headers = headers
	}
	return nil
}

func (d *headerManipulationDecorator) translateHeaderManipulation(
	trafficPolicy *v1.TrafficPolicySpec,
) *networkingv1alpha3spec.Headers {
	headerManipulation := trafficPolicy.GetPolicy().GetHeaderManipulation()
	return trafficpolicyutils.TranslateHeaderManipulation(headerManipulation)
}
