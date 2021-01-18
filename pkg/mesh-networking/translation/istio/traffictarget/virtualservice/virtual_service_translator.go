package virtualservice

import (
	"context"
	"reflect"
	"sort"

	"github.com/golang/protobuf/proto"
	v1alpha3sets "github.com/solo-io/external-apis/pkg/api/istio/networking.istio.io/v1alpha3/sets"
	"github.com/solo-io/gloo-mesh/pkg/api/common.matchers.mesh.gloo.solo.io/v1alpha1"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/traffictarget/utils"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/selectorutils"
	skv2sets "github.com/solo-io/skv2/contrib/pkg/sets"
	"github.com/solo-io/skv2/pkg/ezkube"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/rotisserie/eris"
	discoveryv1alpha2 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1alpha2"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/fieldutils"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/hostutils"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/metautils"
	"github.com/solo-io/skv2/pkg/equalityutils"
	networkingv1alpha3spec "istio.io/api/networking/v1alpha3"
	networkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	"k8s.io/apimachinery/pkg/util/sets"
)

//go:generate mockgen -source ./virtual_service_translator.go -destination mocks/virtual_service_translator.go

// the VirtualService translator translates a TrafficTarget into a VirtualService.
type Translator interface {
	/*
		Translate translates the appropriate VirtualService for the given TrafficTarget.
		returns nil if no VirtualService is required for the TrafficTarget (i.e. if no VirtualService features are required, such as subsets).

		If sourceMeshInstallation is specified, hostnames in the translated VirtualService will use global FQDNs if the trafficTarget
		exists in a different cluster from the specified mesh (i.e. is a federated traffic target). Otherwise, assume translation
		for cluster that the trafficTarget exists in and use local FQDNs.

		Errors caused by invalid user config will be reported using the Reporter.

		Note that the input snapshot TrafficTargetSet contains the given TrafficTarget.
	*/
	Translate(
		ctx context.Context,
		in input.LocalSnapshot,
		trafficTarget *discoveryv1alpha2.TrafficTarget,
		sourceMeshInstallation *discoveryv1alpha2.MeshSpec_MeshInstallation,
		reporter reporting.Reporter,
	) *networkingv1alpha3.VirtualService
}

type translator struct {
	userVirtualServices v1alpha3sets.VirtualServiceSet
	clusterDomains      hostutils.ClusterDomainRegistry
	decoratorFactory    decorators.Factory
}

func NewTranslator(
	userVirtualServices v1alpha3sets.VirtualServiceSet,
	clusterDomains hostutils.ClusterDomainRegistry,
	decoratorFactory decorators.Factory,
) Translator {
	return &translator{
		userVirtualServices: userVirtualServices,
		clusterDomains:      clusterDomains,
		decoratorFactory:    decoratorFactory,
	}
}

