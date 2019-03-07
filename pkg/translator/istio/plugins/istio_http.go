package plugins

import (
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/supergloo/pkg/api/external/istio/networking/v1alpha3"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/translator/utils"
)

type istioHttpPlugin struct{}

func NewIstioHttpPlugin() *istioHttpPlugin {
	return &istioHttpPlugin{}
}

func (p *istioHttpPlugin) Init(params InitParams) error {
	return nil
}

var notImplementedErr = errors.Errorf("rule type not implemented")

func (*istioHttpPlugin) ProcessRoute(params Params, in v1.RoutingRuleSpec, out *v1alpha3.HTTPRoute) error {
	switch rule := in.RuleType.(type) {
	case *v1.RoutingRuleSpec_TrafficShifting:
		if err := processTrafficShiftingRule(params.Upstreams, rule.TrafficShifting, out); err != nil {
			return errors.Wrapf(err, "processing traffic shifting rule")
		}
	case *v1.RoutingRuleSpec_FaultInjection:
		return notImplementedErr
	case *v1.RoutingRuleSpec_RequestTimeout:
		return notImplementedErr
	case *v1.RoutingRuleSpec_Retries:
		return notImplementedErr
	case *v1.RoutingRuleSpec_CorsPolicy:
		return notImplementedErr
	case *v1.RoutingRuleSpec_Mirror:
		return notImplementedErr
	case *v1.RoutingRuleSpec_HeaderManipulation:
		return notImplementedErr
	default:
		return errors.Errorf("unknown rule type %v", in.RuleType)
	}

	return nil
}

func processTrafficShiftingRule(upstreams gloov1.UpstreamList, rule *v1.TrafficShifting, out *v1alpha3.HTTPRoute) error {
	if rule.Destinations == nil || len(rule.Destinations.Destinations) == 0 {
		return errors.Errorf("traffic shifting destinations cannot be missing or empty")
	}
	var shiftedDestinations []*v1alpha3.HTTPRouteDestination

	var totalWeights uint32
	for _, dest := range rule.Destinations.Destinations {
		totalWeights += dest.Weight
	}

	var totalIstioWeights int32
	for i, dest := range rule.Destinations.Destinations {

		if dest.Destination == nil {
			return errors.Errorf("destination %v invalid must provide target upstream", i)
		}

		upstream, err := upstreams.Find(dest.Destination.Upstream.Strings())
		if err != nil {
			return errors.Wrapf(err, "invalid upstream destination")
		}

		host, err := utils.GetHostForUpstream(upstream)
		if err != nil {
			return errors.Wrapf(err, "could not find host for upstream")
		}

		intPort, err := utils.GetPortForUpstream(upstream)
		if err != nil {
			return errors.Wrapf(err, "could not find port for upstream")
		}
		port := &v1alpha3.PortSelector{
			Port: &v1alpha3.PortSelector_Number{Number: intPort},
		}

		labels := utils.GetLabelsForUpstream(upstream)
		if dest.Weight == 0 {
			return errors.Errorf("weight cannot be 0")
		}

		weight := int32(dest.Weight * 100 / totalWeights)
		totalIstioWeights += weight

		shiftedDestinations = append(shiftedDestinations, &v1alpha3.HTTPRouteDestination{
			Destination: &v1alpha3.Destination{
				Host:   host,
				Subset: utils.SubsetName(labels),
				Port:   port,
			},
			Weight: weight,
		})
	}
	// adjust weight in case rounding error occurred
	if weightNeeded := 100 - totalIstioWeights; weightNeeded != 0 {
		shiftedDestinations[0].Weight += weightNeeded
	}
	out.Route = shiftedDestinations

	return nil
}
