package istio

import (
	"context"
	"math"
	"sort"
	"strings"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"

	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	"github.com/solo-io/supergloo/pkg/api/external/istio/networking/v1alpha3"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/translator/istio/plugins"
	"github.com/solo-io/supergloo/pkg/translator/utils"
)

func (t *translator) makeVirtualServiceForHost(
	ctx context.Context,
	params plugins.Params,
	writeNamespace string,
	host string,
	destinationPortAndLabelSets []utils.LabelsPortTuple,
	routingRules v1.RoutingRuleList,
	upstreams gloov1.UpstreamList,
	resourceErrs reporter.ResourceErrors,
) *v1alpha3.VirtualService {

	// group rules by their matcher
	// we will then create a corresponding http rule
	// on the virtual service that contains all the relevant rules
	rulesPerMatcher := utils.NewRulesByMatcher(routingRules)

	vs := initVirtualService(writeNamespace, host)

	var destPorts []uint32
addUniquePorts:
	for _, set := range destinationPortAndLabelSets {
		for _, port := range destPorts {
			if set.Port == port {
				continue addUniquePorts
			}
		}
		destPorts = append(destPorts, set.Port)
	}

	// add a rule for each dest port
	for _, port := range destPorts {
		t.applyRouteRules(
			params,
			host,
			port,
			rulesPerMatcher,
			upstreams,
			resourceErrs,
			vs,
		)
	}

	sortByMatcherSpecificity(vs.Http)

	if len(vs.Http) == 0 {
		// create a default route that sends all traffic to the destination host
		vs.Http = []*v1alpha3.HTTPRoute{{
			Route: []*v1alpha3.HTTPRouteDestination{{
				Destination: &v1alpha3.Destination{Host: host}},
			}},
		}
	}

	return vs
}

func (t *translator) applyRouteRules(
	params plugins.Params,
	destinationHost string,
	destinationPort uint32,
	rulesPerMatcher utils.RulesByMatcher,
	upstreams gloov1.UpstreamList,
	resourceErrs reporter.ResourceErrors,
	out *v1alpha3.VirtualService) {

	var istioRoutes []*v1alpha3.HTTPRoute

	// find rules for this host and apply them
	// each unique matcher becomes an http rule in the virtual
	// service for this host
	for _, rules := range rulesPerMatcher.Sort() {
		// initialize report func
		report := func(err error, format string, args ...interface{}) {
			for _, rr := range rules {
				resourceErrs.AddError(rr, errors.Wrapf(err, format, args...))
			}
		}

		// these should be identical for all the rules
		matcher := rules[0].RequestMatchers
		sourceSelector := rules[0].SourceSelector

		// convert the sourceSelector object to source labels
		sourceLabelSets, err := labelSetsFromSelector(sourceSelector, upstreams)
		if err != nil {
			report(err, "invalid source selector")
			continue
		}

		istioMatcher := createIstioMatcher(sourceLabelSets, destinationPort, matcher)

		route := t.createRoute(
			params,
			destinationHost,
			rules,
			istioMatcher,
			upstreams,
			resourceErrs,
		)

		istioRoutes = append(istioRoutes, route)
	}

	out.Http = append(out.Http, istioRoutes...)
}

func initVirtualService(writeNamespace, host string) *v1alpha3.VirtualService {
	return &v1alpha3.VirtualService{
		Metadata: core.Metadata{
			Namespace: writeNamespace,
			Name:      utils.SanitizeName(host),
		},
		Hosts:    []string{host},
		Gateways: []string{"mesh"},
	}
}

func labelSetsFromSelector(selector *v1.PodSelector, upstreams gloov1.UpstreamList) ([]map[string]string, error) {
	selectedUpstreams, err := utils.UpstreamsForSelector(selector, upstreams)
	if err != nil {
		return nil, errors.Wrapf(err, "selecting upstreams")
	}

	var labelSets []map[string]string
	for _, us := range selectedUpstreams {
		labelSets = append(labelSets, utils.GetLabelsForUpstream(us))
	}

	return labelSets, nil
}

func isCatchAllMatcher(istioMatcher *v1alpha3.HTTPMatchRequest) bool {
	if istioMatcher.Uri == nil {
		return true
	}
	switch path := istioMatcher.Uri.MatchType.(type) {
	case *v1alpha3.StringMatch_Prefix:
		return path.Prefix == "/"
	case *v1alpha3.StringMatch_Regex:
		return path.Regex == "/.*" || path.Regex == ".*"
	}
	return false
}

func shortestPathLength(istioMatchers []*v1alpha3.HTTPMatchRequest) int {
	shortestPath := math.MaxInt64
	for _, m := range istioMatchers {
		switch path := m.Uri.MatchType.(type) {
		case *v1alpha3.StringMatch_Prefix:
			if pathLen := len(path.Prefix); pathLen < shortestPath {
				shortestPath = pathLen
			}
		default:
			continue
		}
	}
	return shortestPath
}

