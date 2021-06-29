package cors

import (
	"github.com/rotisserie/eris"
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	v1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/gogoutils"
	networkingv1alpha3spec "istio.io/api/networking/v1alpha3"
)

const (
	decoratorName = "cors"
)

func init() {
	decorators.Register(decoratorConstructor)
}

func decoratorConstructor(_ decorators.Parameters) decorators.Decorator {
	return NewCorsDecorator()
}

// handles setting Cors on a VirtualService
type corsDecorator struct{}

var _ decorators.TrafficPolicyVirtualServiceDecorator = &corsDecorator{}

func NewCorsDecorator() *corsDecorator {
	return &corsDecorator{}
}

func (d *corsDecorator) DecoratorName() string {
	return decoratorName
}

func (d *corsDecorator) ApplyTrafficPolicyToVirtualService(
	appliedPolicy *discoveryv1.DestinationStatus_AppliedTrafficPolicy,
	_ *discoveryv1.Destination,
	_ *discoveryv1.MeshInstallation,
	output *networkingv1alpha3spec.HTTPRoute,
	registerField decorators.RegisterField,
) error {
	cors, err := d.translateCors(appliedPolicy.Spec)
	if err != nil {
		return err
	}
	if cors != nil {
		if err := registerField(&output.CorsPolicy, cors); err != nil {
			return err
		}
		output.CorsPolicy = cors
	}
	return nil
}

func (d *corsDecorator) translateCors(
	trafficPolicy *v1.TrafficPolicySpec,
) (*networkingv1alpha3spec.CorsPolicy, error) {
	corsPolicy := trafficPolicy.GetPolicy().GetCorsPolicy()
	if corsPolicy == nil {
		return nil, nil
	}
	var allowOrigins []*networkingv1alpha3spec.StringMatch
	for i, allowOrigin := range corsPolicy.GetAllowOrigins() {
		var stringMatch *networkingv1alpha3spec.StringMatch
		switch matchType := allowOrigin.GetMatchType().(type) {
		case *v1.TrafficPolicySpec_Policy_CorsPolicy_StringMatch_Exact:
			stringMatch = &networkingv1alpha3spec.StringMatch{MatchType: &networkingv1alpha3spec.StringMatch_Exact{Exact: allowOrigin.GetExact()}}
		case *v1.TrafficPolicySpec_Policy_CorsPolicy_StringMatch_Prefix:
			stringMatch = &networkingv1alpha3spec.StringMatch{MatchType: &networkingv1alpha3spec.StringMatch_Prefix{Prefix: allowOrigin.GetPrefix()}}
		case *v1.TrafficPolicySpec_Policy_CorsPolicy_StringMatch_Regex:
			stringMatch = &networkingv1alpha3spec.StringMatch{MatchType: &networkingv1alpha3spec.StringMatch_Regex{Regex: allowOrigin.GetRegex()}}
		default:
			return nil, eris.Errorf("AllowOrigins[%d].MatchType has unexpected type %T", i, matchType)
		}
		allowOrigins = append(allowOrigins, stringMatch)
	}
	translatedCorsPolicy := &networkingv1alpha3spec.CorsPolicy{
		AllowOrigins:     allowOrigins,
		AllowMethods:     corsPolicy.GetAllowMethods(),
		AllowHeaders:     corsPolicy.GetAllowHeaders(),
		ExposeHeaders:    corsPolicy.GetExposeHeaders(),
		MaxAge:           gogoutils.DurationProtoToGogo(corsPolicy.GetMaxAge()),
		AllowCredentials: gogoutils.BoolProtoToGogo(corsPolicy.GetAllowCredentials()),
	}

	return translatedCorsPolicy, nil
}
