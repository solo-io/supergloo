package istio

import (
	"context"
	"strings"

	"github.com/solo-io/supergloo/pkg/translator/utils"

	"github.com/mitchellh/hashstructure"
	"github.com/solo-io/solo-kit/pkg/utils/nameutils"
	"github.com/solo-io/solo-kit/pkg/utils/stringutils"
	"go.uber.org/multierr"

	"github.com/gogo/protobuf/proto"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/solo-kit/pkg/utils/contextutils"
	gloov1 "github.com/solo-io/supergloo/pkg/api/external/gloo/v1"
	"github.com/solo-io/supergloo/pkg/api/external/istio/networking/v1alpha3"
	"github.com/solo-io/supergloo/pkg/api/v1"
)

type MeshRoutingSyncer struct {
	// only read/write crds from these namespaces
	// needed so we can clean up hanging crds
	writeNamespaces []string
	// for reconciling only our resources
	// override write selector to prevent conflicts between multiple routing syncers
	// else leave nil
	writeSelector             map[string]string
	destinationRuleReconciler v1alpha3.DestinationRuleReconciler
	virtualServiceReconciler  v1alpha3.VirtualServiceReconciler
	// ilackarms: todo: implement reporter
	reporter reporter.Reporter
}

func NewMeshRoutingSyncer(writeNamespaces []string,
	writeSelector map[string]string, // for reconciling only our resources
	destinationRuleReconciler v1alpha3.DestinationRuleReconciler,
	virtualServiceReconciler v1alpha3.VirtualServiceReconciler,
	reporter reporter.Reporter) *MeshRoutingSyncer {
	if writeSelector == nil {
		writeSelector = map[string]string{"reconciler.solo.io": "supergloo.istio.routing"}
	}
	return &MeshRoutingSyncer{
		writeNamespaces:           writeNamespaces,
		writeSelector:             writeSelector,
		destinationRuleReconciler: destinationRuleReconciler,
		virtualServiceReconciler:  virtualServiceReconciler,
		reporter:                  reporter,
	}
}

func sanitizeName(name string) string {
	name = strings.Replace(name, ".", "-", -1)
	name = strings.Replace(name, "[", "", -1)
	name = strings.Replace(name, "]", "", -1)
	name = strings.Replace(name, ":", "-", -1)
	name = strings.Replace(name, " ", "-", -1)
	name = strings.Replace(name, "\n", "", -1)
	return name
}

func updateMetadataForWriting(meta *core.Metadata, writeSelector map[string]string) {
	meta.Name = nameutils.SanitizeName(sanitizeName(meta.Name))
	if meta.Annotations == nil {
		meta.Annotations = make(map[string]string)
	}
	meta.Annotations["created_by"] = "supergloo"
	if meta.Labels == nil && len(writeSelector) > 0 {
		meta.Labels = make(map[string]string)
	}
	for k, v := range writeSelector {
		meta.Labels[k] = v
	}
}

func (s *MeshRoutingSyncer) Sync(ctx context.Context, snap *v1.TranslatorSnapshot) error {
	ctx = contextutils.WithLogger(ctx, "mesh-routing-syncer")
	logger := contextutils.LoggerFrom(ctx)
	meshes := snap.Meshes.List()
	upstreams := snap.Upstreams.List()
	rules := snap.Routingrules.List()

	logger.Infof("begin sync %v (%v meshes, %v upstreams, %v rules)", snap.Hash(),
		len(meshes), len(upstreams), len(rules))
	defer logger.Infof("end sync %v", snap.Hash())
	logger.Debugf("%v", snap)

	destinationRules, err := destinationRulesForUpstreams(rules, meshes, upstreams)
	if err != nil {
		return errors.Wrapf(err, "creating subsets from snapshot")
	}

	virtualServices, err := virtualServicesForRules(rules, meshes, upstreams)
	if err != nil {
		return errors.Wrapf(err, "creating virtual services from snapshot")
	}

	// nothing to do until we create some
	if len(virtualServices) == 0 {
		return nil
	}

	for _, res := range destinationRules {
		updateMetadataForWriting(&res.Metadata, s.writeSelector)
	}
	for _, res := range virtualServices {
		updateMetadataForWriting(&res.Metadata, s.writeSelector)
	}
	return s.writeIstioCrds(ctx, destinationRules, virtualServices)
}

