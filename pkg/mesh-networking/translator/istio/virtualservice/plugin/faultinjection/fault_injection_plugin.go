package faultinjection

import (
	"github.com/rotisserie/eris"
	discoveryv1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	istiov1alpha3spec "istio.io/api/networking/v1alpha3"
	"reflect"
)

const (
	pluginName = "fault-injection"
)

// handles setting FaultInjection on a VirtualService
type faultInjectionPlugin struct{}

func NewFaultInjectionPlugin() *faultInjectionPlugin {
	return &faultInjectionPlugin{}
}

func (p *faultInjectionPlugin) PluginName() string {
	return pluginName
}

func (p *faultInjectionPlugin) ProcessTrafficPolicy(trafficPolicySpec *v1alpha1.TrafficPolicySpec,  _ *discoveryv1alpha1.MeshService, output *istiov1alpha3spec.HTTPRoute) error {
	faultInjection, err := translateFaultInjection(trafficPolicySpec)
	if err != nil {
		return err
	}
	if faultInjection != nil {
		if output.Fault != nil  && !reflect.DeepEqual(output.Fault, faultInjection)  {
			return eris.Errorf("fault injection was already defined by a previous traffic policy")
		}
		output.Fault = faultInjection
	}
	return nil
}

func translateFaultInjection(validatedPolicy *v1alpha1.TrafficPolicySpec) (*istiov1alpha3spec.HTTPFaultInjection, error) {
	faultInjection := validatedPolicy.GetFaultInjection()
	if faultInjection == nil {
		return nil, nil
	}
	var translatedFaultInjection *istiov1alpha3spec.HTTPFaultInjection
	switch injectionType := faultInjection.GetFaultInjectionType().(type) {
	case *v1alpha1.TrafficPolicySpec_FaultInjection_Abort_:
		abort := faultInjection.GetAbort()
		switch abortType := abort.GetErrorType().(type) {
		case *v1alpha1.TrafficPolicySpec_FaultInjection_Abort_HttpStatus:
			translatedFaultInjection = &istiov1alpha3spec.HTTPFaultInjection{
				Abort: &istiov1alpha3spec.HTTPFaultInjection_Abort{
					ErrorType:  &istiov1alpha3spec.HTTPFaultInjection_Abort_HttpStatus{HttpStatus: abort.GetHttpStatus()},
					Percentage: &istiov1alpha3spec.Percent{Value: faultInjection.GetPercentage()},
				}}
		default:
			return nil, eris.Errorf("Abort.ErrorType has unexpected type %T", abortType)
		}
	case *v1alpha1.TrafficPolicySpec_FaultInjection_Delay_:
		delay := faultInjection.GetDelay()
		switch delayType := delay.GetHttpDelayType().(type) {
		case *v1alpha1.TrafficPolicySpec_FaultInjection_Delay_FixedDelay:
			translatedFaultInjection = &istiov1alpha3spec.HTTPFaultInjection{
				Delay: &istiov1alpha3spec.HTTPFaultInjection_Delay{
					HttpDelayType: &istiov1alpha3spec.HTTPFaultInjection_Delay_FixedDelay{FixedDelay: delay.GetFixedDelay()},
					Percentage:    &istiov1alpha3spec.Percent{Value: faultInjection.GetPercentage()},
				}}
		case *v1alpha1.TrafficPolicySpec_FaultInjection_Delay_ExponentialDelay:
			translatedFaultInjection = &istiov1alpha3spec.HTTPFaultInjection{
				Delay: &istiov1alpha3spec.HTTPFaultInjection_Delay{
					HttpDelayType: &istiov1alpha3spec.HTTPFaultInjection_Delay_ExponentialDelay{ExponentialDelay: delay.GetExponentialDelay()},
					Percentage:    &istiov1alpha3spec.Percent{Value: faultInjection.GetPercentage()},
				}}
		default:
			return nil, eris.Errorf("Delay.HTTPDelayType has unexpected type %T", delayType)
		}
	default:
		return nil, eris.Errorf("FaultInjection.FaultInjectionType has unexpected type %T", injectionType)
	}
	return translatedFaultInjection, nil
}
