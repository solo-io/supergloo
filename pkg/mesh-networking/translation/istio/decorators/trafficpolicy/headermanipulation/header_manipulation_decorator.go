package headermanipulation

import (
	discoveryv1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	"github.com/solo-io/smh/pkg/mesh-networking/decorators"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/istio/decorators/trafficpolicy"
	istiov1alpha3spec "istio.io/api/networking/v1alpha3"
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

var _ trafficpolicy.VirtualServiceDecorator = &headerManipulationDecorator{}

func NewHeaderManipulationDecorator() *headerManipulationDecorator {
	return &headerManipulationDecorator{}
}

func (h *headerManipulationDecorator) DecoratorName() string {
	return decoratorName
}

func (h *headerManipulationDecorator) DecorateVirtualService(
	appliedPolicy *discoveryv1alpha1.MeshServiceStatus_AppliedTrafficPolicy,
	_ *discoveryv1alpha1.MeshService,
	output *istiov1alpha3spec.HTTPRoute,
	registerField decorators.RegisterField,
) error {
	headers := h.translateHeaderManipulation(appliedPolicy.GetSpec())
	if headers != nil {
		if err := registerField(&output.Headers, headers); err != nil {
			return err
		}
		output.Headers = headers
	}
	return nil
}

func (h *headerManipulationDecorator) translateHeaderManipulation(
	trafficPolicy *v1alpha1.TrafficPolicySpec,
) *istiov1alpha3spec.Headers {
	headerManipulation := trafficPolicy.GetHeaderManipulation()
	if headerManipulation == nil {
		return nil
	}
	return &istiov1alpha3spec.Headers{
		Request: &istiov1alpha3spec.Headers_HeaderOperations{
			Add:    headerManipulation.GetAppendRequestHeaders(),
			Remove: headerManipulation.GetRemoveRequestHeaders(),
		},
		Response: &istiov1alpha3spec.Headers_HeaderOperations{
			Add:    headerManipulation.GetAppendResponseHeaders(),
			Remove: headerManipulation.GetRemoveResponseHeaders(),
		},
	}
}
