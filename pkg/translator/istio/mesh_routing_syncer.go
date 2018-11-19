package istio

import (
	"context"

	"github.com/solo-io/envoy-gloo/bazel-envoy-gloo/external/go_sdk/src/strings"
	"github.com/solo-io/supergloo/pkg/defaults"

	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"

	"github.com/solo-io/solo-kit/pkg/utils/contextutils"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/errors"
	gloov1 "github.com/solo-io/supergloo/pkg/api/external/gloo/v1"
	"github.com/solo-io/supergloo/pkg/api/external/istio/networking/v1alpha3"
	"github.com/solo-io/supergloo/pkg/api/v1"
)

type MeshRoutingSyncer struct {
	WriteSelector             map[string]string // for reconciling only our resources
	WriteNamespace            string
	DestinationRuleReconciler v1alpha3.DestinationRuleReconciler
	VirtualServiceReconciler  v1alpha3.VirtualServiceReconciler
	Reporter                  reporter.Reporter
}

func getLabelsForUpstream(us *gloov1.Upstream) map[string]string {
	switch specType := us.UpstreamSpec.UpstreamType.(type) {
	case *gloov1.UpstreamSpec_Kube:
		return specType.Kube.Selector
	}
	// default to using the labels from the upstream
	return us.Metadata.Labels
}

func processRule(rule *v1.RoutingRule, meshes v1.MeshList, upstreams gloov1.UpstreamList) (*v1alpha3.VirtualService, error) {
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
	// we can only write our crds to a namespace istio watches
	// just pick the first one for now
	// if empty, it defaults to supergloo-system & default
	validNamespaces := istioMesh.Istio.WatchNamespaces
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
		return nil, errors.Errorf("routing rule %v is not in a namespace that belongs to target mesh %v",
			rule.Metadata.Ref(), mesh.Metadata.Ref())
	}

	istioMatcher, err := createIstioMatcher(rule, upstreams)
	if err != nil {
		return nil, errors.Wrapf(err, "creating istio matcher")
	}

	return &v1alpha3.VirtualService{
		Metadata: core.Metadata{
			Name:      "supergloo-" + rule.Metadata.Name,
			Namespace: rule.Metadata.Namespace,
		},
		Gateways: []string{"mesh"},
		Http: []*v1alpha3.HTTPRoute{{
			Match: istioMatcher,
			Route: istioRoute,
		}},
	}, nil
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

func createDestinations(rule *v1.RoutingRule, upstreams gloov1.UpstreamList) []*v1alpha3.HTTPRouteDestination {
	if len(rule.DestinationSelector) == 0 {
		return []*v1alpha3.HTTPRouteDestination{{}}
	}
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

// TODO: if mesh has tls enabled (istio), set tls config on destination rule to istio_mutual

func (s *MeshRoutingSyncer) Sync(ctx context.Context, snap *v1.TranslatorSnapshot) error {
	ctx = contextutils.WithLogger(ctx, "mesh-routing-syncer")
	logger := contextutils.LoggerFrom(ctx)
	logger.Infof("begin sync %v (%v meshes, %v upstreams)", snap.Hash(),
		len(snap.Meshes), len(snap.Upstreams))
	defer logger.Infof("end sync %v", snap.Hash())
	logger.Debugf("%v", snap)

	var virtualServices v1alpha3.VirtualServiceList

	reporterErrs := make(reporter.ResourceErrors)

	meshes := snap.Meshes.List()
	for _, rule := range snap.Routingrules.List() {
		vs, err := processRule(rule, meshes)
		if err != nil {
			logger.Warnf("error in rule %v: %v", rule.Metadata.Ref(), err)
			reporterErrs.AddError(rule, err)
		}
		vs := &v1alpha3.VirtualService{
			// in istio api, this is equivalent to []string{"mesh"}
			// which includes all pods in the mesh, with no selectors
			// and no ingresses
			Gateways: []string{},
			Hosts:    hosts,
			Http:     routes,
		}
		virtualServices = append(virtualServices, vs)
	}

	destinationRules := createDestinationRules(false, snap.Upstreams.List())
	virtualServices, err := createVirtualServices(snap.Meshes.List(), snap.Upstreams.List())
	if err != nil {
		return errors.Wrapf(err, "creating virtual services from snapshot")
	}
	for _, res := range destinationRules {
		resources.UpdateMetadata(res, func(meta *core.Metadata) {
			meta.Namespace = s.WriteNamespace
			if meta.Annotations == nil {
				meta.Annotations = make(map[string]string)
			}
			meta.Annotations["created_by"] = "supergloo"
			for k, v := range s.WriteSelector {
				meta.Labels[k] = v
			}
		})
	}
	for _, res := range virtualServices {
		resources.UpdateMetadata(res, func(meta *core.Metadata) {
			meta.Namespace = s.WriteNamespace
			if meta.Annotations == nil {
				meta.Annotations = make(map[string]string)
			}
			if meta.Labels == nil {
				meta.Labels = make(map[string]string)
			}
			meta.Annotations["created_by"] = "supergloo"
			for k, v := range s.WriteSelector {
				meta.Labels[k] = v
			}
		})
	}
	return s.writeIstioCrds(ctx, destinationRules, virtualServices)
}

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

func convertRout2e(originalDestination *gloov1.Destination, route []*v1.HTTPRouteDestination, upstreams gloov1.UpstreamList) ([]*v1alpha3.HTTPRouteDestination, error) {
	var istioMatch []*v1alpha3.HTTPRouteDestination
	for _, m := range route {
		destination := originalDestination
		if m.AlternateDestination != nil {
			destination = m.AlternateDestination
		}
		istioDestination, err := convertDestination(destination, upstreams)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to convert destination %v", destination)
		}
		istioMatch = append(istioMatch, &v1alpha3.HTTPRouteDestination{
			Destination:           istioDestination,
			Weight:                m.Weight,
			RemoveRequestHeaders:  m.RemoveRequestHeaders,
			RemoveResponseHeaders: m.RemoveResponseHeaders,
			AppendRequestHeaders:  m.AppendRequestHeaders,
			AppendResponseHeaders: m.AppendResponseHeaders,
		})
	}
	return istioMatch, nil
}
