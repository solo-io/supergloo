package meshservice

import (
	"github.com/rotisserie/eris"
	discoveryv1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	"github.com/solo-io/smh/pkg/mesh-networking/reporter"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/utils/fieldutils"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/utils/hostutils"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/utils/metautils"
	istiov1alpha3spec "istio.io/api/networking/v1alpha3"
	istiov1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
)

// the VirtualService translator translates a MeshService into a VirtualService.
type Translator interface {
	// Translate translates the appropriate resources for the given Mesh.
	//
	// returns empty sets if no VirtualService or DestinationRule is required for the MeshService (i.e. if no VirtualService/DestinationRule features are required, such as subsets).
	//
	// Errors caused by invalid user config will be reported using the Reporter.
	TranslateMesh(
		mesh *discoveryv1alpha1.Mesh,
		reporter reporter.Reporter,
	) (*istiov1alpha3.VirtualService, *istiov1alpha3.DestinationRule)
}

type translator struct {
	clusterDomains hostutils.ClusterDomainRegistry
	pluginFactory  plugins.Factory
}

func NewTranslator(clusterDomains hostutils.ClusterDomainRegistry, pluginFactory plugins.Factory) Translator {
	return &translator{clusterDomains: clusterDomains, pluginFactory: pluginFactory}
}

// translate the appropriate VirtualService for the given MeshService.
// returns nil if no VirtualService is required for the MeshService (i.e. if no VirtualService features are required, such as subsets).
// The input snapshot MeshServiceSet contains n the
func (t *translator) Translate(
	in input.Snapshot,
	meshService *discoveryv1alpha1.MeshService,
	reporter reporter.Reporter,
) *istiov1alpha3.VirtualService {
	plugins := t.pluginFactory.MakePlugins(
		t.clusterDomains,
		in,
	)
	virtualService := t.initializeVirtualService(meshService)
	// register the owners of the virtualservice fields
	virtualServiceFields := fieldutils.NewOwnershipRegistry()

	for _, plug := range plugins {
		if simplePlugin, ok := plug.(plugins.SimplePlugin); ok {
			simplePlugin.Process(meshService, virtualService)
		}
	}

	for _, policy := range meshService.Status.AppliedTrafficPolicies {
		baseRoute := initializeBaseRoute(policy.Spec)
		for _, plug := range plugins {
			if trafficPolicyPlugin, ok := plug.(plugins.TrafficPolicyPlugin); ok {
				if err := trafficPolicyPlugin.ProcessTrafficPolicy(
					policy,
					meshService,
					baseRoute,
					virtualServiceFields,
				); err != nil {
					reporter.ReportTrafficPolicy(meshService, policy.Ref, eris.Wrapf(err, "%v", plug.PluginName()))
				}
			}
		}

		// set a default destination for the route (to the target meshservice)
		// if a plugin has not already set it
		t.setDefaultDestination(baseRoute, meshService)

		// construct a copy of a route for each service port
		// required because Istio needs the destination port for every route
		routesPerPort := duplicateRouteForEachPort(baseRoute, meshService.Spec.KubeService.Ports)

		// split routes with multiple HTTP matchers for easier route sorting later on
		var routesWithSingleMatcher []*istiov1alpha3spec.HTTPRoute
		for _, route := range routesPerPort {
			splitRoutes := splitRouteByMatchers(route)
			routesWithSingleMatcher = append(routesWithSingleMatcher, splitRoutes...)
		}

		virtualService.Spec.Http = append(virtualService.Spec.Http, routesWithSingleMatcher...)
	}

	for _, policy := range meshService.Status.AppliedAccessPolicies {
		for _, plug := range plugins {
			if accessPolicyPlugin, ok := plug.(plugins.AccessPolicyPlugin); ok {
				if err := accessPolicyPlugin.ProcessAccessPolicy(
					policy,
					meshService,
					virtualService,
					virtualServiceFields,
				); err != nil {
					reporter.ReportAccessPolicy(meshService, policy.Ref, eris.Wrapf(err, "%v", plug.PluginName()))
				}
			}
		}
	}

	if len(virtualService.Spec.Http) == 0 {
		// no need to create this VirtualService as it has no effect
		return nil
	}

	return virtualService
}

