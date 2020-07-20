package cors

import (
	"github.com/rotisserie/eris"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/decorators"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators/trafficpolicy"
	istiov1alpha3spec "istio.io/api/networking/v1alpha3"
)

const (
	decoratorName = "cors"
)

func init() {
	decorators.Register(decoratorConstructor)
}

func decoratorConstructor(params decorators.Parameters) decorators.Decorator {
	return NewCorsDecorator()
}

// handles setting Cors on a VirtualService
type corsDecorator struct{}

var _ trafficpolicy.VirtualServiceDecorator = &corsDecorator{}

func NewCorsDecorator() *corsDecorator {
	return &corsDecorator{}
}

func (d *corsDecorator) DecoratorName() string {
	return decoratorName
}

func (d *corsDecorator) ApplyToVirtualService(
	appliedPolicy *discoveryv1alpha2.MeshServiceStatus_AppliedTrafficPolicy,
	_ *discoveryv1alpha2.MeshService,
	output *istiov1alpha3spec.HTTPRoute,
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
	trafficPolicy *v1alpha2.TrafficPolicySpec,
) (*istiov1alpha3spec.CorsPolicy, error) {
	corsPolicy := trafficPolicy.CorsPolicy
	if corsPolicy == nil {
		return nil, nil
	}
	var allowOrigins []*istiov1alpha3spec.StringMatch
	for i, allowOrigin := range corsPolicy.GetAllowOrigins() {
		var stringMatch *istiov1alpha3spec.StringMatch
		switch matchType := allowOrigin.GetMatchType().(type) {
		case *v1alpha2.TrafficPolicySpec_StringMatch_Exact:
			stringMatch = &istiov1alpha3spec.StringMatch{MatchType: &istiov1alpha3spec.StringMatch_Exact{Exact: allowOrigin.GetExact()}}
		case *v1alpha2.TrafficPolicySpec_StringMatch_Prefix:
			stringMatch = &istiov1alpha3spec.StringMatch{MatchType: &istiov1alpha3spec.StringMatch_Prefix{Prefix: allowOrigin.GetPrefix()}}
		case *v1alpha2.TrafficPolicySpec_StringMatch_Regex:
			stringMatch = &istiov1alpha3spec.StringMatch{MatchType: &istiov1alpha3spec.StringMatch_Regex{Regex: allowOrigin.GetRegex()}}
		default:
			return nil, eris.Errorf("AllowOrigins[%d].MatchType has unexpected type %T", i, matchType)
		}
		allowOrigins = append(allowOrigins, stringMatch)
	}
	translatedCorsPolicy := &istiov1alpha3spec.CorsPolicy{
		AllowOrigins:     allowOrigins,
		AllowMethods:     corsPolicy.GetAllowMethods(),
		AllowHeaders:     corsPolicy.GetAllowHeaders(),
		ExposeHeaders:    corsPolicy.GetExposeHeaders(),
		MaxAge:           corsPolicy.GetMaxAge(),
		AllowCredentials: corsPolicy.GetAllowCredentials(),
	}

	return translatedCorsPolicy, nil
}