func getIstioMeshForRule(rule *v1.RoutingRule, meshes v1.MeshList) (*v1.Istio, error) {
	if rule.TargetMesh == nil {
		return nil, errors.Errorf("target_mesh required")
	}
	mesh, err := meshes.Find(rule.TargetMesh.Namespace, rule.TargetMesh.Name)
	if err != nil {
		return nil, errors.Wrapf(err, "finding target mesh %v", rule.TargetMesh)
	}
	istioMesh, ok := mesh.MeshType.(*v1.Mesh_Istio)
	if !ok {
		// not our mesh, we don't care
		return nil, nil
	}
	if istioMesh.Istio == nil {
		return nil, errors.Errorf("target istio mesh is invalid")
	}
	return istioMesh.Istio, nil
}

func subsetName(labels map[string]string) string {
	keys, values := stringutils.KeysAndValues(labels)
	name := ""
	for i := range keys {
		name += keys[i] + "-" + values[i] + "-"
	}
	name = strings.TrimSuffix(name, "-")
	return sanitizeName(name)
}

// destinationrules
func destinationRulesForUpstreams(rules v1.RoutingRuleList, meshes v1.MeshList, upstreams gloov1.UpstreamList) (v1alpha3.DestinationRuleList, error) {
	var meshesWithRouteRules v1.MeshList
	for _, rule := range rules {
		mesh, err := meshes.Find(rule.TargetMesh.Namespace, rule.TargetMesh.Name)
		if err != nil {
			return nil, errors.Wrapf(err, "finding target mesh %v", rule.TargetMesh)
		}
		if _, err := getIstioMeshForRule(rule, meshes); err != nil {
			return nil, err
		}
		var found bool
		for _, addedMesh := range meshesWithRouteRules {
			if mesh == addedMesh {
				found = true
				break
			}
		}
		if !found {
			meshesWithRouteRules = append(meshesWithRouteRules, mesh)
		}
	}
	if len(meshesWithRouteRules) == 0 {
		return nil, nil
	}

	var destinationRules v1alpha3.DestinationRuleList
	for _, mesh := range meshesWithRouteRules {
		mtlsEnabled := mesh.Encryption != nil && mesh.Encryption.TlsEnabled
		labelsByHost := make(map[string][]map[string]string)
		for _, us := range upstreams {
			labels := utils.GetLabelsForUpstream(us)
			host, err := utils.GetHostForUpstream(us)
			if err != nil {
				return nil, errors.Wrapf(err, "getting host for upstream")
			}
			labelsByHost[host] = append(labelsByHost[host], labels)
		}
		for host, labelSets := range labelsByHost {
			var subsets []*v1alpha3.Subset
			for _, labels := range labelSets {
				subsets = append(subsets, &v1alpha3.Subset{
					Name:   subsetName(labels),
					Labels: labels,
				})
			}
			var trafficPolicy *v1alpha3.TrafficPolicy
			if mtlsEnabled {
				trafficPolicy = &v1alpha3.TrafficPolicy{
					Tls: &v1alpha3.TLSSettings{
						Mode: v1alpha3.TLSSettings_ISTIO_MUTUAL,
					},
				}
			}
			destinationRules = append(destinationRules, &v1alpha3.DestinationRule{
				Metadata: core.Metadata{
					Namespace: mesh.Metadata.Namespace,
					Name:      mesh.Metadata.Name + "-" + host,
				},
				Host:          host,
				TrafficPolicy: trafficPolicy,
				Subsets:       subsets,
			})
		}
	}

	return destinationRules.Sort(), nil
}

