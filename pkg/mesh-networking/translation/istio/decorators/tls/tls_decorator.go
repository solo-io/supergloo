package tls

import (
	"github.com/rotisserie/eris"
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	v1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators"
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
	appliedPolicy *v1.AppliedTrafficPolicy,
	_ *discoveryv1.Destination,
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
	trafficPolicy *v1.TrafficPolicySpec,
) (*networkingv1alpha3spec.ClientTLSSettings, error) {
	// If TrafficPolicy doesn't specify mTLS configuration, use global default populated upstream during initialization.
	istioMtls := trafficPolicy.GetPolicy().GetMtls().GetIstio()
	if istioMtls == nil {
		return nil, nil
	}
	istioTlsMode, err := MapIstioTlsMode(istioMtls.TlsMode)
	if err != nil {
		return nil, err
	}
	return &networkingv1alpha3spec.ClientTLSSettings{
		Mode: istioTlsMode,
	}, nil
}

func MapIstioTlsMode(tlsMode v1.TrafficPolicySpec_Policy_MTLS_Istio_TLSmode) (networkingv1alpha3spec.ClientTLSSettings_TLSmode, error) {
	switch tlsMode {
	case v1.TrafficPolicySpec_Policy_MTLS_Istio_DISABLE:
		return networkingv1alpha3spec.ClientTLSSettings_DISABLE, nil
	case v1.TrafficPolicySpec_Policy_MTLS_Istio_SIMPLE:
		return networkingv1alpha3spec.ClientTLSSettings_SIMPLE, nil
	case v1.TrafficPolicySpec_Policy_MTLS_Istio_ISTIO_MUTUAL:
		return networkingv1alpha3spec.ClientTLSSettings_ISTIO_MUTUAL, nil
	default:
		return 0, eris.Errorf("unrecognized Istio TLS mode %s", tlsMode)
	}
}
