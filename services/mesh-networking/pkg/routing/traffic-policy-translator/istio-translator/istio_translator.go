package istio_translator

import (
	"context"
	"sort"
	"strings"

	"github.com/rotisserie/eris"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	istio_networking "github.com/solo-io/service-mesh-hub/pkg/api/istio/networking/v1alpha3"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	zephyr_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/clients"
	"github.com/solo-io/service-mesh-hub/pkg/selector"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/mesh-platform/k8s"
	traffic_policy_translator "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/routing/traffic-policy-translator"
	istio_networking_types "istio.io/api/networking/v1alpha3"
	istio_client_networking_types "istio.io/client-go/pkg/apis/networking/v1alpha3"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
)

const (
	TranslatorId = "istio_translator"
)

type IstioTranslator traffic_policy_translator.TrafficPolicyMeshTranslator

func NewIstioTrafficPolicyTranslator(
	dynamicClientGetter mc_manager.DynamicClientGetter,
	meshClient zephyr_discovery.MeshClient,
	meshServiceClient zephyr_discovery.MeshServiceClient,
	resourceSelector selector.ResourceSelector,
	virtualServiceClientFactory istio_networking.VirtualServiceClientFactory,
	destinationRuleClientFactory istio_networking.DestinationRuleClientFactory,
) IstioTranslator {
	return &istioTrafficPolicyTranslator{
		dynamicClientGetter:          dynamicClientGetter,
		meshClient:                   meshClient,
		meshServiceClient:            meshServiceClient,
		resourceSelector:             resourceSelector,
		virtualServiceClientFactory:  virtualServiceClientFactory,
		destinationRuleClientFactory: destinationRuleClientFactory,
	}
}

type istioTrafficPolicyTranslator struct {
	dynamicClientGetter          mc_manager.DynamicClientGetter
	meshClient                   zephyr_discovery.MeshClient
	meshServiceClient            zephyr_discovery.MeshServiceClient
	virtualServiceClientFactory  istio_networking.VirtualServiceClientFactory
	destinationRuleClientFactory istio_networking.DestinationRuleClientFactory
	resourceSelector             selector.ResourceSelector
}

var (
	NoSpecifiedPortError = func(svc *zephyr_discovery.MeshService) error {
		return eris.Errorf("Mesh service %s.%s ports list does not include just one entry, so no default can be used. "+
			"Must specify a destination with a port", svc.Name, svc.Namespace)
	}
	MultiClusterSubsetsNotSupported = func(dest *zephyr_networking_types.TrafficPolicySpec_MultiDestination_WeightedDestination) error {
		return eris.Errorf("Multi cluster subsets are currently not supported, found one on destination: %+v", dest)
	}
)

func (i *istioTrafficPolicyTranslator) Name() string {
	return TranslatorId
}

/*
	Translate a TrafficPolicy into the following Istio specific configuration:
	https://istio.io/docs/concepts/traffic-management/

	1. VirtualService - routing rules (e.g. retries, fault injection, traffic shifts)
	2. DestinationRule - post-routing rules (e.g. subset declaration)
*/
func (i *istioTrafficPolicyTranslator) TranslateTrafficPolicy(
	ctx context.Context,
	meshService *zephyr_discovery.MeshService,
	mesh *zephyr_discovery.Mesh,
	mergedTrafficPolicies []*zephyr_networking.TrafficPolicy,
) *zephyr_networking_types.TrafficPolicyStatus_TranslatorError {
	if mesh.Spec.GetIstio() == nil {
		return nil
	}
	destinationRuleClient, virtualServiceClient, err := i.fetchClientsForMeshService(ctx, meshService)
	if err != nil {
		return i.errorToStatus(err)
	}
	translatorError := i.ensureDestinationRule(ctx, meshService, destinationRuleClient)
	if translatorError != nil {
		return translatorError
	}
	translatorError = i.ensureVirtualService(ctx, meshService, mergedTrafficPolicies, virtualServiceClient)
	if translatorError != nil {
		return translatorError
	}
	return nil
}

