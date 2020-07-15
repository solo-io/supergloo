package faultinjection

import (
	"github.com/rotisserie/eris"
	discoveryv1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/istio/meshservice/virtualservice"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/istio/plugins"
	istiov1alpha3spec "istio.io/api/networking/v1alpha3"
)

const (
	pluginName = "fault-injection"
)

func init() {
	plugins.Register(pluginConstructor)
}

func pluginConstructor(params plugins.Parameters) plugins.Plugin {
	return NewFaultInjectionPlugin()
}

// handles setting FaultInjection on a VirtualService
type faultInjectionPlugin struct{}

var _ virtualservice.TrafficPolicyPlugin = &faultInjectionPlugin{}

func NewFaultInjectionPlugin() *faultInjectionPlugin {
	return &faultInjectionPlugin{}
}

func (p *faultInjectionPlugin) PluginName() string {
	return pluginName
}

func (p *faultInjectionPlugin) ProcessTrafficPolicy(
	appliedPolicy *discoveryv1alpha1.MeshServiceStatus_AppliedTrafficPolicy,
	_ *discoveryv1alpha1.MeshService,
	output *istiov1alpha3spec.HTTPRoute,
	registerField virtualservice.RegisterField,
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

func translateFaultInjection(validatedPolicy *v1alpha1.TrafficPolicySpec) (*istiov1alpha3spec.HTTPFaultInjection, error) {
	faultInjection := validatedPolicy.FaultInjection
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
