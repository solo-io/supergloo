package istio

import (
	"context"
	"math"
	"sort"
	"strings"

	"github.com/solo-io/solo-kit/pkg/utils/hashutils"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"

	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	"github.com/solo-io/supergloo/pkg/api/external/istio/networking/v1alpha3"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/translator/istio/plugins"
	"github.com/solo-io/supergloo/pkg/translator/utils"
	"k8s.io/apimachinery/pkg/labels"
)

// we create a routing rule for each unique matcher
type rulesByMatcher struct {
	rules map[uint64]v1.RoutingRuleList
}

func newRulesByMatcher(rules v1.RoutingRuleList) rulesByMatcher {
	rbm := make(map[uint64]v1.RoutingRuleList)
	for _, rule := range rules {
		hash := hashutils.HashAll(
			rule.SourceSelector,
			rule.RequestMatchers,
		)
		rbm[hash] = append(rbm[hash], rule)
	}

	return rulesByMatcher{rules: rbm}
}

func (rbm rulesByMatcher) sort() []v1.RoutingRuleList {
	var (
		hashes       []uint64
		rulesForHash []v1.RoutingRuleList
	)
	for hash, rules := range rbm.rules {
		hashes = append(hashes, hash)
		rulesForHash = append(rulesForHash, rules)
	}
	sort.SliceStable(rulesForHash, func(i, j int) bool {
		return hashes[i] < hashes[j]
	})
	return rulesForHash
}

func (t *translator) makeVirtualServiceForHost(
	ctx context.Context,
	params plugins.Params,
	writeNamespace string,
	host string,
	destinationPortAndLabelSets []labelsPortTuple,
	routingRules v1.RoutingRuleList,
	upstreams gloov1.UpstreamList,
	resourceErrs reporter.ResourceErrors,
) *v1alpha3.VirtualService {

	// group rules by their matcher
	// we will then create a corresponding http rule
	// on the virtual service that contains all the relevant rules
	rulesPerMatcher := newRulesByMatcher(routingRules)

	vs := initVirtualService(writeNamespace, host)

	// add a rule for each dest port
	for _, set := range destinationPortAndLabelSets {
		t.applyRouteRules(
			params,
			host,
			set.port,
			rulesPerMatcher,
			upstreams,
			resourceErrs,
			vs,
		)
	}

	return vs
}

func (t *translator) applyRouteRules(
	params plugins.Params,
	destinationHost string,
	destinationPort uint32,
	rulesPerMatcher rulesByMatcher,
	upstreams gloov1.UpstreamList,
	resourceErrs reporter.ResourceErrors,
	out *v1alpha3.VirtualService) {

	var istioRoutes []*v1alpha3.HTTPRoute

	// find rules for this host and apply them
	// each unique matcher becomes an http rule in the virtual
	// service for this host
	for _, rules := range rulesPerMatcher.sort() {
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

	sortByMatcherSpecificity(istioRoutes)

	if len(istioRoutes) == 0 {
		// create a default route that sends all traffic to the destination host
		istioRoutes = []*v1alpha3.HTTPRoute{{
			Route: []*v1alpha3.HTTPRouteDestination{{
				Destination: &v1alpha3.Destination{Host: destinationHost}},
			}},
		}
	}

	out.Http = istioRoutes
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
		useRule, err := appliesToDestination(destinationHost, rr.DestinationSelector, upstreams)
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

func appliesToDestination(destinationHost string, destinationSelector *v1.PodSelector, upstreams gloov1.UpstreamList) (bool, error) {
	if destinationSelector == nil {
		return true, nil
	}
	switch selector := destinationSelector.SelectorType.(type) {
	case *v1.PodSelector_LabelSelector_:
		// true if an upstream exists whose selector falls within the rr's selector
		// and the host in question is that upstream's host
		for _, us := range upstreams {
			hostForUpstream, err := utils.GetHostForUpstream(us)
			if err != nil {
				return false, errors.Wrapf(err, "getting host for upstream")
			}
			// we only care about the host in question
			if destinationHost != hostForUpstream {
				continue
			}

			upstreamLabels := utils.GetLabelsForUpstream(us)
			labelsMatch := labels.SelectorFromSet(selector.LabelSelector.LabelsToMatch).Matches(labels.Set(upstreamLabels))
			if !labelsMatch {
				continue
			}

			// we found an upstream with the correct host and labels
			return true, nil
		}
	case *v1.PodSelector_UpstreamSelector_:
		for _, ref := range selector.UpstreamSelector.Upstreams {
			us, err := upstreams.Find(ref.Strings())
			if err != nil {
				return false, err
			}
			hostForUpstream, err := utils.GetHostForUpstream(us)
			if err != nil {
				return false, errors.Wrapf(err, "getting host for upstream")
			}
			if hostForUpstream == destinationHost {
				return true, nil
			}
		}
	case *v1.PodSelector_NamespaceSelector_:
		for _, us := range upstreams {
			hostForUpstream, err := utils.GetHostForUpstream(us)
			if err != nil {
				return false, errors.Wrapf(err, "getting host for upstream")
			}
			// we only care about the host in question
			if destinationHost != hostForUpstream {
				continue
			}

			var usInSelectedNamespace bool
			for _, ns := range selector.NamespaceSelector.Namespaces {
				namespaceForUpstream := utils.GetNamespaceForUpstream(us)
				if ns == namespaceForUpstream {
					usInSelectedNamespace = true
					break
				}
			}
			if !usInSelectedNamespace {
				continue
			}

			// we found an upstream with the correct host and namespace
			return true, nil

		}
	}
	return false, nil
}