// get DestinationRule and VirtualService clients for MeshService's cluster
func (i *istioTrafficPolicyTranslator) fetchClientsForMeshService(
	ctx context.Context,
	meshService *zephyr_discovery.MeshService,
) (istio_networking.DestinationRuleClient, istio_networking.VirtualServiceClient, error) {
	clusterName, err := i.getClusterNameForMeshService(ctx, meshService)
	if err != nil {
		return nil, nil, err
	}
	dynamicClient, err := i.dynamicClientGetter.GetClientForCluster(ctx, clusterName)
	if err != nil {
		return nil, nil, err
	}
	return i.destinationRuleClientFactory(dynamicClient), i.virtualServiceClientFactory(dynamicClient), nil
}

func (i *istioTrafficPolicyTranslator) ensureDestinationRule(
	ctx context.Context,
	meshService *zephyr_discovery.MeshService,
	destinationRuleClient istio_networking.DestinationRuleClient,
) *zephyr_networking_types.TrafficPolicyStatus_TranslatorError {
	destinationRule := &istio_client_networking_types.DestinationRule{
		ObjectMeta: clients.ResourceRefToObjectMeta(meshService.Spec.GetKubeService().GetRef()),
		Spec: istio_networking_types.DestinationRule{
			Host: buildServiceHostname(meshService),
			TrafficPolicy: &istio_networking_types.TrafficPolicy{
				Tls: &istio_networking_types.TLSSettings{
					Mode: istio_networking_types.TLSSettings_ISTIO_MUTUAL,
				},
			},
		},
	}
	// Only attempt to create if does not already exist
	err := destinationRuleClient.CreateDestinationRule(ctx, destinationRule)
	if err != nil && !errors.IsAlreadyExists(err) {
		return i.errorToStatus(err)
	}
	return nil
}

func (i *istioTrafficPolicyTranslator) ensureVirtualService(
	ctx context.Context,
	meshService *zephyr_discovery.MeshService,
	mergedTrafficPolicies []*zephyr_networking.TrafficPolicy,
	virtualServiceClient istio_networking.VirtualServiceClient,
) *zephyr_networking_types.TrafficPolicyStatus_TranslatorError {
	computedVirtualService, err := i.translateIntoVirtualService(ctx, meshService, mergedTrafficPolicies)
	if err != nil {
		return i.errorToStatus(err)
	}
	// The translator will attempt to create a virtual service for every MeshService.
	// However, a mesh service with no Http, Tcp, or Tls Routes is invalid, so as we only support Http here,
	// we simply return nil. This will not be true for MeshServices which have been configured by TrafficPolicies
	if len(computedVirtualService.Spec.GetHttp()) == 0 {
		return nil
	}
	// Upsert computed VirtualService
	err = virtualServiceClient.UpsertVirtualServiceSpec(ctx, computedVirtualService)
	if err != nil {
		return i.errorToStatus(err)
	}
	return nil
}

func (i *istioTrafficPolicyTranslator) translateIntoVirtualService(
	ctx context.Context,
	meshService *zephyr_discovery.MeshService,
	trafficPolicies []*zephyr_networking.TrafficPolicy,
) (*istio_client_networking_types.VirtualService, error) {
	virtualService := &istio_client_networking_types.VirtualService{
		ObjectMeta: clients.ResourceRefToObjectMeta(meshService.Spec.GetKubeService().GetRef()),
		Spec: istio_networking_types.VirtualService{
			Hosts: []string{buildServiceHostname(meshService)},
		},
	}
	var allHttpRoutes []*istio_networking_types.HTTPRoute
	for _, trafficPolicy := range trafficPolicies {
		httpRoutes, err := i.translateIntoHTTPRoutes(ctx, meshService, trafficPolicy)
		if err != nil {
			return nil, err
		}
		allHttpRoutes = append(allHttpRoutes, httpRoutes...)
	}
	sort.Sort(SpecificitySortableRoutes(allHttpRoutes))
	virtualService.Spec.Http = allHttpRoutes
	return virtualService, nil
}