func (t *translator) initializeVirtualService(meshService *discoveryv1alpha1.MeshService) *istiov1alpha3.VirtualService {
	meta := metautils.TranslatedObjectMeta(
		meshService.Spec.KubeService.Ref,
		meshService.Annotations,
	)

	hosts := []string{t.clusterDomains.GetServiceLocalFQDN(meshService.Spec.KubeService.Ref)}

	return &istiov1alpha3.VirtualService{
		ObjectMeta: meta,
		Spec: istiov1alpha3spec.VirtualService{
			Hosts: hosts,
		},
	}
}

func initializeBaseRoute(trafficPolicy *v1alpha1.TrafficPolicySpec) *istiov1alpha3spec.HTTPRoute {
	return &istiov1alpha3spec.HTTPRoute{
		Match: translateRequestMatchers(trafficPolicy),
	}
}

func (t *translator) setDefaultDestination(baseRoute *istiov1alpha3spec.HTTPRoute, meshService *discoveryv1alpha1.MeshService) {
	// if a route destination is already set, we don't need to modify the route
	if baseRoute.Route != nil {
		return
	}

	baseRoute.Route = []*istiov1alpha3spec.HTTPRouteDestination{{
		Destination: &istiov1alpha3spec.Destination{
			Host: t.clusterDomains.GetServiceLocalFQDN(meshService.Spec.GetKubeService().GetRef()),
		},
	}}
}