// virtualservices
func virtualServicesForRules(rules v1.RoutingRuleList, meshes v1.MeshList, upstreams gloov1.UpstreamList) (v1alpha3.VirtualServiceList, error) {
	// separate config changes for each mesh
	meshesByRule := make(map[*v1.RoutingRule]v1.MeshList)
	for _, rule := range rules {
		istioMesh, err := getIstioMeshForRule(rule, meshes)
		if err == nil && istioMesh == nil {
			// not our mesh, ignore rule
			continue
		}

		if err := validateRule(rule, meshes); err != nil {
			return nil, err
		}
		mesh, err := meshes.Find(rule.TargetMesh.Strings())
		if err != nil {
			// should never happen, error is already caught
			return nil, errors.Wrapf(err, "mesh find fail: should have already been caught")
		}
		meshesByRule[rule] = append(meshesByRule[rule], mesh)
	}

	// 1 virtualservice per unique host
	var hosts []string
	for _, rule := range rules {
		destUpstreams, err := upstreamsForRule(rule, upstreams)
		if err != nil {
			return nil, err
		}
	addUniqueHosts:
		for _, us := range destUpstreams {
			host, err := utils.GetHostForUpstream(us)
			if err != nil {
				return nil, errors.Wrapf(err, "getting host for upstream")
			}
			for _, added := range hosts {
				if host == added {
					continue addUniqueHosts
				}
			}
			hosts = append(hosts, host)
		}
	}

	// for each host
	// for each mesh
	// create one virtualservice for that host-mesh's set of rules
	rulesByMeshByHost := make(map[string]map[*v1.Mesh]v1.RoutingRuleList)
	for rule, meshList := range meshesByRule {
		for _, host := range hosts {
			if rulesByMeshByHost[host] == nil {
				rulesByMeshByHost[host] = make(map[*v1.Mesh]v1.RoutingRuleList)
			}
			// copy the rule for each mesh
			for _, mesh := range meshList {
				rulesByMeshByHost[host][mesh] = append(rulesByMeshByHost[host][mesh], rule)
			}
		}
	}

	var virtualServices v1alpha3.VirtualServiceList
	for host, rulesByMesh := range rulesByMeshByHost {
		for mesh, rules := range rulesByMesh {
			vs, err := virtualServiceForHost(host, rules, mesh, upstreams)
			if err != nil {
				return nil, errors.Wrapf(err, "creating virtual service for rules %v", rules)
			}
			virtualServices = append(virtualServices, vs)
		}
	}
	return virtualServices, nil
}

func validateRule(rule *v1.RoutingRule, meshes v1.MeshList) error {
	istioMesh, err := getIstioMeshForRule(rule, meshes)
	if err != nil {
		return err
	}
	// we can only write our crds to a namespace istio watches
	// just pick the first one for now
	// if empty, all namespaces are valid
	validNamespaces := istioMesh.WatchNamespaces
	if len(validNamespaces) == 0 {
		return nil
	}
	var found bool
	for _, ns := range validNamespaces {
		if ns == rule.Metadata.Namespace {
			found = true
			break
		}
	}
	if !found {
		return errors.Errorf("routing rule %v is not in a namespace that belongs to target mesh",
			rule.Metadata.Ref())
	}
	return nil
}

func upstreamsForRule(rule *v1.RoutingRule, upstreams gloov1.UpstreamList) (gloov1.UpstreamList, error) {
	var destinationUpstreams gloov1.UpstreamList
	if len(rule.Destinations) == 0 {
		return upstreams, nil
	}
	for _, dest := range rule.Destinations {
		ups, err := upstreams.Find(dest.Strings())
		if err != nil {
			return nil, errors.Wrapf(err, "invalid destination for rule %v", dest)
		}
		destinationUpstreams = append(destinationUpstreams, ups)
	}
	return destinationUpstreams, nil
}

