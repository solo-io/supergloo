package istio

import (
	"sort"
	"strings"

	"github.com/pkg/errors"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/go-utils/hashutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/pkg/api/external/istio/networking/v1alpha3"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/translator/istio/plugins"
	"github.com/solo-io/supergloo/pkg/translator/utils"
	"k8s.io/apimachinery/pkg/labels"
)

// A container for the entire set of config for a single istio mesh
type MeshConfig struct {
	DesinationRules v1alpha3.DestinationRuleList
	VirtualServices v1alpha3.VirtualService
}

// todo: first create all desintation rules for all subsets of each upstream
// then we need to apply the MUTUAL or ISTIO_MUTUAL policy depending on
// whether mtls is enabled, and if so, if the user is using a selfsignedcert
// if MUTUAL, also need to provide the paths for the certs/keys
// i assume these are loaded to pilot somewhere from a secret

type translator struct {
	writeNamespace string
	plugins        []plugins.Plugin
}

// we create a routing rule for each unique matcher
type rulesByMatcher struct {
	rules map[uint64]matcherRules
}

func newRulesByMatcher() rulesByMatcher {
	return rulesByMatcher{rules: make(map[uint64]matcherRules)}
}

func (rbm rulesByMatcher) add(matcher []*gloov1.Matcher, rule *v1.RoutingRule) {
	hash := hashutils.HashAll(matcher)
	rulesForMatcher := rbm.rules[hash]
	rulesForMatcher.matcher = matcher
	rulesForMatcher.rules = append(rulesForMatcher.rules, rule)
}

func (rbm rulesByMatcher) sort() []matcherRules {
	var (
		rulesForMatcher []matcherRules
		hashes          []uint64
	)
	for hash, rules := range rbm.rules {
		hashes = append(hashes, hash)
		rulesForMatcher = append(rulesForMatcher, rules)
	}
	sort.SliceStable(rulesForMatcher, func(i, j int) bool {
		return hashes[i] < hashes[j]
	})
	return rulesForMatcher
}

type matcherRules struct {
	matcher []*gloov1.Matcher
	rules   v1.RoutingRuleList
}

// produces a complete istio config
func (t *translator) Translate(snapshot *v1.ConfigSnapshot) (*MeshConfig, reporter.ResourceErrors, error) {
	labelSetsByHost, err := labelsByHost(snapshot.Upstreams.List())
	if err != nil {
		return nil, nil, errors.Wrapf(err, "internal error: getting hosts from upstreams")
	}

	meshes := snapshot.Meshes.List()
	meshGroups := snapshot.Meshgroups.List()
	upstreams := snapshot.Upstreams.List()
	routingRules := snapshot.Routingrules.List()
	encryptionRules := snapshot.Ecryptionrules.List()

	resourceErrs := make(reporter.ResourceErrors)
	resourceErrs.Accept(meshes.AsInputResources()...)
	resourceErrs.Accept(meshGroups.AsInputResources()...)
	resourceErrs.Accept(routingRules.AsInputResources()...)
	resourceErrs.Accept(encryptionRules.AsInputResources()...)

	// group rules by their matcher
	// we will then create a corresponding http rule
	// on the virtual service that contains all the relevant rules
	rulesPerMatcher := newRulesByMatcher()
	for _, rr := range routingRules {
		rulesPerMatcher.add(rr.RequestMatchers, rr)
	}

	var destinationRules v1alpha3.DestinationRuleList
	var virtualServices v1alpha3.VirtualServiceList
	for host, labelSets := range labelSetsByHost {
		dr := initDestinationRule(t.writeNamespace, host, labelSets)
		vs := initVirtualService(t.writeNamespace, host)

		// find rules for this host and apply them
		for _, matcherAndRules := range rulesPerMatcher.sort() {

			// each unique matcher becomes an http rule in the virtual
			// service for this host

			matcher := matcherAndRules.matcher
			rules := matcherAndRules.rules

			create := func() *v1alpha3.HTTPRoute {
				return &v1alpha3.HTTPRoute{
					Match: []*v1alpha3.HTTPMatchRequest{},
				}
			}

			// create an http rule for the matcher

			// if rr does not apply to this host (destination), skip
			useRule, err := appliesToHost(host, rr, upstreams)
			if err != nil {
				return nil, nil, errors.Wrapf(err, "error: inspecting routing rule selector")
			}

			if !useRule {
				continue
			}

			matcher, err := createIstioMatcher(rr.RequestMatchers)
		}
	}

	// initialize destination rules
	destinationRules, err := destinationRulesFromUpstreams(t.writeNamespace, snapshot.Upstreams.List())
	if err != nil {
		return nil, nil, errors.Wrapf(err, "internal error: generating destination rules from upstreams")
	}
	// initialize virtual services
	virtualServices, err := virtualServicesFromUpstreams(t.writeNamespace, snapshot.Upstreams.List())
	if err != nil {
		return nil, nil, errors.Wrapf(err, "internal error: generating destination rules from upstreams")
	}

	return nil, nil, nil
}