// construct a copy of a route for each service port
// required because Istio needs the destination port for every route
// if the service has multiple service ports defined
func duplicateRouteForEachPort(baseRoute *istiov1alpha3spec.HTTPRoute, ports []*discoveryv1alpha1.MeshServiceSpec_KubeService_KubeServicePort) []*istiov1alpha3spec.HTTPRoute {
	if len(ports) == 1 {
		// no need to specify port for single-port service
		return []*istiov1alpha3spec.HTTPRoute{baseRoute}
	}
	var routesWithPort []*istiov1alpha3spec.HTTPRoute
	for _, port := range ports {
		// create a separate set of matchers for each port on the destination service
		var matchersWithPort []*istiov1alpha3spec.HTTPMatchRequest

		for _, matcher := range baseRoute.Match {
			matcher := matcher.DeepCopy()
			matcher.Port = port.GetPort()
			matchersWithPort = append(matchersWithPort, matcher)
		}

		var destinationsWithPort []*istiov1alpha3spec.HTTPRouteDestination

		for _, destination := range baseRoute.Route {
			destination := destination.DeepCopy()
			destination.Destination.Port = &istiov1alpha3spec.PortSelector{
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

func splitRouteByMatchers(baseRoute *istiov1alpha3spec.HTTPRoute) []*istiov1alpha3spec.HTTPRoute {
	if len(baseRoute.Match) < 1 {
		return []*istiov1alpha3spec.HTTPRoute{baseRoute}
	}

	var singleMatcherRoutes []*istiov1alpha3spec.HTTPRoute
	for _, match := range baseRoute.Match {
		singleMatcherRoute := baseRoute.DeepCopy()
		singleMatcherRoute.Match = []*istiov1alpha3spec.HTTPMatchRequest{match}
		singleMatcherRoutes = append(singleMatcherRoutes, singleMatcherRoute)
	}
	return singleMatcherRoutes
}

func translateRequestMatchers(
	trafficPolicy *v1alpha1.TrafficPolicySpec,
) []*istiov1alpha3spec.HTTPMatchRequest {
	// Generate HttpMatchRequests for SourceSelector, one per namespace.
	var sourceMatchers []*istiov1alpha3spec.HTTPMatchRequest
	// Set SourceNamespace and SourceLabels.
	for _, sourceSelector := range trafficPolicy.GetSourceSelector() {
		if len(sourceSelector.GetLabels()) > 0 ||
			len(sourceSelector.GetNamespaces()) > 0 {
			if len(sourceSelector.GetNamespaces()) > 0 {
				for _, namespace := range sourceSelector.GetNamespaces() {
					matchRequest := &istiov1alpha3spec.HTTPMatchRequest{
						SourceNamespace: namespace,
						SourceLabels:    sourceSelector.GetLabels(),
					}
					sourceMatchers = append(sourceMatchers, matchRequest)
				}
			} else {
				sourceMatchers = append(sourceMatchers, &istiov1alpha3spec.HTTPMatchRequest{
					SourceLabels: sourceSelector.GetLabels(),
				})
			}
		}
	}
	if trafficPolicy.GetHttpRequestMatchers() == nil {
		return sourceMatchers
	}
	// If HttpRequestMatchers exist, generate cartesian product of sourceMatchers and httpRequestMatchers.
	var translatedRequestMatchers []*istiov1alpha3spec.HTTPMatchRequest
	// If SourceSelector is nil, generate an HttpMatchRequest without SourceSelector match criteria
	if len(sourceMatchers) == 0 {
		sourceMatchers = append(sourceMatchers, &istiov1alpha3spec.HTTPMatchRequest{})
	}
	// Set QueryParams, Headers, WithoutHeaders, Uri, and Method.
	for _, sourceMatcher := range sourceMatchers {
		for _, matcher := range trafficPolicy.GetHttpRequestMatchers() {
			httpMatcher := &istiov1alpha3spec.HTTPMatchRequest{
				SourceNamespace: sourceMatcher.GetSourceNamespace(),
				SourceLabels:    sourceMatcher.GetSourceLabels(),
			}
			headerMatchers, inverseHeaderMatchers := translateRequestMatcherHeaders(matcher.GetHeaders())
			uriMatcher := translateRequestMatcherPathSpecifier(matcher)
			var method *istiov1alpha3spec.StringMatch
			if matcher.GetMethod() != nil {
				method = &istiov1alpha3spec.StringMatch{MatchType: &istiov1alpha3spec.StringMatch_Exact{Exact: matcher.GetMethod().GetMethod().String()}}
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

func translateRequestMatcherHeaders(matchers []*v1alpha1.TrafficPolicySpec_HeaderMatcher) (
	map[string]*istiov1alpha3spec.StringMatch, map[string]*istiov1alpha3spec.StringMatch,
) {
	headerMatchers := map[string]*istiov1alpha3spec.StringMatch{}
	inverseHeaderMatchers := map[string]*istiov1alpha3spec.StringMatch{}
	var matcherMap map[string]*istiov1alpha3spec.StringMatch
	if matchers != nil {
		for _, matcher := range matchers {
			matcherMap = headerMatchers
			if matcher.GetInvertMatch() {
				matcherMap = inverseHeaderMatchers
			}
			if matcher.GetRegex() {
				matcherMap[matcher.GetName()] = &istiov1alpha3spec.StringMatch{
					MatchType: &istiov1alpha3spec.StringMatch_Regex{Regex: matcher.GetValue()},
				}
			} else {
				matcherMap[matcher.GetName()] = &istiov1alpha3spec.StringMatch{
					MatchType: &istiov1alpha3spec.StringMatch_Exact{Exact: matcher.GetValue()},
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

func translateRequestMatcherQueryParams(matchers []*v1alpha1.TrafficPolicySpec_QueryParameterMatcher) map[string]*istiov1alpha3spec.StringMatch {
	var translatedQueryParamMatcher map[string]*istiov1alpha3spec.StringMatch
	if matchers != nil {
		translatedQueryParamMatcher = map[string]*istiov1alpha3spec.StringMatch{}
		for _, matcher := range matchers {
			if matcher.GetRegex() {
				translatedQueryParamMatcher[matcher.GetName()] = &istiov1alpha3spec.StringMatch{
					MatchType: &istiov1alpha3spec.StringMatch_Regex{Regex: matcher.GetValue()},
				}
			} else {
				translatedQueryParamMatcher[matcher.GetName()] = &istiov1alpha3spec.StringMatch{
					MatchType: &istiov1alpha3spec.StringMatch_Exact{Exact: matcher.GetValue()},
				}
			}
		}
	}
	return translatedQueryParamMatcher
}

func translateRequestMatcherPathSpecifier(matcher *v1alpha1.TrafficPolicySpec_HttpMatcher) *istiov1alpha3spec.StringMatch {
	if matcher != nil && matcher.GetPathSpecifier() != nil {
		switch pathSpecifierType := matcher.GetPathSpecifier().(type) {
		case *v1alpha1.TrafficPolicySpec_HttpMatcher_Exact:
			return &istiov1alpha3spec.StringMatch{MatchType: &istiov1alpha3spec.StringMatch_Exact{Exact: pathSpecifierType.Exact}}
		case *v1alpha1.TrafficPolicySpec_HttpMatcher_Prefix:
			return &istiov1alpha3spec.StringMatch{MatchType: &istiov1alpha3spec.StringMatch_Prefix{Prefix: pathSpecifierType.Prefix}}
		case *v1alpha1.TrafficPolicySpec_HttpMatcher_Regex:
			return &istiov1alpha3spec.StringMatch{MatchType: &istiov1alpha3spec.StringMatch_Regex{Regex: pathSpecifierType.Regex}}
		}
	}
	return nil
}