func (i *istioTrafficPolicyTranslator) translateIntoHTTPRoutes(
	ctx context.Context,
	meshService *zephyr_discovery.MeshService,
	trafficPolicy *zephyr_networking.TrafficPolicy,
) ([]*istio_networking_types.HTTPRoute, error) {
	var err error
	var faultInjection *istio_networking_types.HTTPFaultInjection
	var corsPolicy *istio_networking_types.CorsPolicy
	var requestMatchers []*istio_networking_types.HTTPMatchRequest
	var mirrorPercentage *istio_networking_types.Percent
	var mirror *istio_networking_types.Destination
	var trafficShift []*istio_networking_types.HTTPRouteDestination
	if faultInjection, err = i.translateFaultInjection(trafficPolicy); err != nil {
		return nil, err
	}
	if corsPolicy, err = i.translateCorsPolicy(trafficPolicy); err != nil {
		return nil, err
	}
	if requestMatchers, err = i.translateRequestMatchers(trafficPolicy); err != nil {
		return nil, err
	}
	if trafficPolicy.Spec.GetMirror() != nil {
		mirrorPercentage = &istio_networking_types.Percent{Value: trafficPolicy.Spec.GetMirror().GetPercentage()}
	}
	if mirror, err = i.translateMirror(ctx, meshService, trafficPolicy); err != nil {
		return nil, err
	}
	if trafficShift, err = i.translateDestinationRoutes(ctx, meshService, trafficPolicy); err != nil {
		return nil, err
	}
	retries := i.translateRetries(trafficPolicy)
	headerManipulation := i.translateHeaderManipulation(trafficPolicy)
	var httpRoutes []*istio_networking_types.HTTPRoute

	if len(requestMatchers) == 0 {
		// If no matchers are present return a single route with no matchers
		httpRoutes = append(httpRoutes, &istio_networking_types.HTTPRoute{
			Route:            trafficShift,
			Timeout:          trafficPolicy.Spec.GetRequestTimeout(),
			Fault:            faultInjection,
			CorsPolicy:       corsPolicy,
			Retries:          retries,
			MirrorPercentage: mirrorPercentage,
			Mirror:           mirror,
			Headers:          headerManipulation,
		})
	} else {
		httpRoutes = make([]*istio_networking_types.HTTPRoute, 0, len(requestMatchers))
		// flatten HTTPMatchRequests, i.e. create an HTTPRoute per HTTPMatchRequest
		// this facilitates sorting the HTTPRoutes to produce a well-defined ordering of precedence
		for _, requestMatcher := range requestMatchers {
			httpRoutes = append(httpRoutes, &istio_networking_types.HTTPRoute{
				Match:            []*istio_networking_types.HTTPMatchRequest{requestMatcher},
				Route:            trafficShift,
				Timeout:          trafficPolicy.Spec.GetRequestTimeout(),
				Fault:            faultInjection,
				CorsPolicy:       corsPolicy,
				Retries:          retries,
				MirrorPercentage: mirrorPercentage,
				Mirror:           mirror,
				Headers:          headerManipulation,
			})
		}
	}
	return httpRoutes, nil
}

