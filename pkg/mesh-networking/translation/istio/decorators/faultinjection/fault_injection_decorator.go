package faultinjection

import (
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	v1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/trafficpolicyutils"
	networkingv1alpha3spec "istio.io/api/networking/v1alpha3"
)

const (
	decoratorName = "fault-injection"
)

func init() {
	decorators.Register(decoratorConstructor)
}

func decoratorConstructor(params decorators.Parameters) decorators.Decorator {
	return NewFaultInjectionDecorator()
}

// handles setting FaultInjection on a VirtualService
type faultInjectionDecorator struct{}

var _ decorators.TrafficPolicyVirtualServiceDecorator = &faultInjectionDecorator{}

func NewFaultInjectionDecorator() *faultInjectionDecorator {
	return &faultInjectionDecorator{}
}

func (d *faultInjectionDecorator) DecoratorName() string {
	return decoratorName
}

func (d *faultInjectionDecorator) ApplyTrafficPolicyToVirtualService(
	appliedPolicy *v1.AppliedTrafficPolicy,
	_ *discoveryv1.Destination,
	_ *discoveryv1.MeshInstallation,
	output *networkingv1alpha3spec.HTTPRoute,
	registerField decorators.RegisterField,
) error {
	faultInjection, err := translateFaultInjection(appliedPolicy.Spec)
	if err != nil {
		return err
	}
	if faultInjection != nil {
		if err := registerField(&output.Fault, faultInjection); err != nil {
			return err
		}
		output.Fault = faultInjection
	}
	return nil
}

func translateFaultInjection(validatedPolicy *v1.TrafficPolicySpec) (*networkingv1alpha3spec.HTTPFaultInjection, error) {
	faultInjection := validatedPolicy.GetPolicy().GetFaultInjection()
	return trafficpolicyutils.TranslateFault(faultInjection)
}