func mergeRulesByMatch(rules v1.RoutingRuleList) (v1.RoutingRuleList, error) {
	rulesByUniqueMatch := make(map[uint64]v1.RoutingRuleList)
	for _, rule := range rules {
		type uniqueMatch struct {
			Sources         []*core.ResourceRef
			Destinations    []*core.ResourceRef
			RequestMatchers []*gloov1.Matcher
		}
		hash, _ := hashstructure.Hash(uniqueMatch{
			Sources:         rule.Sources,
			Destinations:    rule.Destinations,
			RequestMatchers: rule.RequestMatchers,
		}, nil)
		rulesByUniqueMatch[hash] = append(rulesByUniqueMatch[hash], rule)
	}
	var mergedRules v1.RoutingRuleList
	for _, rulesForMatch := range rulesByUniqueMatch {
		mergedRule := &v1.RoutingRule{
			Sources:         rulesForMatch[0].Sources,
			Destinations:    rulesForMatch[0].Destinations,
			RequestMatchers: rulesForMatch[0].RequestMatchers,
		}
		for _, rule := range rulesForMatch {
			if rule.TrafficShifting != nil {
				if mergedRule.TrafficShifting != nil {
					return nil, errors.Errorf("TrafficShifting redefined in for match %v", rule.Metadata.Ref())
				}
				mergedRule.TrafficShifting = rule.TrafficShifting
			}
			if rule.FaultInjection != nil {
				if mergedRule.FaultInjection != nil {
					return nil, errors.Errorf("FaultInjection redefined in for match %v", rule.Metadata.Ref())
				}
				mergedRule.FaultInjection = rule.FaultInjection
			}
			if rule.Timeout != nil {
				if mergedRule.Timeout != nil {
					return nil, errors.Errorf("Timeout redefined in for match %v", rule.Metadata.Ref())
				}
				mergedRule.Timeout = rule.Timeout
			}
			if rule.Retries != nil {
				if mergedRule.Retries != nil {
					return nil, errors.Errorf("Retries redefined in for match %v", rule.Metadata.Ref())
				}
				mergedRule.Retries = rule.Retries
			}
			if rule.CorsPolicy != nil {
				if mergedRule.CorsPolicy != nil {
					return nil, errors.Errorf("CorsPolicy redefined in for match %v", rule.Metadata.Ref())
				}
				mergedRule.CorsPolicy = rule.CorsPolicy
			}
			if rule.Mirror != nil {
				if mergedRule.Mirror != nil {
					return nil, errors.Errorf("Mirror redefined in for match %v", rule.Metadata.Ref())
				}
				mergedRule.Mirror = rule.Mirror
			}
			if rule.HeaderManipulaition != nil {
				if mergedRule.HeaderManipulaition != nil {
					return nil, errors.Errorf("HeaderManipulaition redefined in for match %v", rule.Metadata.Ref())
				}
				mergedRule.HeaderManipulaition = rule.HeaderManipulaition
			}
		}
		mergedRules = append(mergedRules, mergedRule)
	}
	return mergedRules.Sort(), nil
}

func virtualServiceForHost(host string, rules v1.RoutingRuleList, mesh *v1.Mesh, upstreams gloov1.UpstreamList) (*v1alpha3.VirtualService, error) {
	rules, err := mergeRulesByMatch(rules)
	if err != nil {
		return nil, errors.Wrapf(err, "route rules incompatible")
	}

	var istioRules []*v1alpha3.HTTPRoute
	for _, rule := range rules {
		// each rule gets its own HTTPRoute

		// matcher is the same regardless of destination
		// upstreams are used for our SOURCES here
		// this requires upstreams to be created for our source pods
		match, err := createIstioMatcher(rule, upstreams)
		if err != nil {
			return nil, errors.Wrapf(err, "creating istio matcher")
		}

		// default: single destination, original
		route := []*v1alpha3.DestinationWeight{{
			Destination: &v1alpha3.Destination{
				Host: host,
			},
		}}
		if rule.TrafficShifting != nil && len(rule.TrafficShifting.Destinations) > 0 {
			route, err = createLoadBalancedRoute(rule.TrafficShifting.Destinations, upstreams)
			if err != nil {
				return nil, errors.Wrapf(err, "creating multi destination route")
			}
		}
		httpRoute := &v1alpha3.HTTPRoute{
			Match: match,
			Route: route,
		}
		if err := addHttpFeatures(rule, httpRoute, upstreams); err != nil {
			return nil, errors.Wrapf(err, "adding http features to route")
		}

		istioRules = append(istioRules, httpRoute)
	}

	return &v1alpha3.VirtualService{
		Metadata: core.Metadata{
			Name:      mesh.Metadata.Name + "-" + host,
			Namespace: mesh.Metadata.Namespace,
		},
		Hosts: []string{host},
		// in istio api, this is equivalent to []string{"mesh"}
		// which includes all pods in the mesh, with no selectors
		// and no ingresses
		Gateways: []string{"mesh"},
		Http:     istioRules,
		//[]*v1alpha3.HTTPRoute{{
		//	Match: istioMatchers,
		//	Route: istioDestinations,
		//}},
	}, nil
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

func createLoadBalancedRoute(destinations []*v1.WeightedDestination, upstreams gloov1.UpstreamList) ([]*v1alpha3.DestinationWeight, error) {
	var istioDestinations []*v1alpha3.DestinationWeight
	for _, dest := range destinations {
		upstream, err := upstreams.Find(dest.Upstream.Strings())
		if err != nil {
			return nil, errors.Wrapf(err, "invalid destination %v", dest)
		}
		host, err := utils.GetHostForUpstream(upstream)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get host for upstream")
		}
		labels := utils.GetLabelsForUpstream(upstream)
		var port *v1alpha3.PortSelector
		intPort, err := utils.GetPortForUpstream(upstream)
		if err != nil {
			return nil, errors.Wrapf(err, "getting port for upstream")
		}
		if intPort > 0 {
			port = &v1alpha3.PortSelector{
				Port: &v1alpha3.PortSelector_Number{Number: intPort},
			}
		}
		istioDestinations = append(istioDestinations, &v1alpha3.DestinationWeight{
			Destination: &v1alpha3.Destination{
				Host:   host,
				Subset: subsetName(labels),
				Port:   port,
			},
		})
	}
	return istioDestinations, nil
}