func sortByMatcherSpecificity(istioRoutes []*v1alpha3.HTTPRoute) {
	less := func(i, j int) bool {
		route1, route2 := istioRoutes[i], istioRoutes[j]
		// put catch-all matchers last
		if len(route1.Match) == 0 {
			return false
		}
		for _, match := range route1.Match {
			if isCatchAllMatcher(match) {
				return false
			}
		}
		// heuristic, it's not clear that having a matcher with shorter path length
		// means it matches more stuff
		return shortestPathLength(route1.Match) > shortestPathLength(route2.Match)
	}
	sort.SliceStable(istioRoutes, less)
}

func (t *translator) createRoute(
	params plugins.Params,
	destinationHost string,
	rules v1.RoutingRuleList,
	istioMatcher []*v1alpha3.HTTPMatchRequest,
	upstreams gloov1.UpstreamList,
	resourceErrs reporter.ResourceErrors) *v1alpha3.HTTPRoute {

	out := &v1alpha3.HTTPRoute{
		Match: istioMatcher,

		// default: single destination, original host, no subset
		// traffic shifting may overwrite, so traffic shifting plugin should come first
		Route: []*v1alpha3.HTTPRouteDestination{{
			Destination: &v1alpha3.Destination{
				Host: destinationHost,
			},
		}},
	}
	for _, rr := range rules {
		// if rr does not apply to this host (destination), skip
		useRule, err := utils.RuleAppliesToDestination(destinationHost, rr.DestinationSelector, upstreams)
		if err != nil {
			resourceErrs.AddError(rr, errors.Wrapf(err, "invalid destination selector"))
			continue
		}

		if !useRule {
			continue
		}

		for _, plug := range t.plugins {
			routingPlugin, ok := plug.(plugins.RoutingPlugin)
			if !ok {
				continue
			}
			if rr.Spec == nil {
				resourceErrs.AddError(rr, errors.Errorf("spec cannot be empty"))
				continue
			}
			if err := routingPlugin.ProcessRoute(params, *rr.Spec, out); err != nil {
				resourceErrs.AddError(rr, errors.Wrapf(err, "applying route rule failed"))
			}
		}
	}
	return out
}

func createIstioMatcher(sourceLabelSets []map[string]string, destPort uint32, matcher []*gloov1.Matcher) []*v1alpha3.HTTPMatchRequest {
	var istioMatcher []*v1alpha3.HTTPMatchRequest

	// override for default istioMatcher
	switch {
	case len(matcher) == 0 && len(sourceLabelSets) == 0:
		// default, catch-all istioMatcher is simply nil
	case len(matcher) == 0 && len(sourceLabelSets) > 0:
		for _, sourceLabels := range sourceLabelSets {
			istioMatcher = append(istioMatcher, convertMatcher(sourceLabels, destPort, &gloov1.Matcher{
				PathSpecifier: &gloov1.Matcher_Prefix{
					Prefix: "/",
				},
			}))
		}
	case matcher != nil && len(sourceLabelSets) == 0:
		for _, match := range matcher {
			istioMatcher = append(istioMatcher, convertMatcher(nil, destPort, match))
		}
	case matcher != nil && len(sourceLabelSets) > 0:
		for _, match := range matcher {
			for _, source := range sourceLabelSets {
				istioMatcher = append(istioMatcher, convertMatcher(source, destPort, match))
			}
		}
	}
	return istioMatcher
}

func convertMatcher(sourceSelector map[string]string, destPort uint32, match *gloov1.Matcher) *v1alpha3.HTTPMatchRequest {
	var uri *v1alpha3.StringMatch
	if match.PathSpecifier != nil {
		switch path := match.PathSpecifier.(type) {
		case *gloov1.Matcher_Exact:
			uri = &v1alpha3.StringMatch{
				MatchType: &v1alpha3.StringMatch_Exact{
					Exact: path.Exact,
				},
			}
		case *gloov1.Matcher_Regex:
			uri = &v1alpha3.StringMatch{
				MatchType: &v1alpha3.StringMatch_Regex{
					Regex: path.Regex,
				},
			}
		case *gloov1.Matcher_Prefix:
			uri = &v1alpha3.StringMatch{
				MatchType: &v1alpha3.StringMatch_Prefix{
					Prefix: path.Prefix,
				},
			}
		}
	}
	var methods *v1alpha3.StringMatch
	if len(match.Methods) > 0 {
		methods = &v1alpha3.StringMatch{
			MatchType: &v1alpha3.StringMatch_Regex{
				Regex: strings.Join(match.Methods, "|"),
			},
		}
	}
	var headers map[string]*v1alpha3.StringMatch
	if len(match.Headers) > 0 {
		headers = make(map[string]*v1alpha3.StringMatch)
		for _, v := range match.Headers {
			if v.Regex {
				headers[v.Name] = &v1alpha3.StringMatch{
					MatchType: &v1alpha3.StringMatch_Regex{
						Regex: v.Value,
					},
				}
			} else {
				headers[v.Name] = &v1alpha3.StringMatch{
					MatchType: &v1alpha3.StringMatch_Exact{
						Exact: v.Value,
					},
				}
			}
		}
	}
	return &v1alpha3.HTTPMatchRequest{
		Uri:          uri,
		Method:       methods,
		Headers:      headers,
		SourceLabels: sourceSelector,
		Port:         destPort,
	}
}
