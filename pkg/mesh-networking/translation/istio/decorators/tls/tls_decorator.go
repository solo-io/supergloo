package tls

import (
	"github.com/rotisserie/eris"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators"
	networkingv1alpha3spec "istio.io/api/networking/v1alpha3"
)

const (
	decoratorName = "tls"
)

func init() {
	decorators.Register(decoratorConstructor)
}

func decoratorConstructor(_ decorators.Parameters) decorators.Decorator {
	return NewTlsDecorator()
}

// Handles setting TLS on a DestinationRule.
type tlsDecorator struct{}

var _ decorators.TrafficPolicyDestinationRuleDecorator = &tlsDecorator{}

func NewTlsDecorator() *tlsDecorator {
	return &tlsDecorator{}
}

func (d *tlsDecorator) DecoratorName() string {
	return decoratorName
}

func (d *tlsDecorator) ApplyTrafficPolicyToDestinationRule(
	appliedPolicy *discoveryv1alpha2.TrafficTargetStatus_AppliedTrafficPolicy,
	_ *discoveryv1alpha2.TrafficTarget,
	output *networkingv1alpha3spec.DestinationRule,
	registerField decorators.RegisterField,
) error {
	tlsSettings, err := d.translateTlsSettings(appliedPolicy.Spec)
	if err != nil {
		return err
	}
	if tlsSettings != nil {
		if err := registerField(&output.TrafficPolicy.Tls, tlsSettings); err != nil {
			return err
		}
		output.TrafficPolicy.Tls = tlsSettings
	}

	return nil
}

func (d *tlsDecorator) translateTlsSettings(
	trafficPolicy *v1alpha2.TrafficPolicySpec,
) (*networkingv1alpha3spec.ClientTLSSettings, error) {
	// If TrafficPolicy doesn't specify mTLS configuration, use global default populated upstream during initialization.
	if trafficPolicy.GetMtls().GetIstio() == nil {
		return nil, nil
	}
	istioTlsMode, err := MapIstioTlsMode(trafficPolicy.Mtls.Istio.TlsMode)
	if err != nil {
		return nil, err
	}
	return &networkingv1alpha3spec.ClientTLSSettings{
		Mode: istioTlsMode,
	}, nil
}

func MapIstioTlsMode(tlsMode v1alpha2.TrafficPolicySpec_MTLS_Istio_TLSmode) (networkingv1alpha3spec.ClientTLSSettings_TLSmode, error) {
	switch tlsMode {
	case v1alpha2.TrafficPolicySpec_MTLS_Istio_DISABLE:
		return networkingv1alpha3spec.ClientTLSSettings_DISABLE, nil
	case v1alpha2.TrafficPolicySpec_MTLS_Istio_SIMPLE:
		return networkingv1alpha3spec.ClientTLSSettings_SIMPLE, nil
	case v1alpha2.TrafficPolicySpec_MTLS_Istio_ISTIO_MUTUAL:
		return networkingv1alpha3spec.ClientTLSSettings_ISTIO_MUTUAL, nil
	default:
		return 0, eris.Errorf("unrecognized Istio TLS mode %s", tlsMode)
	}
}
