package virtualservice

import (
	"reflect"
	"sort"

	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators"
	"github.com/solo-io/skv2/pkg/ezkube"

	"github.com/rotisserie/eris"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/equalityutils"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/fieldutils"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/hostutils"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/metautils"
	networkingv1alpha3spec "istio.io/api/networking/v1alpha3"
	networkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
)

//go:generate mockgen -source ./virtual_service_translator.go -destination mocks/virtual_service_translator.go

// the VirtualService translator translates a TrafficTarget into a VirtualService.
type Translator interface {
	// Translate translates the appropriate VirtualService for the given TrafficTarget.
	// returns nil if no VirtualService is required for the TrafficTarget (i.e. if no VirtualService features are required, such as subsets).
	//
	// Errors caused by invalid user config will be reported using the Reporter.
	//
	// Note that the input snapshot TrafficTargetSet contains the given TrafficTarget.
	Translate(
		in input.Snapshot,
		trafficTarget *discoveryv1alpha2.TrafficTarget,
		reporter reporting.Reporter,
	) *networkingv1alpha3.VirtualService
}

type translator struct {
	clusterDomains   hostutils.ClusterDomainRegistry
	decoratorFactory decorators.Factory
}

func NewTranslator(clusterDomains hostutils.ClusterDomainRegistry, decoratorFactory decorators.Factory) Translator {
	return &translator{clusterDomains: clusterDomains, decoratorFactory: decoratorFactory}
}

// translate the appropriate VirtualService for the given TrafficTarget.
// returns nil if no VirtualService is required for the TrafficTarget (i.e. if no VirtualService features are required, such as subsets).
// The input snapshot TrafficTargetSet contains n the
func (t *translator) Translate(
	in input.Snapshot,
	trafficTarget *discoveryv1alpha2.TrafficTarget,
	reporter reporting.Reporter,
) *networkingv1alpha3.VirtualService {
	kubeService := trafficTarget.Spec.GetKubeService()

	if kubeService == nil {
		// TODO(ilackarms): non kube services currently unsupported
		return nil
	}

	virtualService := t.initializeVirtualService(trafficTarget)
	// register the owners of the virtualservice fields
	virtualServiceFields := fieldutils.NewOwnershipRegistry()
	vsDecorators := t.decoratorFactory.MakeDecorators(decorators.Parameters{
		ClusterDomains: t.clusterDomains,
		Snapshot:       in,
	})

	for _, policy := range trafficTarget.Status.AppliedTrafficPolicies {
		baseRoute := initializeBaseRoute(policy.Spec)
		registerField := registerFieldFunc(virtualServiceFields, virtualService, policy.Ref)
		for _, decorator := range vsDecorators {

			if trafficPolicyDecorator, ok := decorator.(decorators.TrafficPolicyVirtualServiceDecorator); ok {
				if err := trafficPolicyDecorator.ApplyTrafficPolicyToVirtualService(
					policy,
					trafficTarget,
					baseRoute,
					registerField,
				); err != nil {
					reporter.ReportTrafficPolicyToTrafficTarget(trafficTarget, policy.Ref, eris.Wrapf(err, "%v", decorator.DecoratorName()))
				}
			}
		}

		// Avoid appending an HttpRoute that will have no affect, which occurs if no decorators mutate the baseRoute with any non-match config.
		if equalityutils.Equals(initializeBaseRoute(policy.Spec), baseRoute) {
			continue
		}

		// set a default destination for the route (to the target traffictarget)
		// if a decorator has not already set it
		t.setDefaultDestination(baseRoute, trafficTarget)

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

	return virtualService
}

// construct the callback for registering fields in the virtual service
func registerFieldFunc(
	virtualServiceFields fieldutils.FieldOwnershipRegistry,
	virtualService *networkingv1alpha3.VirtualService,
	policyRef ezkube.ResourceId,
) decorators.RegisterField {
	return func(fieldPtr, val interface{}) error {
		fieldVal := reflect.ValueOf(fieldPtr).Elem().Interface()

		if equalityutils.Equals(fieldVal, val) {
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

func (t *translator) initializeVirtualService(trafficTarget *discoveryv1alpha2.TrafficTarget) *networkingv1alpha3.VirtualService {
	meta := metautils.TranslatedObjectMeta(
		trafficTarget.Spec.GetKubeService().Ref,
		trafficTarget.Annotations,
	)

	hosts := []string{t.clusterDomains.GetServiceLocalFQDN(trafficTarget.Spec.GetKubeService().Ref)}

	return &networkingv1alpha3.VirtualService{
		ObjectMeta: meta,
		Spec: networkingv1alpha3spec.VirtualService{
			Hosts: hosts,
		},
	}
}

func initializeBaseRoute(trafficPolicy *v1alpha2.TrafficPolicySpec) *networkingv1alpha3spec.HTTPRoute {
	return &networkingv1alpha3spec.HTTPRoute{
		Match: translateRequestMatchers(trafficPolicy),
	}
}

func (t *translator) setDefaultDestination(baseRoute *networkingv1alpha3spec.HTTPRoute, trafficTarget *discoveryv1alpha2.TrafficTarget) {
	// if a route destination is already set, we don't need to modify the route
	if baseRoute.Route != nil {
		return
	}

	baseRoute.Route = []*networkingv1alpha3spec.HTTPRouteDestination{{
		Destination: &networkingv1alpha3spec.Destination{
			Host: t.clusterDomains.GetServiceLocalFQDN(trafficTarget.Spec.GetKubeService().GetRef()),
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

func translateRequestMatcherHeaders(matchers []*v1alpha2.TrafficPolicySpec_HeaderMatcher) (
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
