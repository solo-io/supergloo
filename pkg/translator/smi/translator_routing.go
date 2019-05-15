package istio

import (
	"context"
	"math"
	"sort"
	"strings"

	"github.com/solo-io/go-utils/kubeutils"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"

	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	"github.com/solo-io/supergloo/pkg/api/external/istio/networking/v1alpha3"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/translator/istio/plugins"
	"github.com/solo-io/supergloo/pkg/translator/utils"
)

func routingRulesForHost(host string, routingRules v1.RoutingRuleList, upstreams gloov1.UpstreamList) (v1.RoutingRuleList, error) {

	// routing rules for this vs
	var myRoutingRules v1.RoutingRuleList
	for _, rr := range routingRules {
		ruleAppliesToHost, err := utils.HostnameSelected(host, rr.DestinationSelector, upstreams)
		if err != nil {
			return nil, err
		}
		if ruleAppliesToHost {

			myRoutingRules = append(myRoutingRules, rr)
		}
	}

	return myRoutingRules, nil
}

func matcherForRule(sources []map[string]string, upstreamsForService gloov1.UpstreamList, rule *v1.RoutingRule) ([]*v1alpha3.HTTPMatchRequest, error) {
	// no sources == all sources
	if len(sources) == 0 {
		sources = []map[string]string{{}}
	}
	upstreamsForRule, err := utils.UpstreamsForSelector(rule.DestinationSelector, upstreamsForService)
	if err != nil {
		return nil, err
	}

	portsForMatcher, err := utils.PortsFromUpstreams(upstreamsForRule)
	if err != nil {
		return nil, err
	}

	// if no request matchers specified, create a catchall /
	requestMatchers := rule.RequestMatchers
	if len(requestMatchers) == 0 {
		requestMatchers = []*gloov1.Matcher{{
			PathSpecifier: &gloov1.Matcher_Prefix{
				Prefix: "/",
			},
		}}
	}

	// create a separate match for
	// each source label set
	// each port
	// each req matcher
	var matches []*v1alpha3.HTTPMatchRequest
	for _, port := range portsForMatcher {
		for _, sourceLabels := range sources {
			for _, sgMatcher := range requestMatchers {
				istioMatch := convertMatcher(sourceLabels, port, sgMatcher)
				matches = append(matches, istioMatch)
			}
		}
	}

	return matches, nil
}

// calculates the matcher overlap
// returns the union between the two matchers
// plus the unique from list1, list2
func matchersUnion(matchers1, matchers2 []*v1alpha3.HTTPMatchRequest) ([]*v1alpha3.HTTPMatchRequest, []*v1alpha3.HTTPMatchRequest, []*v1alpha3.HTTPMatchRequest) {
	var uniqueList1, uniqueList2, overlap []*v1alpha3.HTTPMatchRequest
	for _, m1 := range matchers1 {
		isUnique := true
		for _, m2 := range matchers2 {
			if m1.Equal(m2) {
				overlap = append(overlap, m1)
				isUnique = false
				break
			}
		}
		if isUnique {
			uniqueList1 = append(uniqueList1, m1)
		}
	}
	for _, m2 := range matchers2 {
		isUnique := true
		for _, m1 := range matchers1 {
			if m1.Equal(m2) {
				isUnique = false
				break
			}
		}
		if isUnique {
			uniqueList2 = append(uniqueList2, m2)
		}
	}

	return overlap, uniqueList1, uniqueList2
}

func (t *translator) makeVirtualServiceForHost(
	ctx context.Context,
	params plugins.Params,
	writeNamespace string,
	service *utils.UpstreamService,
	allRoutingRules v1.RoutingRuleList,
	upstreams gloov1.UpstreamList,
	resourceErrs reporter.ResourceErrors,
) (*v1alpha3.VirtualService, error) {
	routingRules, err := routingRulesForHost(service.Host, allRoutingRules, upstreams)
	if err != nil {
		return nil, err
	}
	// no need to create a virtual service, no rules will be applied
	if len(routingRules) == 0 {
		return nil, nil
	}

	// make a route for each rule
	// merge by matcher
	var routes []*v1alpha3.HTTPRoute
	for _, rule := range routingRules {
		sourceUpstreams, err := utils.UpstreamsForSelector(rule.SourceSelector, upstreams)
		if err != nil {
			return nil, err
		}

		// TODO(ilackarms): support filtering upstreams by
		// the mesh they belong to. we should not be comparing
		// source upstreams to all upstreams, but only upstreams
		// in our mesh. we can use dicovery data for this
		//
		// if the user specified a subset of sources,
		// we must add them to the matchers
		// nil case is valid here (all upstreams)
		var sourceLabelSets []map[string]string
		if len(sourceUpstreams) < len(upstreams) {
			sourceLabelSets, err = utils.LabelsFromUpstreams(sourceUpstreams)
			if err != nil {
				return nil, err
			}
		}

		matcher, err := matcherForRule(sourceLabelSets, service.Upstreams, rule)
		if err != nil {
			resourceErrs.AddError(rule, err)
			continue
		}
		// we should apply our feature to previously
		// created routes that may have an overlapping matcher
		// this way a feature does not get blocked by a
		// matcher with higher order in the list
		for _, route := range routes {
			overlap, _, newMatcher := matchersUnion(route.Match, matcher)
			// remove the matchers that overlapped with this route
			matcher = newMatcher

			// apply the rule to the overlapping route
			if len(overlap) > 0 {
				t.applyRuleToRoute(params, route, rule, resourceErrs)
			}
		}
		// make a new route for our with the remaining matchers
		if len(matcher) > 0 {
			route := &v1alpha3.HTTPRoute{
				Match: matcher,
				// default: single destination, original host, no subset
				// traffic shifting may overwrite, so traffic shifting plugin should come first
				Route: []*v1alpha3.HTTPRouteDestination{{
					Destination: &v1alpha3.Destination{
						Host: service.Host,
					},
				}},
			}
			t.applyRuleToRoute(params, route, rule, resourceErrs)
			routes = append(routes, route)
		}
	}

	sortByMatcherSpecificity(routes)

	return &v1alpha3.VirtualService{
		Metadata: core.Metadata{
			Namespace: writeNamespace,
			Name:      kubeutils.SanitizeName(service.Host),
		},
		Hosts:    []string{service.Host},
		Gateways: []string{"mesh"},
		Http:     routes,
	}, nil
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

func (t *translator) applyRuleToRoute(params plugins.Params, route *v1alpha3.HTTPRoute, rr *v1.RoutingRule, resourceErrs reporter.ResourceErrors) {

	for _, plug := range t.plugins {
		routingPlugin, ok := plug.(plugins.RoutingPlugin)
		if !ok {
			continue
		}
		if rr.Spec == nil {
			resourceErrs.AddError(rr, errors.Errorf("spec cannot be empty"))
			continue
		}
		if err := routingPlugin.ProcessRoute(params, *rr.Spec, route); err != nil {
			resourceErrs.AddError(rr, errors.Wrapf(err, "applying route rule failed"))
		}
	}
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
