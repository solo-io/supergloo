package istio

import (
	"context"
	"fmt"
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/solo-kit/pkg/utils/contextutils"
	gloov1 "github.com/solo-io/supergloo/pkg/api/external/gloo/v1"
	"github.com/solo-io/supergloo/pkg/api/external/istio/networking/v1alpha3"
	"github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/defaults"
)

type MeshRoutingSyncer struct {
	WriteSelector             map[string]string // for reconciling only our resources
	WriteNamespace            string
	DestinationRuleReconciler v1alpha3.DestinationRuleReconciler
	VirtualServiceReconciler  v1alpha3.VirtualServiceReconciler
	Reporter                  reporter.Reporter
}

func updateMetadataForWriting(meta *core.Metadata, writeNamespace string, writeSelector map[string]string) {
	meta.Namespace = writeNamespace
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
	for _, res := range destinationRules {
		updateMetadataForWriting(&res.Metadata, s.WriteNamespace, s.WriteSelector)
	}
	for _, res := range virtualServices {
		updateMetadataForWriting(&res.Metadata, s.WriteNamespace, s.WriteSelector)
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

func subsetName(host string, labels map[string]string) string {
	return fmt.Sprintf("%v-%+v", host, labels)
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
			labels := getLabelsForUpstream(us)
			host, err := getHostForUpstream(us)
			if err != nil {
				return nil, errors.Wrapf(err, "getting host for upstream")
			}
			labelsByHost[host] = append(labelsByHost[host], labels)
		}
		for host, labelSets := range labelsByHost {
			var subsets []*v1alpha3.Subset
			for _, labels := range labelSets {
				subsets = append(subsets, &v1alpha3.Subset{
					Name:   subsetName(host, labels),
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
	var virtualServices v1alpha3.VirtualServiceList
	for _, rule := range rules {
		vs, err := virtualServicesForRule(rule, meshes, upstreams)
		if err != nil {
			return nil, errors.Wrapf(err, "creating virtual service for rule %v", rule)
		}
		virtualServices = append(virtualServices, vs...)
	}
	return virtualServices, nil
}

func virtualServicesForRule(rule *v1.RoutingRule, meshes v1.MeshList, upstreams gloov1.UpstreamList) (v1alpha3.VirtualServiceList, error) {
	istioMesh, err := getIstioMeshForRule(rule, meshes)
	if err != nil {
		return nil, err
	}
	// we can only write our crds to a namespace istio watches
	// just pick the first one for now
	// if empty, it defaults to supergloo-system & default
	validNamespaces := istioMesh.WatchNamespaces
	if len(validNamespaces) == 0 {
		validNamespaces = []string{defaults.Namespace, "default"}
	}
	var found bool
	for _, ns := range validNamespaces {
		if ns == rule.Metadata.Namespace {
			found = true
			break
		}
	}
	if !found {
		return nil, errors.Errorf("routing rule %v is not in a namespace that belongs to target mesh",
			rule.Metadata.Ref())
	}

	// matcher is the same regardless of destination
	istioMatcher, err := createIstioMatcher(rule, upstreams)
	if err != nil {
		return nil, errors.Wrapf(err, "creating istio matcher")
	}
	var destinationUpstreams gloov1.UpstreamList
	if len(rule.Destinations) == 0 {
		destinationUpstreams = upstreams
	} else {
		for _, dest := range rule.Destinations {
			ups, err := upstreams.Find(dest.Strings())
			if err != nil {
				return nil, errors.Wrapf(err, "invalid destination for rule %v", dest)
			}
			destinationUpstreams = append(destinationUpstreams, ups)
		}
	}

	var virtualServices v1alpha3.VirtualServiceList
	for _, us := range destinationUpstreams {
		labels := getLabelsForUpstream(us)
		host, err := getHostForUpstream(us)
		if err != nil {
			return nil, errors.Wrapf(err, "getting host for upstream")
		}

		destinations, err := createIstioDestinations(host, labels, rule, upstreams)
		if err != nil {
			return nil, errors.Wrapf(err, "creating istio destinations")
		}
		vs := &v1alpha3.VirtualService{
			Metadata: core.Metadata{
				Name:      us.Metadata.Name + "-" + rule.Metadata.Name,
				Namespace: rule.Metadata.Namespace,
			},
			Hosts: []string{host},
			// in istio api, this is equivalent to []string{"mesh"}
			// which includes all pods in the mesh, with no selectors
			// and no ingresses
			Gateways: []string{"mesh"},
			Http: []*v1alpha3.HTTPRoute{{
				Match: istioMatcher,
				Route: destinations,
			}},
		}
		if err := addHttpFeatures(rule, vs, upstreams); err != nil {
			return nil, errors.Wrapf(err, "failed to add http features to virtual service")
		}
		virtualServices = append(virtualServices, vs)
	}

	return virtualServices, nil
}

func createIstioMatcher(rule *v1.RoutingRule, upstreams gloov1.UpstreamList) ([]*v1alpha3.HTTPMatchRequest, error) {
	var sourceLabelSets []map[string]string
	for _, src := range rule.Sources {
		upstream, err := upstreams.Find(src.Strings())
		if err != nil {
			return nil, errors.Wrapf(err, "invalid source %v", src)
		}
		labels := getLabelsForUpstream(upstream)
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

func createIstioDestinations(orinalHost string, originalLabels map[string]string, rule *v1.RoutingRule, upstreams gloov1.UpstreamList) ([]*v1alpha3.HTTPRouteDestination, error) {
	if rule.TrafficShifting == nil || len(rule.TrafficShifting.Destinations) == 0 {
		return []*v1alpha3.HTTPRouteDestination{{
			Destination: &v1alpha3.Destination{
				Host:   orinalHost,
				Subset: subsetName(orinalHost, originalLabels),
			},
		}}, nil
	}
	var istioDestinations []*v1alpha3.HTTPRouteDestination
	for _, dest := range rule.TrafficShifting.Destinations {
		upstream, err := upstreams.Find(dest.Upstream.Strings())
		if err != nil {
			return nil, errors.Wrapf(err, "invalid destination %v", dest)
		}
		host, err := getHostForUpstream(upstream)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get host for upstream")
		}
		labels := getLabelsForUpstream(upstream)
		var port *v1alpha3.PortSelector
		intPort, err := getPortForUpstream(upstream)
		if err != nil {
			return nil, errors.Wrapf(err, "getting port for upstream")
		}
		if intPort > 0 {
			port = &v1alpha3.PortSelector{
				Port: &v1alpha3.PortSelector_Number{Number: intPort},
			}
		}
		istioDestinations = append(istioDestinations, &v1alpha3.HTTPRouteDestination{
			Destination: &v1alpha3.Destination{
				Host:   host,
				Subset: subsetName(host, labels),
				Port:   port,
			},
		})
	}
	return istioDestinations, nil
}

func addHttpFeatures(rule *v1.RoutingRule, virtualService *v1alpha3.VirtualService, upstreams gloov1.UpstreamList) error {
	for _, http := range virtualService.Http {
		http.CorsPolicy = rule.CorsPolicy
		http.Retries = rule.Retries
		http.Timeout = rule.Timeout
		if rule.Mirror != nil {
			us, err := upstreams.Find(rule.Mirror.Upstream.Strings())
			labels := getLabelsForUpstream(us)
			host, err := getHostForUpstream(us)
			if err != nil {
				return errors.Wrapf(err, "getting host for upstream")
			}
			http.Mirror = &v1alpha3.Destination{
				Host:   host,
				Subset: subsetName(host, labels),
			}
		}
		if rule.HeaderManipulaition != nil {
			http.RemoveRequestHeaders = rule.HeaderManipulaition.RemoveRequestHeaders
			http.AppendRequestHeaders = rule.HeaderManipulaition.AppendRequestHeaders
			http.RemoveResponseHeaders = rule.HeaderManipulaition.RemoveResponseHeaders
			http.AppendResponseHeaders = rule.HeaderManipulaition.AppendResponseHeaders
		}
	}
	return nil
}

// util functions
func (s *MeshRoutingSyncer) writeIstioCrds(ctx context.Context, destinationRules v1alpha3.DestinationRuleList, virtualServices v1alpha3.VirtualServiceList) error {
	opts := clients.ListOpts{
		Ctx:      ctx,
		Selector: s.WriteSelector,
	}
	contextutils.LoggerFrom(ctx).Infof("reconciling %v destination rules", len(destinationRules))
	if err := s.DestinationRuleReconciler.Reconcile(s.WriteNamespace, destinationRules, preserveDestinationRule, opts); err != nil {
		return errors.Wrapf(err, "reconciling destination rules")
	}
	contextutils.LoggerFrom(ctx).Infof("reconciling %v virtual services", len(virtualServices))
	if err := s.VirtualServiceReconciler.Reconcile(s.WriteNamespace, virtualServices, preserveVirtualService, opts); err != nil {
		return errors.Wrapf(err, "reconciling virtual services")
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

func getLabelsForUpstream(us *gloov1.Upstream) map[string]string {
	switch specType := us.UpstreamSpec.UpstreamType.(type) {
	case *gloov1.UpstreamSpec_Kube:
		return specType.Kube.Selector
	}
	// default to using the labels from the upstream
	return us.Metadata.Labels
}

func getHostForUpstream(us *gloov1.Upstream) (string, error) {
	hosts, err := getHostsForUpstream(us)
	if err != nil {
		return "", errors.Wrapf(err, "failed to get hosts for upstream")
	}
	if len(hosts) < 1 {
		return "", errors.Errorf("failed to get hosts for upstream")
	}
	return hosts[0], nil
}

func getHostsForUpstream(us *gloov1.Upstream) ([]string, error) {
	switch specType := us.UpstreamSpec.UpstreamType.(type) {
	case *gloov1.UpstreamSpec_Aws:
		return nil, errors.Errorf("aws not implemented")
	case *gloov1.UpstreamSpec_Azure:
		return nil, errors.Errorf("azure not implemented")
	case *gloov1.UpstreamSpec_Kube:
		return []string{
			fmt.Sprintf("%v.%v.svc.cluster.local", specType.Kube.ServiceName, specType.Kube.ServiceNamespace),
			specType.Kube.ServiceName,
		}, nil
	case *gloov1.UpstreamSpec_Static:
		var hosts []string
		for _, h := range specType.Static.Hosts {
			hosts = append(hosts, h.Addr)
		}
		return hosts, nil
	}
	return nil, errors.Errorf("unsupported upstream type %v", us)
}

func getPortForUpstream(us *gloov1.Upstream) (uint32, error) {
	switch specType := us.UpstreamSpec.UpstreamType.(type) {
	case *gloov1.UpstreamSpec_Aws:
		return 0, errors.Errorf("aws not implemented")
	case *gloov1.UpstreamSpec_Azure:
		return 0, errors.Errorf("azure not implemented")
	case *gloov1.UpstreamSpec_Kube:
		return specType.Kube.ServicePort, nil
	case *gloov1.UpstreamSpec_Static:
		// TODO(ilackarms): handle cases where port changes between hosts
		for _, h := range specType.Static.Hosts {
			return h.Port, nil
		}
		return 0, errors.Errorf("no hosts found on static upstream")
	}
	return 0, errors.Errorf("unknown upstream type")
}