func addHttpFeatures(rule *v1.RoutingRule, http *v1alpha3.HTTPRoute, upstreams gloov1.UpstreamList) error {
	http.Fault = rule.FaultInjection
	http.CorsPolicy = rule.CorsPolicy
	http.Retries = rule.Retries
	http.Timeout = rule.Timeout
	if rule.Mirror != nil {
		us, err := upstreams.Find(rule.Mirror.Upstream.Strings())
		labels := utils.GetLabelsForUpstream(us)
		host, err := utils.GetHostForUpstream(us)
		if err != nil {
			return errors.Wrapf(err, "getting host for upstream")
		}
		http.Mirror = &v1alpha3.Destination{
			Host:   host,
			Subset: subsetName(labels),
		}
	}
	if rule.HeaderManipulaition != nil {
		//http.RemoveRequestHeaders = rule.HeaderManipulaition.RemoveRequestHeaders
		http.AppendHeaders = rule.HeaderManipulaition.AppendRequestHeaders
		http.RemoveResponseHeaders = rule.HeaderManipulaition.RemoveResponseHeaders
		//http.AppendResponseHeaders = rule.HeaderManipulaition.AppendResponseHeaders
	}
	return nil
}

// util functions
func (s *MeshRoutingSyncer) writeIstioCrds(ctx context.Context, destinationRules v1alpha3.DestinationRuleList, virtualServices v1alpha3.VirtualServiceList) error {
	opts := clients.ListOpts{
		Ctx:      ctx,
		Selector: s.writeSelector,
	}
	contextutils.LoggerFrom(ctx).Infof("reconciling %v destination rules", len(destinationRules))
	drByNamespace := make(v1alpha3.DestinationrulesByNamespace)
	drByNamespace.Add(destinationRules...)
	for _, ns := range s.writeNamespaces {
		if err := s.destinationRuleReconciler.Reconcile(ns, drByNamespace[ns], preserveDestinationRule, opts); err != nil {
			return errors.Wrapf(err, "reconciling destination rules")
		}
		delete(drByNamespace, ns)
	}
	if len(drByNamespace) != 0 {
		var err error
		for ns, destinationRules := range drByNamespace {
			err = multierr.Append(err, errors.Errorf("internal error: "+
				"%v destination rules were generated for namespace %v, but only "+
				"%v are available namespaces for writing", len(destinationRules), ns, s.writeNamespaces))
		}
	}

	contextutils.LoggerFrom(ctx).Infof("reconciling %v virtual services", len(virtualServices))
	vsByNamespace := make(v1alpha3.VirtualservicesByNamespace)
	vsByNamespace.Add(virtualServices...)
	for _, ns := range s.writeNamespaces {
		if err := s.virtualServiceReconciler.Reconcile(ns, vsByNamespace[ns], preserveVirtualService, opts); err != nil {
			return errors.Wrapf(err, "reconciling virtual services")
		}
		delete(drByNamespace, ns)
	}
	if len(vsByNamespace) != 0 {
		var err error
		for ns, virtualServices := range vsByNamespace {
			err = multierr.Append(err, errors.Errorf("internal error: "+
				"%v virtual services were generated for namespace %v, but only "+
				"%v are available namespaces for writing", len(virtualServices), ns, s.writeNamespaces))
		}
	}
	return nil
}

func preserveDestinationRule(original, desired *v1alpha3.DestinationRule) (bool, error) {
	original.Metadata = desired.Metadata
	original.Status = desired.Status
	return !proto.Equal(original, desired), nil
}

func preserveVirtualService(original, desired *v1alpha3.VirtualService) (bool, error) {
	original.Metadata = desired.Metadata
	original.Status = desired.Status
	return !proto.Equal(original, desired), nil
}
