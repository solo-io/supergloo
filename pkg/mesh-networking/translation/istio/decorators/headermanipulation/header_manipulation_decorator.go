package headermanipulation

import (
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators"
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
	appliedPolicy *discoveryv1alpha2.MeshServiceStatus_AppliedTrafficPolicy,
	_ *discoveryv1alpha2.MeshService,
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
	trafficPolicy *v1alpha2.TrafficPolicySpec,
) *networkingv1alpha3spec.Headers {
	headerManipulation := trafficPolicy.GetHeaderManipulation()
	if headerManipulation == nil {
		return nil
	}
	return &networkingv1alpha3spec.Headers{
		Request: &networkingv1alpha3spec.Headers_HeaderOperations{
			Add:    headerManipulation.GetAppendRequestHeaders(),
			Remove: headerManipulation.GetRemoveRequestHeaders(),
		},
		Response: &networkingv1alpha3spec.Headers_HeaderOperations{
			Add:    headerManipulation.GetAppendResponseHeaders(),
			Remove: headerManipulation.GetRemoveResponseHeaders(),
		},
	}
}
