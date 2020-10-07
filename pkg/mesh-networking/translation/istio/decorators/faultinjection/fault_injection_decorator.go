package faultinjection

import (
	"github.com/rotisserie/eris"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators"
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
	appliedPolicy *discoveryv1alpha2.TrafficTargetStatus_AppliedTrafficPolicy,
	_ *discoveryv1alpha2.TrafficTarget,
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

func translateFaultInjection(validatedPolicy *v1alpha2.TrafficPolicySpec) (*networkingv1alpha3spec.HTTPFaultInjection, error) {
	faultInjection := validatedPolicy.FaultInjection
	if faultInjection == nil {
		return nil, nil
	}
	if faultInjection.GetFaultInjectionType() == nil {
		return nil, eris.New("FaultInjection type must be specified.")
	}
	var translatedFaultInjection *networkingv1alpha3spec.HTTPFaultInjection
	switch injectionType := faultInjection.GetFaultInjectionType().(type) {
	case *v1alpha2.TrafficPolicySpec_FaultInjection_Abort_:
		translatedFaultInjection = &networkingv1alpha3spec.HTTPFaultInjection{
			Abort: &networkingv1alpha3spec.HTTPFaultInjection_Abort{
				ErrorType: &networkingv1alpha3spec.HTTPFaultInjection_Abort_HttpStatus{
					HttpStatus: faultInjection.GetAbort().GetHttpStatus(),
				},
				Percentage: &networkingv1alpha3spec.Percent{Value: faultInjection.GetPercentage()},
			},
		}
	case *v1alpha2.TrafficPolicySpec_FaultInjection_FixedDelay:
		translatedFaultInjection = &networkingv1alpha3spec.HTTPFaultInjection{
			Delay: &networkingv1alpha3spec.HTTPFaultInjection_Delay{
				HttpDelayType: &networkingv1alpha3spec.HTTPFaultInjection_Delay_FixedDelay{FixedDelay: faultInjection.GetFixedDelay()},
				Percentage:    &networkingv1alpha3spec.Percent{Value: faultInjection.GetPercentage()},
			},
		}
	case *v1alpha2.TrafficPolicySpec_FaultInjection_ExponentialDelay:
		translatedFaultInjection = &networkingv1alpha3spec.HTTPFaultInjection{
			Delay: &networkingv1alpha3spec.HTTPFaultInjection_Delay{
				HttpDelayType: &networkingv1alpha3spec.HTTPFaultInjection_Delay_ExponentialDelay{ExponentialDelay: faultInjection.GetExponentialDelay()},
				Percentage:    &networkingv1alpha3spec.Percent{Value: faultInjection.GetPercentage()},
			},
		}
	default:
		return nil, eris.Errorf("FaultInjection.FaultInjectionType has unexpected type %T", injectionType)
	}
	return translatedFaultInjection, nil
}