// Translate a VirtualService for the TrafficTarget.
// If sourceMeshInstallation is nil, assume that VirtualService is colocated to the trafficTarget and use local FQDNs.
func (t *translator) Translate(
	_ context.Context,
	in input.LocalSnapshot,
	trafficTarget *discoveryv1alpha2.TrafficTarget,
	sourceMeshInstallation *discoveryv1alpha2.MeshSpec_MeshInstallation,
	reporter reporting.Reporter,
) *networkingv1alpha3.VirtualService {
	kubeService := trafficTarget.Spec.GetKubeService()

	if kubeService == nil {
		// TODO(ilackarms): non kube services currently unsupported
		return nil
	}

	sourceCluster := kubeService.Ref.ClusterName
	if sourceMeshInstallation != nil {
		sourceCluster = sourceMeshInstallation.Cluster
	}

	destinationFQDN := t.clusterDomains.GetDestinationFQDN(sourceCluster, trafficTarget.Spec.GetKubeService().Ref)

	virtualService := t.initializeVirtualService(trafficTarget, sourceMeshInstallation, destinationFQDN)
	// register the owners of the virtualservice fields
	virtualServiceFields := fieldutils.NewOwnershipRegistry()
	vsDecorators := t.decoratorFactory.MakeDecorators(decorators.Parameters{
		ClusterDomains: t.clusterDomains,
		Snapshot:       in,
	})

	appliedTpsByRequestMatcher := groupAppliedTpsByRequestMatcher(trafficTarget.Status.AppliedTrafficPolicies)

	for _, tpsByRequestMatcher := range appliedTpsByRequestMatcher {

		// initialize base route for TP's group by request matcher
		baseRoute := initializeBaseRoute(tpsByRequestMatcher[0].Spec, sourceCluster)

		for _, policy := range tpsByRequestMatcher {
			// nil baseRoute indicates that this cluster is not selected by the WorkloadSelector and thus should not be translated
			if baseRoute == nil {
				continue
			}

			registerField := registerFieldFunc(virtualServiceFields, virtualService, policy.Ref)
			for _, decorator := range vsDecorators {

				if trafficPolicyDecorator, ok := decorator.(decorators.TrafficPolicyVirtualServiceDecorator); ok {
					if err := trafficPolicyDecorator.ApplyTrafficPolicyToVirtualService(
						policy,
						trafficTarget,
						sourceMeshInstallation,
						baseRoute,
						registerField,
					); err != nil {
						reporter.ReportTrafficPolicyToTrafficTarget(trafficTarget, policy.Ref, eris.Wrapf(err, "%v", decorator.DecoratorName()))
					}
				}
			}
		}

		// Avoid appending an HttpRoute that will have no affect, which occurs if no decorators mutate the baseRoute with any non-match config.
		if equalityutils.DeepEqual(initializeBaseRoute(tpsByRequestMatcher[0].Spec, sourceCluster), baseRoute) {
			continue
		}

		// set a default destination for the route (to the target traffictarget)
		// if a decorator has not already set it
		t.setDefaultDestination(baseRoute, destinationFQDN)

		// construct a copy of a route for each service port
		// required because Istio needs the destination port for every route
		routesPerPort := duplicateRouteForEachPort(baseRoute, kubeService.Ports)

		// split routes with multiple HTTP matchers into one matcher per route for easier route sorting later on
		var routesWithSingleMatcher []*networkingv1alpha3spec.HTTPRoute
		for _, route := range routesPerPort {
			splitRoutes := splitRouteByMatchers(route)
			routesWithSingleMatcher = append(routesWithSingleMatcher, splitRoutes...)
		}

		virtualService.Spec.Http = append(virtualService.Spec.Http, routesWithSingleMatcher...)
	}

	sort.Sort(RoutesBySpecificity(virtualService.Spec.Http))

	if len(virtualService.Spec.Http) == 0 {
		// no need to create this VirtualService as it has no effect
		return nil
	}

	if t.userVirtualServices == nil {
		return virtualService
	}

	// detect and report error on intersecting config if enabled in settings
	if errs := conflictsWithUserVirtualService(
		t.userVirtualServices,
		virtualService,
	); len(errs) > 0 {
		for _, err := range errs {
			for _, policy := range trafficTarget.Status.AppliedTrafficPolicies {
				reporter.ReportTrafficPolicyToTrafficTarget(trafficTarget, policy.Ref, err)
			}
		}
		return nil
	}

	return virtualService
}

// ensure that only a single VirtualService HTTPRoute gets created per TrafficPolicy request matcher
// by first grouping TrafficPolicies by semantically equivalent request matchers
func groupAppliedTpsByRequestMatcher(
	appliedTps []*discoveryv1alpha2.TrafficTargetStatus_AppliedTrafficPolicy,
) [][]*discoveryv1alpha2.TrafficTargetStatus_AppliedTrafficPolicy {
	var allGroupedTps [][]*discoveryv1alpha2.TrafficTargetStatus_AppliedTrafficPolicy

	for _, appliedTp := range appliedTps {
		var grouped bool
		for i, groupedTps := range allGroupedTps {
			// append to existing group
			if requestMatchersEqual(appliedTp.Spec, groupedTps[0].Spec) {
				allGroupedTps[i] = append(groupedTps, appliedTp)
				grouped = true
				break
			}
		}
		// create new group
		if !grouped {
			allGroupedTps = append(allGroupedTps, []*discoveryv1alpha2.TrafficTargetStatus_AppliedTrafficPolicy{appliedTp})
		}
	}

	return allGroupedTps
}

func requestMatchersEqual(tp1, tp2 *v1alpha2.TrafficPolicySpec) bool {
	return workloadSelectorListsEqual(tp1.GetSourceSelector(), tp2.GetSourceSelector()) &&
		httpRequestMatchersEqual(tp1.GetHttpRequestMatchers(), tp2.GetHttpRequestMatchers())
}

func httpRequestMatchersEqual(matchers1, matchers2 []*v1alpha2.TrafficPolicySpec_HttpMatcher) bool {
	if len(matchers1) != len(matchers2) {
		return false
	}
	for i := range matchers1 {
		if !proto.Equal(matchers1[i], matchers2[i]) {
			return false
		}
	}
	return true
}