func (i *istioTrafficPolicyTranslator) translateRequestMatchers(
	trafficPolicy *zephyr_networking.TrafficPolicy,
) ([]*istio_networking_types.HTTPMatchRequest, error) {
	// Generate HttpMatchRequests for SourceSelector, one per namespace.
	var sourceMatchers []*istio_networking_types.HTTPMatchRequest
	// Set SourceNamespace and SourceLabels.
	if len(trafficPolicy.Spec.GetSourceSelector().GetLabels()) > 0 ||
		len(trafficPolicy.Spec.GetSourceSelector().GetNamespaces()) > 0 {
		if len(trafficPolicy.Spec.GetSourceSelector().GetNamespaces()) > 0 {
			for _, namespace := range trafficPolicy.Spec.GetSourceSelector().GetNamespaces() {
				matchRequest := &istio_networking_types.HTTPMatchRequest{
					SourceNamespace: namespace,
					SourceLabels:    trafficPolicy.Spec.GetSourceSelector().GetLabels(),
				}
				sourceMatchers = append(sourceMatchers, matchRequest)
			}
		} else {
			sourceMatchers = append(sourceMatchers, &istio_networking_types.HTTPMatchRequest{
				SourceLabels: trafficPolicy.Spec.GetSourceSelector().GetLabels(),
			})
		}
	}
	if trafficPolicy.Spec.GetHttpRequestMatchers() == nil {
		return sourceMatchers, nil
	}
	// If HttpRequestMatchers exist, generate cartesian product of sourceMatchers and httpRequestMatchers.
	var translatedRequestMatchers []*istio_networking_types.HTTPMatchRequest
	// If SourceSelector is nil, generate an HttpMatchRequest without SourceSelector match criteria
	if len(sourceMatchers) == 0 {
		sourceMatchers = append(sourceMatchers, &istio_networking_types.HTTPMatchRequest{})
	}
	// Set QueryParams, Headers, WithoutHeaders, Uri, and Method.
	for _, sourceMatcher := range sourceMatchers {
		for _, matcher := range trafficPolicy.Spec.GetHttpRequestMatchers() {
			httpMatcher := &istio_networking_types.HTTPMatchRequest{
				SourceNamespace: sourceMatcher.GetSourceNamespace(),
				SourceLabels:    sourceMatcher.GetSourceLabels(),
			}
			headerMatchers, inverseHeaderMatchers := i.translateRequestMatcherHeaders(matcher.GetHeaders())
			uriMatcher, err := i.translateRequestMatcherPathSpecifier(matcher)
			if err != nil {
				return nil, err
			}
			var method *istio_networking_types.StringMatch
			if matcher.GetMethod() != nil {
				method = &istio_networking_types.StringMatch{MatchType: &istio_networking_types.StringMatch_Exact{Exact: matcher.GetMethod().GetMethod().String()}}
			}
			httpMatcher.QueryParams = i.translateRequestMatcherQueryParams(matcher.GetQueryParameters())
			httpMatcher.Headers = headerMatchers
			httpMatcher.WithoutHeaders = inverseHeaderMatchers
			httpMatcher.Uri = uriMatcher
			httpMatcher.Method = method
			translatedRequestMatchers = append(translatedRequestMatchers, httpMatcher)
		}
	}
	return translatedRequestMatchers, nil
}

func (i *istioTrafficPolicyTranslator) translateRequestMatcherPathSpecifier(matcher *zephyr_networking_types.TrafficPolicySpec_HttpMatcher) (*istio_networking_types.StringMatch, error) {
	if matcher != nil && matcher.GetPathSpecifier() != nil {
		switch pathSpecifierType := matcher.GetPathSpecifier().(type) {
		case *zephyr_networking_types.TrafficPolicySpec_HttpMatcher_Exact:
			return &istio_networking_types.StringMatch{MatchType: &istio_networking_types.StringMatch_Exact{Exact: matcher.GetExact()}}, nil
		case *zephyr_networking_types.TrafficPolicySpec_HttpMatcher_Prefix:
			return &istio_networking_types.StringMatch{MatchType: &istio_networking_types.StringMatch_Prefix{Prefix: matcher.GetPrefix()}}, nil
		case *zephyr_networking_types.TrafficPolicySpec_HttpMatcher_Regex:
			return &istio_networking_types.StringMatch{MatchType: &istio_networking_types.StringMatch_Regex{Regex: matcher.GetRegex()}}, nil
		default:
			return nil, eris.Errorf("RequestMatchers[].PathSpecifier has unexpected type %T", pathSpecifierType)
		}
	}
	return nil, nil
}