func createIstioMatcher(rule *v1.RoutingRule, upstreams gloov1.UpstreamList) ([]*v1alpha3.HTTPMatchRequest, error) {
	var sourceLabelSets []map[string]string
	for _, src := range rule.Sources {
		upstream, err := upstreams.Find(src.Strings())
		if err != nil {
			return nil, errors.Wrapf(err, "invalid source %v", src)
		}
		labels := utils.GetLabelsForUpstream(upstream)
		sourceLabelSets = append(sourceLabelSets, labels)
	}

	var istioMatcher []*v1alpha3.HTTPMatchRequest

	// override for default istioMatcher
	requestMatchers := rule.RequestMatchers
	switch {
	case requestMatchers == nil && len(sourceLabelSets) == 0:

		// default, catch-all istioMatcher:
		istioMatcher = []*v1alpha3.HTTPMatchRequest{{
			Uri: &v1alpha3.StringMatch{
				MatchType: &v1alpha3.StringMatch_Prefix{
					Prefix: "/",
				},
			},
		}}
	case requestMatchers == nil && len(sourceLabelSets) > 0:
		istioMatcher = []*v1alpha3.HTTPMatchRequest{}
		for _, sourceLabels := range sourceLabelSets {
			istioMatcher = append(istioMatcher, convertMatcher(sourceLabels, &gloov1.Matcher{
				PathSpecifier: &gloov1.Matcher_Prefix{
					Prefix: "/",
				},
			}))
		}
	case requestMatchers != nil && len(sourceLabelSets) == 0:
		istioMatcher = []*v1alpha3.HTTPMatchRequest{}
		for _, match := range requestMatchers {
			istioMatcher = append(istioMatcher, convertMatcher(nil, match))
		}
	case requestMatchers != nil && len(sourceLabelSets) > 0:
		istioMatcher = []*v1alpha3.HTTPMatchRequest{}
		for _, match := range requestMatchers {
			for _, source := range sourceLabelSets {
				istioMatcher = append(istioMatcher, convertMatcher(source, match))
			}
		}
	}
	return istioMatcher, nil
}

func convertMatcher(sourceSelector map[string]string, match *gloov1.Matcher) *v1alpha3.HTTPMatchRequest {
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
	}
}

func appliesToHost(host string, rr *v1.RoutingRule, upstreams gloov1.UpstreamList) (bool, error) {
	if rr.DestinationSelector == nil {
		return true, nil
	}
	switch selector := rr.DestinationSelector.SelectorType.(type) {
	case *v1.PodSelector_LabelSelector_:
		// true if an upstream exists whose selector falls within the rr's selector
		// and the host in question is that upstream's host
		for _, us := range upstreams {
			hostForUpstream, err := utils.GetHostForUpstream(us)
			if err != nil {
				return false, errors.Wrapf(err, "getting host for upstream")
			}
			// we only care about the host in question
			if host != hostForUpstream {
				continue
			}

			upstreamLabels := utils.GetLabelsForUpstream(us)
			labelsMatch := labels.AreLabelsInWhiteList(upstreamLabels, selector.LabelSelector.LabelsToMatch)
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
			if hostForUpstream == host {
				return true, nil
			}
		}
	case *v1.PodSelector_NamespaceSelector_:
		//
		for _, us := range upstreams {
			hostForUpstream, err := utils.GetHostForUpstream(us)
			if err != nil {
				return false, errors.Wrapf(err, "getting host for upstream")
			}
			// we only care about the host in question
			if host != hostForUpstream {
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

func initDestinationRule(writeNamespace, host string, labelSets []map[string]string) *v1alpha3.DestinationRule {

	var subsets []*v1alpha3.Subset
	for _, labels := range labelSets {
		if len(labels) == 0 {
			continue
		}
		subsets = append(subsets, &v1alpha3.Subset{
			Name:   subsetName(labels),
			Labels: labels,
		})
	}
	return &v1alpha3.DestinationRule{
		Metadata: core.Metadata{
			Namespace: writeNamespace,
			Name:      host,
		},
		Host:    host,
		Subsets: subsets,
	}
}

func initVirtualService(writeNamespace, host string) *v1alpha3.VirtualService {
	return &v1alpha3.VirtualService{
		Metadata: core.Metadata{
			Namespace: writeNamespace,
			Name:      host,
		},
		Hosts:    []string{host},
		Gateways: []string{"mesh"},
	}
}