// return true if workload selectors' labels and namespaces are equivalent, ignore clusters
func workloadSelectorsEqual(ws1, ws2 *v1alpha2.WorkloadSelector) bool {
	return reflect.DeepEqual(ws1.Labels, ws2.Labels) &&
		sets.NewString(ws1.Namespaces...).Equal(sets.NewString(ws2.Namespaces...))
}

// return true if two lists of WorkloadSelectors are semantically equivalent, abstracting away order
func workloadSelectorListsEqual(wsList1, wsList2 []*v1alpha2.WorkloadSelector) bool {
	if len(wsList1) != len(wsList2) {
		return false
	}
	matchedWs2 := sets.NewInt()
	for _, ws1 := range wsList1 {
		var matched bool
		for i, ws2 := range wsList2 {
			if matchedWs2.Has(i) {
				continue
			}
			if workloadSelectorsEqual(ws1, ws2) {
				matchedWs2.Insert(i)
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	return true
}

// construct the callback for registering fields in the virtual service
func registerFieldFunc(
	virtualServiceFields fieldutils.FieldOwnershipRegistry,
	virtualService *networkingv1alpha3.VirtualService,
	policyRef ezkube.ResourceId,
) decorators.RegisterField {
	return func(fieldPtr, val interface{}) error {
		fieldVal := reflect.ValueOf(fieldPtr).Elem().Interface()

		if equalityutils.DeepEqual(fieldVal, val) {
			return nil
		}
		if err := virtualServiceFields.RegisterFieldOwnership(
			virtualService,
			fieldPtr,
			[]ezkube.ResourceId{policyRef},
			&v1alpha2.TrafficPolicy{},
			0, //TODO(ilackarms): priority
		); err != nil {
			return err
		}
		return nil
	}
}

func (t *translator) initializeVirtualService(
	trafficTarget *discoveryv1alpha2.TrafficTarget,
	sourceMeshInstallation *discoveryv1alpha2.MeshSpec_MeshInstallation,
	destinationFQDN string,
) *networkingv1alpha3.VirtualService {
	var meta metav1.ObjectMeta
	if sourceMeshInstallation != nil {
		meta = metautils.FederatedObjectMeta(
			trafficTarget.Spec.GetKubeService().Ref,
			sourceMeshInstallation,
			trafficTarget.Annotations,
		)
	} else {
		meta = metautils.TranslatedObjectMeta(
			trafficTarget.Spec.GetKubeService().Ref,
			trafficTarget.Annotations,
		)
	}

	return &networkingv1alpha3.VirtualService{
		ObjectMeta: meta,
		Spec: networkingv1alpha3spec.VirtualService{
			Hosts: []string{destinationFQDN},
		},
	}
}

// Returns nil to prevent translating the trafficPolicy if the sourceClusterName is not selected by the WorkloadSelector
func initializeBaseRoute(trafficPolicy *v1alpha2.TrafficPolicySpec, sourceClusterName string) *networkingv1alpha3spec.HTTPRoute {
	if !selectorutils.WorkloadSelectorContainsCluster(trafficPolicy.SourceSelector, sourceClusterName) {
		return nil
	}
	return &networkingv1alpha3spec.HTTPRoute{
		Match: translateRequestMatchers(trafficPolicy),
	}
}

func (t *translator) setDefaultDestination(
	baseRoute *networkingv1alpha3spec.HTTPRoute,
	destinationFQDN string,
) {
	// if a route destination is already set, we don't need to modify the route
	if baseRoute.Route != nil {
		return
	}

	baseRoute.Route = []*networkingv1alpha3spec.HTTPRouteDestination{{
		Destination: &networkingv1alpha3spec.Destination{
			Host: destinationFQDN,
		},
	}}
}

// construct a copy of a route for each service port
// required because Istio needs the destination port for every route
// if the service has multiple service ports defined
func duplicateRouteForEachPort(baseRoute *networkingv1alpha3spec.HTTPRoute, ports []*discoveryv1alpha2.TrafficTargetSpec_KubeService_KubeServicePort) []*networkingv1alpha3spec.HTTPRoute {
	if len(ports) == 1 {
		// no need to specify port for single-port service
		return []*networkingv1alpha3spec.HTTPRoute{baseRoute}
	}
	var routesWithPort []*networkingv1alpha3spec.HTTPRoute
	for _, port := range ports {
		// create a separate set of matchers for each port on the destination service
		var matchersWithPort []*networkingv1alpha3spec.HTTPMatchRequest

		for _, matcher := range baseRoute.Match {
			matcher := matcher.DeepCopy()
			matcher.Port = port.GetPort()
			matchersWithPort = append(matchersWithPort, matcher)
		}

		var destinationsWithPort []*networkingv1alpha3spec.HTTPRouteDestination

		for _, destination := range baseRoute.Route {
			destination := destination.DeepCopy()
			destination.Destination.Port = &networkingv1alpha3spec.PortSelector{
				Number: port.GetPort(),
			}
			destinationsWithPort = append(destinationsWithPort, destination)
		}

		routeWithPort := baseRoute.DeepCopy()
		routeWithPort.Match = matchersWithPort
		routeWithPort.Route = destinationsWithPort
		routesWithPort = append(routesWithPort, routeWithPort)
	}

	return routesWithPort
}

func splitRouteByMatchers(baseRoute *networkingv1alpha3spec.HTTPRoute) []*networkingv1alpha3spec.HTTPRoute {
	if len(baseRoute.Match) < 1 {
		return []*networkingv1alpha3spec.HTTPRoute{baseRoute}
	}

	var singleMatcherRoutes []*networkingv1alpha3spec.HTTPRoute
	for _, match := range baseRoute.Match {
		singleMatcherRoute := baseRoute.DeepCopy()
		singleMatcherRoute.Match = []*networkingv1alpha3spec.HTTPMatchRequest{match}
		singleMatcherRoutes = append(singleMatcherRoutes, singleMatcherRoute)
	}
	return singleMatcherRoutes
}

func translateRequestMatchers(
	trafficPolicy *v1alpha2.TrafficPolicySpec,
) []*networkingv1alpha3spec.HTTPMatchRequest {
	// Generate HttpMatchRequests for SourceSelector, one per namespace.
	var sourceMatchers []*networkingv1alpha3spec.HTTPMatchRequest
	// Set SourceNamespace and SourceLabels.
	for _, sourceSelector := range trafficPolicy.GetSourceSelector() {
		if len(sourceSelector.GetLabels()) > 0 ||
			len(sourceSelector.GetNamespaces()) > 0 {
			if len(sourceSelector.GetNamespaces()) > 0 {
				for _, namespace := range sourceSelector.GetNamespaces() {
					matchRequest := &networkingv1alpha3spec.HTTPMatchRequest{
						SourceNamespace: namespace,
						SourceLabels:    sourceSelector.GetLabels(),
					}
					sourceMatchers = append(sourceMatchers, matchRequest)
				}
			} else {
				sourceMatchers = append(sourceMatchers, &networkingv1alpha3spec.HTTPMatchRequest{
					SourceLabels: sourceSelector.GetLabels(),
				})
			}
		}
	}
	if trafficPolicy.GetHttpRequestMatchers() == nil {
		return sourceMatchers
	}
	// If HttpRequestMatchers exist, generate cartesian product of sourceMatchers and httpRequestMatchers.
	var translatedRequestMatchers []*networkingv1alpha3spec.HTTPMatchRequest
	// If SourceSelector is nil, generate an HttpMatchRequest without SourceSelector match criteria
	if len(sourceMatchers) == 0 {
		sourceMatchers = append(sourceMatchers, &networkingv1alpha3spec.HTTPMatchRequest{})
	}
	// Set QueryParams, Headers, WithoutHeaders, Uri, and Method.
	for _, sourceMatcher := range sourceMatchers {
		for _, matcher := range trafficPolicy.GetHttpRequestMatchers() {
			httpMatcher := &networkingv1alpha3spec.HTTPMatchRequest{
				SourceNamespace: sourceMatcher.GetSourceNamespace(),
				SourceLabels:    sourceMatcher.GetSourceLabels(),
			}
			headerMatchers, inverseHeaderMatchers := translateRequestMatcherHeaders(matcher.GetHeaders())
			uriMatcher := translateRequestMatcherPathSpecifier(matcher)
			var method *networkingv1alpha3spec.StringMatch
			if matcher.GetMethod() != nil {
				method = &networkingv1alpha3spec.StringMatch{MatchType: &networkingv1alpha3spec.StringMatch_Exact{Exact: matcher.GetMethod().GetMethod().String()}}
			}
			httpMatcher.QueryParams = translateRequestMatcherQueryParams(matcher.GetQueryParameters())
			httpMatcher.Headers = headerMatchers
			httpMatcher.WithoutHeaders = inverseHeaderMatchers
			httpMatcher.Uri = uriMatcher
			httpMatcher.Method = method
			translatedRequestMatchers = append(translatedRequestMatchers, httpMatcher)
		}
	}
	return translatedRequestMatchers
}

func translateRequestMatcherHeaders(matchers []*v1alpha1.HeaderMatcher) (
	map[string]*networkingv1alpha3spec.StringMatch, map[string]*networkingv1alpha3spec.StringMatch,
) {
	headerMatchers := map[string]*networkingv1alpha3spec.StringMatch{}
	inverseHeaderMatchers := map[string]*networkingv1alpha3spec.StringMatch{}
	var matcherMap map[string]*networkingv1alpha3spec.StringMatch
	if matchers != nil {
		for _, matcher := range matchers {
			matcherMap = headerMatchers
			if matcher.GetInvertMatch() {
				matcherMap = inverseHeaderMatchers
			}
			if matcher.GetRegex() {
				matcherMap[matcher.GetName()] = &networkingv1alpha3spec.StringMatch{
					MatchType: &networkingv1alpha3spec.StringMatch_Regex{Regex: matcher.GetValue()},
				}
			} else {
				matcherMap[matcher.GetName()] = &networkingv1alpha3spec.StringMatch{
					MatchType: &networkingv1alpha3spec.StringMatch_Exact{Exact: matcher.GetValue()},
				}
			}
		}
	}
	// ensure field is set to nil if empty
	if len(headerMatchers) == 0 {
		headerMatchers = nil
	}
	if len(inverseHeaderMatchers) == 0 {
		inverseHeaderMatchers = nil
	}
	return headerMatchers, inverseHeaderMatchers
}

func translateRequestMatcherQueryParams(matchers []*v1alpha2.TrafficPolicySpec_QueryParameterMatcher) map[string]*networkingv1alpha3spec.StringMatch {
	var translatedQueryParamMatcher map[string]*networkingv1alpha3spec.StringMatch
	if matchers != nil {
		translatedQueryParamMatcher = map[string]*networkingv1alpha3spec.StringMatch{}
		for _, matcher := range matchers {
			if matcher.GetRegex() {
				translatedQueryParamMatcher[matcher.GetName()] = &networkingv1alpha3spec.StringMatch{
					MatchType: &networkingv1alpha3spec.StringMatch_Regex{Regex: matcher.GetValue()},
				}
			} else {
				translatedQueryParamMatcher[matcher.GetName()] = &networkingv1alpha3spec.StringMatch{
					MatchType: &networkingv1alpha3spec.StringMatch_Exact{Exact: matcher.GetValue()},
				}
			}
		}
	}
	return translatedQueryParamMatcher
}

func translateRequestMatcherPathSpecifier(matcher *v1alpha2.TrafficPolicySpec_HttpMatcher) *networkingv1alpha3spec.StringMatch {
	if matcher != nil && matcher.GetPathSpecifier() != nil {
		switch pathSpecifierType := matcher.GetPathSpecifier().(type) {
		case *v1alpha2.TrafficPolicySpec_HttpMatcher_Exact:
			return &networkingv1alpha3spec.StringMatch{MatchType: &networkingv1alpha3spec.StringMatch_Exact{Exact: pathSpecifierType.Exact}}
		case *v1alpha2.TrafficPolicySpec_HttpMatcher_Prefix:
			return &networkingv1alpha3spec.StringMatch{MatchType: &networkingv1alpha3spec.StringMatch_Prefix{Prefix: pathSpecifierType.Prefix}}
		case *v1alpha2.TrafficPolicySpec_HttpMatcher_Regex:
			return &networkingv1alpha3spec.StringMatch{MatchType: &networkingv1alpha3spec.StringMatch_Regex{Regex: pathSpecifierType.Regex}}
		}
	}
	return nil
}

// Return errors for each user-supplied VirtualService that applies to the same hostname as the translated VirtualService
func conflictsWithUserVirtualService(
	userVirtualServices v1alpha3sets.VirtualServiceSet,
	translatedVirtualService *networkingv1alpha3.VirtualService,
) []error {
	// For each user VS, check whether any hosts match any hosts from translated VS
	var errs []error

	// virtual services from RemoteSnapshot only contain non-translated objects
	userVirtualServices.List(func(vs *networkingv1alpha3.VirtualService) (_ bool) {
		// check if common hostnames exist
		commonHostnames := utils.CommonHostnames(vs.Spec.Hosts, translatedVirtualService.Spec.Hosts)
		if len(commonHostnames) > 0 {
			errs = append(
				errs,
				eris.Errorf("Unable to translate AppliedTrafficPolicies to VirtualService, applies to hosts %+v that are already configured by the existing VirtualService %s", commonHostnames, skv2sets.Key(vs)),
			)
		}
		return
	})

	return errs
}