func (i *istioTrafficPolicyTranslator) translateRequestMatcherQueryParams(matchers []*zephyr_networking_types.TrafficPolicySpec_QueryParameterMatcher) map[string]*istio_networking_types.StringMatch {
	var translatedQueryParamMatcher map[string]*istio_networking_types.StringMatch
	if matchers != nil {
		translatedQueryParamMatcher = map[string]*istio_networking_types.StringMatch{}
		for _, matcher := range matchers {
			if matcher.GetRegex() {
				translatedQueryParamMatcher[matcher.GetName()] = &istio_networking_types.StringMatch{
					MatchType: &istio_networking_types.StringMatch_Regex{Regex: matcher.GetValue()},
				}
			} else {
				translatedQueryParamMatcher[matcher.GetName()] = &istio_networking_types.StringMatch{
					MatchType: &istio_networking_types.StringMatch_Exact{Exact: matcher.GetValue()},
				}
			}
		}
	}
	return translatedQueryParamMatcher
}

func (i *istioTrafficPolicyTranslator) translateRequestMatcherHeaders(matchers []*zephyr_networking_types.TrafficPolicySpec_HeaderMatcher) (
	map[string]*istio_networking_types.StringMatch, map[string]*istio_networking_types.StringMatch,
) {
	headerMatchers := map[string]*istio_networking_types.StringMatch{}
	inverseHeaderMatchers := map[string]*istio_networking_types.StringMatch{}
	var matcherMap map[string]*istio_networking_types.StringMatch
	if matchers != nil {
		for _, matcher := range matchers {
			matcherMap = headerMatchers
			if matcher.GetInvertMatch() {
				matcherMap = inverseHeaderMatchers
			}
			if matcher.GetRegex() {
				matcherMap[matcher.GetName()] = &istio_networking_types.StringMatch{
					MatchType: &istio_networking_types.StringMatch_Regex{Regex: matcher.GetValue()},
				}
			} else {
				matcherMap[matcher.GetName()] = &istio_networking_types.StringMatch{
					MatchType: &istio_networking_types.StringMatch_Exact{Exact: matcher.GetValue()},
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

// ensure that subsets declared in this TrafficPolicy are reflected in the relevant kube Service's DestinationRules
// return name of Subset declared in DestinationRule
func (i *istioTrafficPolicyTranslator) translateSubset(
	ctx context.Context,
	destination *zephyr_networking_types.TrafficPolicySpec_MultiDestination_WeightedDestination,
) (string, error) {
	// fetch client for destination's cluster
	clusterName := destination.GetDestination().GetCluster()
	dynamicClient, err := i.dynamicClientGetter.GetClientForCluster(ctx, clusterName)
	if err != nil {
		return "", err
	}
	destinationRuleClient := i.destinationRuleClientFactory(dynamicClient)
	destinationRule, err := destinationRuleClient.GetDestinationRule(ctx, clients.ResourceRefToObjectKey(destination.GetDestination()))
	if err != nil {
		return "", err
	}
	for _, subset := range destinationRule.Spec.GetSubsets() {
		if labels.Equals(subset.GetLabels(), destination.GetSubset()) {
			return subset.GetName(), nil
		}
	}
	// subset doesn't yet exist, update the DestinationRule with it and return its generated name
	subsetName := generateUniqueSubsetName(destination.GetSubset())
	destinationRule.Spec.Subsets = append(destinationRule.Spec.Subsets, &istio_networking_types.Subset{
		Name:   subsetName,
		Labels: destination.GetSubset(),
	})
	err = destinationRuleClient.UpdateDestinationRule(ctx, destinationRule)
	if err != nil {
		return "", err
	}
	return subsetName, nil
}

// For each Destination, create an Istio HTTPRouteDestination
func (i *istioTrafficPolicyTranslator) translateDestinationRoutes(
	ctx context.Context,
	meshService *zephyr_discovery.MeshService,
	trafficPolicy *zephyr_networking.TrafficPolicy,
) ([]*istio_networking_types.HTTPRouteDestination, error) {
	var translatedRouteDestinations []*istio_networking_types.HTTPRouteDestination
	trafficShift := trafficPolicy.Spec.GetTrafficShift()
	if trafficShift != nil {
		for _, destination := range trafficShift.GetDestinations() {
			hostnameForKubeService, isMulticluster, err := i.getHostnameForKubeService(
				ctx,
				meshService,
				destination.GetDestination(),
			)
			if err != nil {
				return nil, err
			}
			httpRouteDestination := &istio_networking_types.HTTPRouteDestination{
				Destination: &istio_networking_types.Destination{
					Host: hostnameForKubeService,
				},
				Weight: int32(destination.GetWeight()),
			}
			// Add port to destination if non-zero
			// If the service backing this destination has one more than one port exposed, and
			// no port is chosen, istio will return an error. Otherwise istio will use the
			// one port available.
			if destination.Port != 0 {
				httpRouteDestination.Destination.Port = &istio_networking_types.PortSelector{
					Number: destination.Port,
				}
			}
			if destination.Subset != nil {
				// multicluster subsets are currently unsupported, so return a status error to invalidate the TrafficPolicy
				if isMulticluster {
					return nil, MultiClusterSubsetsNotSupported(destination)
				}
				subsetName, err := i.translateSubset(ctx, destination)
				if err != nil {
					return nil, err
				}
				httpRouteDestination.Destination.Subset = subsetName
			}
			translatedRouteDestinations = append(translatedRouteDestinations, httpRouteDestination)
		}
	} else {
		if len(meshService.Spec.GetKubeService().GetPorts()) != 1 {
			return nil, NoSpecifiedPortError(meshService)
		}
		// Since only one port is available, use that as the target port for the destination
		defaultServicePort := meshService.Spec.GetKubeService().GetPorts()[0]
		translatedRouteDestinations = []*istio_networking_types.HTTPRouteDestination{
			{
				Destination: &istio_networking_types.Destination{
					Host: buildServiceHostname(meshService),
					Port: &istio_networking_types.PortSelector{
						Number: defaultServicePort.Port,
					},
				},
			},
		}
	}
	return translatedRouteDestinations, nil
}

func (i *istioTrafficPolicyTranslator) translateRetries(trafficPolicy *zephyr_networking.TrafficPolicy) *istio_networking_types.HTTPRetry {
	var translatedRetries *istio_networking_types.HTTPRetry
	retries := trafficPolicy.Spec.GetRetries()
	if retries != nil {
		translatedRetries = &istio_networking_types.HTTPRetry{
			Attempts:      retries.GetAttempts(),
			PerTryTimeout: retries.GetPerTryTimeout(),
		}
	}
	return translatedRetries
}

func (i *istioTrafficPolicyTranslator) translateFaultInjection(trafficPolicy *zephyr_networking.TrafficPolicy) (*istio_networking_types.HTTPFaultInjection, error) {
	var translatedFaultInjection *istio_networking_types.HTTPFaultInjection
	faultInjection := trafficPolicy.Spec.GetFaultInjection()
	if faultInjection != nil {
		switch injectionType := faultInjection.GetFaultInjectionType().(type) {
		case *zephyr_networking_types.TrafficPolicySpec_FaultInjection_Abort_:
			abort := faultInjection.GetAbort()
			switch abortType := abort.GetErrorType().(type) {
			case *zephyr_networking_types.TrafficPolicySpec_FaultInjection_Abort_HttpStatus:
				translatedFaultInjection = &istio_networking_types.HTTPFaultInjection{
					Abort: &istio_networking_types.HTTPFaultInjection_Abort{
						ErrorType:  &istio_networking_types.HTTPFaultInjection_Abort_HttpStatus{HttpStatus: abort.GetHttpStatus()},
						Percentage: &istio_networking_types.Percent{Value: faultInjection.GetPercentage()},
					}}
			default:
				return nil, eris.Errorf("Abort.ErrorType has unexpected type %T", abortType)
			}
		case *zephyr_networking_types.TrafficPolicySpec_FaultInjection_Delay_:
			delay := faultInjection.GetDelay()
			switch delayType := delay.GetHttpDelayType().(type) {
			case *zephyr_networking_types.TrafficPolicySpec_FaultInjection_Delay_FixedDelay:
				translatedFaultInjection = &istio_networking_types.HTTPFaultInjection{
					Delay: &istio_networking_types.HTTPFaultInjection_Delay{
						HttpDelayType: &istio_networking_types.HTTPFaultInjection_Delay_FixedDelay{FixedDelay: delay.GetFixedDelay()},
						Percentage:    &istio_networking_types.Percent{Value: faultInjection.GetPercentage()},
					}}
			case *zephyr_networking_types.TrafficPolicySpec_FaultInjection_Delay_ExponentialDelay:
				translatedFaultInjection = &istio_networking_types.HTTPFaultInjection{
					Delay: &istio_networking_types.HTTPFaultInjection_Delay{
						HttpDelayType: &istio_networking_types.HTTPFaultInjection_Delay_ExponentialDelay{ExponentialDelay: delay.GetExponentialDelay()},
						Percentage:    &istio_networking_types.Percent{Value: faultInjection.GetPercentage()},
					}}
			default:
				return nil, eris.Errorf("Delay.HTTPDelayType has unexpected type %T", delayType)
			}
		default:
			return nil, eris.Errorf("FaultInjection.FaultInjectionType has unexpected type %T", injectionType)
		}
	}
	return translatedFaultInjection, nil
}

func (i *istioTrafficPolicyTranslator) translateMirror(
	ctx context.Context,
	meshService *zephyr_discovery.MeshService,
	trafficPolicy *zephyr_networking.TrafficPolicy,
) (*istio_networking_types.Destination, error) {
	var mirror *istio_networking_types.Destination
	if trafficPolicy.Spec.GetMirror() != nil {
		hostnameForKubeService, _, err := i.getHostnameForKubeService(
			ctx,
			meshService,
			trafficPolicy.Spec.GetMirror().GetDestination(),
		)
		if err != nil {
			return nil, err
		}
		mirror = &istio_networking_types.Destination{
			Host: hostnameForKubeService,
		}
		// Add port to destination if non-zero
		// If the service backing this destination has one more than one port exposed, and
		// no port is chosen, istio will return an error. Otherwise istio will use the
		// one port available.
		if trafficPolicy.Spec.GetMirror().GetPort() != 0 {
			mirror.Port = &istio_networking_types.PortSelector{
				Number: trafficPolicy.Spec.GetMirror().GetPort(),
			}
		}
	}
	return mirror, nil
}

func (i *istioTrafficPolicyTranslator) translateHeaderManipulation(trafficPolicy *zephyr_networking.TrafficPolicy) *istio_networking_types.Headers {
	var translatedHeaderManipulation *istio_networking_types.Headers
	headerManipulation := trafficPolicy.Spec.GetHeaderManipulation()
	if headerManipulation != nil {
		translatedHeaderManipulation = &istio_networking_types.Headers{
			Request: &istio_networking_types.Headers_HeaderOperations{
				Add:    headerManipulation.GetAppendRequestHeaders(),
				Remove: headerManipulation.GetRemoveRequestHeaders(),
			},
			Response: &istio_networking_types.Headers_HeaderOperations{
				Add:    headerManipulation.GetAppendResponseHeaders(),
				Remove: headerManipulation.GetRemoveResponseHeaders(),
			},
		}
	}
	return translatedHeaderManipulation
}

func (i *istioTrafficPolicyTranslator) translateCorsPolicy(trafficPolicy *zephyr_networking.TrafficPolicy) (*istio_networking_types.CorsPolicy, error) {
	var translatedCorsPolicy *istio_networking_types.CorsPolicy
	corsPolicy := trafficPolicy.Spec.GetCorsPolicy()
	if corsPolicy != nil {
		var allowOrigins []*istio_networking_types.StringMatch
		for i, allowOrigin := range corsPolicy.GetAllowOrigins() {
			var stringMatch *istio_networking_types.StringMatch
			switch matchType := allowOrigin.GetMatchType().(type) {
			case *zephyr_networking_types.TrafficPolicySpec_StringMatch_Exact:
				stringMatch = &istio_networking_types.StringMatch{MatchType: &istio_networking_types.StringMatch_Exact{Exact: allowOrigin.GetExact()}}
			case *zephyr_networking_types.TrafficPolicySpec_StringMatch_Prefix:
				stringMatch = &istio_networking_types.StringMatch{MatchType: &istio_networking_types.StringMatch_Prefix{Prefix: allowOrigin.GetPrefix()}}
			case *zephyr_networking_types.TrafficPolicySpec_StringMatch_Regex:
				stringMatch = &istio_networking_types.StringMatch{MatchType: &istio_networking_types.StringMatch_Regex{Regex: allowOrigin.GetRegex()}}
			default:
				return nil, eris.Errorf("AllowOrigins[%d].MatchType has unexpected type %T", i, matchType)
			}
			allowOrigins = append(allowOrigins, stringMatch)
		}
		translatedCorsPolicy = &istio_networking_types.CorsPolicy{
			AllowOrigins:     allowOrigins,
			AllowMethods:     corsPolicy.GetAllowMethods(),
			AllowHeaders:     corsPolicy.GetAllowHeaders(),
			ExposeHeaders:    corsPolicy.GetExposeHeaders(),
			MaxAge:           corsPolicy.GetMaxAge(),
			AllowCredentials: corsPolicy.GetAllowCredentials(),
		}
	}
	return translatedCorsPolicy, nil
}

func (i *istioTrafficPolicyTranslator) errorToStatus(err error) *zephyr_networking_types.TrafficPolicyStatus_TranslatorError {
	return &zephyr_networking_types.TrafficPolicyStatus_TranslatorError{
		TranslatorId: TranslatorId,
		ErrorMessage: err.Error(),
	}
}

func (i *istioTrafficPolicyTranslator) getClusterNameForMeshService(
	ctx context.Context,
	meshService *zephyr_discovery.MeshService,
) (string, error) {
	mesh, err := i.meshClient.GetMesh(ctx, clients.ResourceRefToObjectKey(meshService.Spec.GetMesh()))
	if err != nil {
		return "", err
	}
	return mesh.Spec.GetCluster().GetName(), nil
}

// If destination is in the same namespace as k8s Service, return k8s Service name.namespace
// Else, return k8s Service multicluster DNS name
func (i *istioTrafficPolicyTranslator) getHostnameForKubeService(
	ctx context.Context,
	meshService *zephyr_discovery.MeshService,
	destination *zephyr_core_types.ResourceRef,
) (hostname string, isMulticluster bool, err error) {
	destinationMeshService, err := i.resourceSelector.GetMeshServiceByRefSelector(
		ctx, destination.GetName(), destination.GetNamespace(), destination.GetCluster())
	if err != nil {
		return "", false, err
	}
	if destination.GetCluster() == meshService.Spec.GetKubeService().GetRef().GetCluster() {
		// destination is on the same cluster as the MeshService's k8s Service
		return buildServiceHostname(destinationMeshService), false, nil
	} else {
		// destination is on a remote cluster to the MeshService's k8s Service
		return destinationMeshService.Spec.GetFederation().GetMulticlusterDnsName(), true, nil
	}
}

func buildServiceHostname(meshService *zephyr_discovery.MeshService) string {
	return meshService.Spec.GetKubeService().GetRef().GetName()
}

// sort the label keys, then in order concatenate keys-values
func generateUniqueSubsetName(selectors map[string]string) string {
	var keys []string
	for key, val := range selectors {
		keys = append(keys, key+"-"+val)
	}
	sort.Strings(keys)
	return strings.Join(keys, "_")
}
