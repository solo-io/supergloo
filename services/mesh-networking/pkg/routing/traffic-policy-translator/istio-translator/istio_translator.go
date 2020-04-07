package istio_translator

import (
	"context"
	"sort"
	"strings"

	"github.com/rotisserie/eris"
	core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	discovery_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	networking_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/clients"
	istio_networking "github.com/solo-io/service-mesh-hub/pkg/clients/istio/networking"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/clients/zephyr/discovery"
	"github.com/solo-io/service-mesh-hub/pkg/selector"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/multicluster/manager"
	traffic_policy_translator "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/routing/traffic-policy-translator"
	api_v1alpha3 "istio.io/api/networking/v1alpha3"
	client_v1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
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
	NoSpecifiedPortError = func(svc *discovery_v1alpha1.MeshService) error {
		return eris.Errorf("Mesh service %s.%s ports list does not include just one entry, so no default can be used. "+
			"Must specify a destination with a port", svc.Name, svc.Namespace)
	}
	MultiClusterSubsetsNotSupported = func(dest *networking_types.TrafficPolicySpec_MultiDestination_WeightedDestination) error {
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
	meshService *discovery_v1alpha1.MeshService,
	mesh *discovery_v1alpha1.Mesh,
	mergedTrafficPolicies []*networking_v1alpha1.TrafficPolicy,
) *networking_types.TrafficPolicyStatus_TranslatorError {
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
	meshService *discovery_v1alpha1.MeshService,
) (istio_networking.DestinationRuleClient, istio_networking.VirtualServiceClient, error) {
	clusterName, err := i.getClusterNameForMeshService(ctx, meshService)
	if err != nil {
		return nil, nil, err
	}
	dynamicClient, err := i.dynamicClientGetter.GetClientForCluster(clusterName)
	if err != nil {
		return nil, nil, err
	}
	return i.destinationRuleClientFactory(dynamicClient), i.virtualServiceClientFactory(dynamicClient), nil
}

func (i *istioTrafficPolicyTranslator) ensureDestinationRule(
	ctx context.Context,
	meshService *discovery_v1alpha1.MeshService,
	destinationRuleClient istio_networking.DestinationRuleClient,
) *networking_types.TrafficPolicyStatus_TranslatorError {
	destinationRule := &client_v1alpha3.DestinationRule{
		ObjectMeta: clients.ResourceRefToObjectMeta(meshService.Spec.GetKubeService().GetRef()),
		Spec: api_v1alpha3.DestinationRule{
			Host: buildServiceHostname(meshService),
			TrafficPolicy: &api_v1alpha3.TrafficPolicy{
				Tls: &api_v1alpha3.TLSSettings{
					Mode: api_v1alpha3.TLSSettings_ISTIO_MUTUAL,
				},
			},
		},
	}
	// Only attempt to create if does not already exist
	err := destinationRuleClient.Create(ctx, destinationRule)
	if err != nil && !errors.IsAlreadyExists(err) {
		return i.errorToStatus(err)
	}
	return nil
}

func (i *istioTrafficPolicyTranslator) ensureVirtualService(
	ctx context.Context,
	meshService *discovery_v1alpha1.MeshService,
	mergedTrafficPolicies []*networking_v1alpha1.TrafficPolicy,
	virtualServiceClient istio_networking.VirtualServiceClient,
) *networking_types.TrafficPolicyStatus_TranslatorError {
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
	err = virtualServiceClient.UpsertSpec(ctx, computedVirtualService)
	if err != nil {
		return i.errorToStatus(err)
	}
	return nil
}

func (i *istioTrafficPolicyTranslator) translateIntoVirtualService(
	ctx context.Context,
	meshService *discovery_v1alpha1.MeshService,
	trafficPolicies []*networking_v1alpha1.TrafficPolicy,
) (*client_v1alpha3.VirtualService, error) {
	virtualService := &client_v1alpha3.VirtualService{
		ObjectMeta: clients.ResourceRefToObjectMeta(meshService.Spec.GetKubeService().GetRef()),
		Spec: api_v1alpha3.VirtualService{
			Hosts: []string{buildServiceHostname(meshService)},
		},
	}
	var allHttpRoutes []*api_v1alpha3.HTTPRoute
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
	meshService *discovery_v1alpha1.MeshService,
	trafficPolicy *networking_v1alpha1.TrafficPolicy,
) ([]*api_v1alpha3.HTTPRoute, error) {
	var err error
	var faultInjection *api_v1alpha3.HTTPFaultInjection
	var corsPolicy *api_v1alpha3.CorsPolicy
	var requestMatchers []*api_v1alpha3.HTTPMatchRequest
	var mirrorPercentage *api_v1alpha3.Percent
	var mirror *api_v1alpha3.Destination
	var trafficShift []*api_v1alpha3.HTTPRouteDestination
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
		mirrorPercentage = &api_v1alpha3.Percent{Value: trafficPolicy.Spec.GetMirror().GetPercentage()}
	}
	if mirror, err = i.translateMirror(ctx, meshService, trafficPolicy); err != nil {
		return nil, err
	}
	if trafficShift, err = i.translateDestinationRoutes(ctx, meshService, trafficPolicy); err != nil {
		return nil, err
	}
	retries := i.translateRetries(trafficPolicy)
	headerManipulation := i.translateHeaderManipulation(trafficPolicy)
	var httpRoutes []*api_v1alpha3.HTTPRoute

	if len(requestMatchers) == 0 {
		// If no matchers are present return a single route with no matchers
		httpRoutes = append(httpRoutes, &api_v1alpha3.HTTPRoute{
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
		httpRoutes = make([]*api_v1alpha3.HTTPRoute, 0, len(requestMatchers))
		// flatten HTTPMatchRequests, i.e. create an HTTPRoute per HTTPMatchRequest
		// this facilitates sorting the HTTPRoutes to produce a well-defined ordering of precedence
		for _, requestMatcher := range requestMatchers {
			httpRoutes = append(httpRoutes, &api_v1alpha3.HTTPRoute{
				Match:            []*api_v1alpha3.HTTPMatchRequest{requestMatcher},
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
	trafficPolicy *networking_v1alpha1.TrafficPolicy,
) ([]*api_v1alpha3.HTTPMatchRequest, error) {
	// Generate HttpMatchRequests for SourceSelector, one per namespace.
	var sourceMatchers []*api_v1alpha3.HTTPMatchRequest
	// Set SourceNamespace and SourceLabels.
	if len(trafficPolicy.Spec.GetSourceSelector().GetLabels()) > 0 ||
		len(trafficPolicy.Spec.GetSourceSelector().GetNamespaces()) > 0 {
		if len(trafficPolicy.Spec.GetSourceSelector().GetNamespaces()) > 0 {
			for _, namespace := range trafficPolicy.Spec.GetSourceSelector().GetNamespaces() {
				matchRequest := &api_v1alpha3.HTTPMatchRequest{
					SourceNamespace: namespace,
					SourceLabels:    trafficPolicy.Spec.GetSourceSelector().GetLabels(),
				}
				sourceMatchers = append(sourceMatchers, matchRequest)
			}
		} else {
			sourceMatchers = append(sourceMatchers, &api_v1alpha3.HTTPMatchRequest{
				SourceLabels: trafficPolicy.Spec.GetSourceSelector().GetLabels(),
			})
		}
	}
	if trafficPolicy.Spec.GetHttpRequestMatchers() == nil {
		return sourceMatchers, nil
	}
	// If HttpRequestMatchers exist, generate cartesian product of sourceMatchers and httpRequestMatchers.
	var translatedRequestMatchers []*api_v1alpha3.HTTPMatchRequest
	// If SourceSelector is nil, generate an HttpMatchRequest without SourceSelector match criteria
	if len(sourceMatchers) == 0 {
		sourceMatchers = append(sourceMatchers, &api_v1alpha3.HTTPMatchRequest{})
	}
	// Set QueryParams, Headers, WithoutHeaders, Uri, and Method.
	for _, sourceMatcher := range sourceMatchers {
		for _, matcher := range trafficPolicy.Spec.GetHttpRequestMatchers() {
			httpMatcher := &api_v1alpha3.HTTPMatchRequest{
				SourceNamespace: sourceMatcher.GetSourceNamespace(),
				SourceLabels:    sourceMatcher.GetSourceLabels(),
			}
			headerMatchers, inverseHeaderMatchers := i.translateRequestMatcherHeaders(matcher.GetHeaders())
			uriMatcher, err := i.translateRequestMatcherPathSpecifier(matcher)
			if err != nil {
				return nil, err
			}
			var method *api_v1alpha3.StringMatch
			if matcher.GetMethod() != nil {
				method = &api_v1alpha3.StringMatch{MatchType: &api_v1alpha3.StringMatch_Exact{Exact: matcher.GetMethod().GetMethod().String()}}
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

func (i *istioTrafficPolicyTranslator) translateRequestMatcherPathSpecifier(matcher *networking_types.TrafficPolicySpec_HttpMatcher) (*api_v1alpha3.StringMatch, error) {
	if matcher != nil && matcher.GetPathSpecifier() != nil {
		switch pathSpecifierType := matcher.GetPathSpecifier().(type) {
		case *networking_types.TrafficPolicySpec_HttpMatcher_Exact:
			return &api_v1alpha3.StringMatch{MatchType: &api_v1alpha3.StringMatch_Exact{Exact: matcher.GetExact()}}, nil
		case *networking_types.TrafficPolicySpec_HttpMatcher_Prefix:
			return &api_v1alpha3.StringMatch{MatchType: &api_v1alpha3.StringMatch_Prefix{Prefix: matcher.GetPrefix()}}, nil
		case *networking_types.TrafficPolicySpec_HttpMatcher_Regex:
			return &api_v1alpha3.StringMatch{MatchType: &api_v1alpha3.StringMatch_Regex{Regex: matcher.GetRegex()}}, nil
		default:
			return nil, eris.Errorf("RequestMatchers[].PathSpecifier has unexpected type %T", pathSpecifierType)
		}
	}
	return nil, nil
}

func (i *istioTrafficPolicyTranslator) translateRequestMatcherQueryParams(matchers []*networking_types.TrafficPolicySpec_QueryParameterMatcher) map[string]*api_v1alpha3.StringMatch {
	var translatedQueryParamMatcher map[string]*api_v1alpha3.StringMatch
	if matchers != nil {
		translatedQueryParamMatcher = map[string]*api_v1alpha3.StringMatch{}
		for _, matcher := range matchers {
			if matcher.GetRegex() {
				translatedQueryParamMatcher[matcher.GetName()] = &api_v1alpha3.StringMatch{
					MatchType: &api_v1alpha3.StringMatch_Regex{Regex: matcher.GetValue()},
				}
			} else {
				translatedQueryParamMatcher[matcher.GetName()] = &api_v1alpha3.StringMatch{
					MatchType: &api_v1alpha3.StringMatch_Exact{Exact: matcher.GetValue()},
				}
			}
		}
	}
	return translatedQueryParamMatcher
}

func (i *istioTrafficPolicyTranslator) translateRequestMatcherHeaders(matchers []*networking_types.TrafficPolicySpec_HeaderMatcher) (
	map[string]*api_v1alpha3.StringMatch, map[string]*api_v1alpha3.StringMatch,
) {
	headerMatchers := map[string]*api_v1alpha3.StringMatch{}
	inverseHeaderMatchers := map[string]*api_v1alpha3.StringMatch{}
	var matcherMap map[string]*api_v1alpha3.StringMatch
	if matchers != nil {
		for _, matcher := range matchers {
			matcherMap = headerMatchers
			if matcher.GetInvertMatch() {
				matcherMap = inverseHeaderMatchers
			}
			if matcher.GetRegex() {
				matcherMap[matcher.GetName()] = &api_v1alpha3.StringMatch{
					MatchType: &api_v1alpha3.StringMatch_Regex{Regex: matcher.GetValue()},
				}
			} else {
				matcherMap[matcher.GetName()] = &api_v1alpha3.StringMatch{
					MatchType: &api_v1alpha3.StringMatch_Exact{Exact: matcher.GetValue()},
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
	destination *networking_types.TrafficPolicySpec_MultiDestination_WeightedDestination,
) (string, error) {
	// fetch client for destination's cluster
	clusterName := destination.GetDestination().GetCluster()
	dynamicClient, err := i.dynamicClientGetter.GetClientForCluster(clusterName)
	if err != nil {
		return "", err
	}
	destinationRuleClient := i.destinationRuleClientFactory(dynamicClient)
	destinationRule, err := destinationRuleClient.Get(ctx, clients.ResourceRefToObjectKey(destination.GetDestination()))
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
	destinationRule.Spec.Subsets = append(destinationRule.Spec.Subsets, &api_v1alpha3.Subset{
		Name:   subsetName,
		Labels: destination.GetSubset(),
	})
	err = destinationRuleClient.Update(ctx, destinationRule)
	if err != nil {
		return "", err
	}
	return subsetName, nil
}

// For each Destination, create an Istio HTTPRouteDestination
func (i *istioTrafficPolicyTranslator) translateDestinationRoutes(
	ctx context.Context,
	meshService *discovery_v1alpha1.MeshService,
	trafficPolicy *networking_v1alpha1.TrafficPolicy,
) ([]*api_v1alpha3.HTTPRouteDestination, error) {
	var translatedRouteDestinations []*api_v1alpha3.HTTPRouteDestination
	trafficShift := trafficPolicy.Spec.GetTrafficShift()
	if trafficShift != nil {
		translatedRouteDestinations = []*api_v1alpha3.HTTPRouteDestination{}
		for _, destination := range trafficShift.GetDestinations() {
			hostnameForKubeService, isMulticluster, err := i.getHostnameForKubeService(
				ctx,
				meshService,
				destination.GetDestination(),
			)
			if err != nil {
				return nil, err
			}
			httpRouteDestination := &api_v1alpha3.HTTPRouteDestination{
				Destination: &api_v1alpha3.Destination{
					Host: hostnameForKubeService,
				},
				Weight: int32(destination.GetWeight()),
			}
			// Add port to destination if non-zero
			// If the service backing this destination has one more than one port exposed, and
			// no port is chosen, istio will return an error. Otherwise istio will use the
			// one port available.
			if destination.Port != 0 {
				httpRouteDestination.Destination.Port = &api_v1alpha3.PortSelector{
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
		translatedRouteDestinations = []*api_v1alpha3.HTTPRouteDestination{
			{
				Destination: &api_v1alpha3.Destination{
					Host: buildServiceHostname(meshService),
					Port: &api_v1alpha3.PortSelector{
						Number: defaultServicePort.Port,
					},
				},
			},
		}
	}
	return translatedRouteDestinations, nil
}

func (i *istioTrafficPolicyTranslator) translateRetries(trafficPolicy *networking_v1alpha1.TrafficPolicy) *api_v1alpha3.HTTPRetry {
	var translatedRetries *api_v1alpha3.HTTPRetry
	retries := trafficPolicy.Spec.GetRetries()
	if retries != nil {
		translatedRetries = &api_v1alpha3.HTTPRetry{
			Attempts:      retries.GetAttempts(),
			PerTryTimeout: retries.GetPerTryTimeout(),
		}
	}
	return translatedRetries
}

func (i *istioTrafficPolicyTranslator) translateFaultInjection(trafficPolicy *networking_v1alpha1.TrafficPolicy) (*api_v1alpha3.HTTPFaultInjection, error) {
	var translatedFaultInjection *api_v1alpha3.HTTPFaultInjection
	faultInjection := trafficPolicy.Spec.GetFaultInjection()
	if faultInjection != nil {
		switch injectionType := faultInjection.GetFaultInjectionType().(type) {
		case *networking_types.TrafficPolicySpec_FaultInjection_Abort_:
			abort := faultInjection.GetAbort()
			switch abortType := abort.GetErrorType().(type) {
			case *networking_types.TrafficPolicySpec_FaultInjection_Abort_HttpStatus:
				translatedFaultInjection = &api_v1alpha3.HTTPFaultInjection{
					Abort: &api_v1alpha3.HTTPFaultInjection_Abort{
						ErrorType:  &api_v1alpha3.HTTPFaultInjection_Abort_HttpStatus{HttpStatus: abort.GetHttpStatus()},
						Percentage: &api_v1alpha3.Percent{Value: faultInjection.GetPercentage()},
					}}
			default:
				return nil, eris.Errorf("Abort.ErrorType has unexpected type %T", abortType)
			}
		case *networking_types.TrafficPolicySpec_FaultInjection_Delay_:
			delay := faultInjection.GetDelay()
			switch delayType := delay.GetHttpDelayType().(type) {
			case *networking_types.TrafficPolicySpec_FaultInjection_Delay_FixedDelay:
				translatedFaultInjection = &api_v1alpha3.HTTPFaultInjection{
					Delay: &api_v1alpha3.HTTPFaultInjection_Delay{
						HttpDelayType: &api_v1alpha3.HTTPFaultInjection_Delay_FixedDelay{FixedDelay: delay.GetFixedDelay()},
						Percentage:    &api_v1alpha3.Percent{Value: faultInjection.GetPercentage()},
					}}
			case *networking_types.TrafficPolicySpec_FaultInjection_Delay_ExponentialDelay:
				translatedFaultInjection = &api_v1alpha3.HTTPFaultInjection{
					Delay: &api_v1alpha3.HTTPFaultInjection_Delay{
						HttpDelayType: &api_v1alpha3.HTTPFaultInjection_Delay_ExponentialDelay{ExponentialDelay: delay.GetExponentialDelay()},
						Percentage:    &api_v1alpha3.Percent{Value: faultInjection.GetPercentage()},
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
	meshService *discovery_v1alpha1.MeshService,
	trafficPolicy *networking_v1alpha1.TrafficPolicy,
) (*api_v1alpha3.Destination, error) {
	var mirror *api_v1alpha3.Destination
	if trafficPolicy.Spec.GetMirror() != nil {
		hostnameForKubeService, _, err := i.getHostnameForKubeService(
			ctx,
			meshService,
			trafficPolicy.Spec.GetMirror().GetDestination(),
		)
		if err != nil {
			return nil, err
		}
		mirror = &api_v1alpha3.Destination{
			Host: hostnameForKubeService,
		}
		// Add port to destination if non-zero
		// If the service backing this destination has one more than one port exposed, and
		// no port is chosen, istio will return an error. Otherwise istio will use the
		// one port available.
		if trafficPolicy.Spec.GetMirror().GetPort() != 0 {
			mirror.Port = &api_v1alpha3.PortSelector{
				Number: trafficPolicy.Spec.GetMirror().GetPort(),
			}
		}
	}
	return mirror, nil
}

func (i *istioTrafficPolicyTranslator) translateHeaderManipulation(trafficPolicy *networking_v1alpha1.TrafficPolicy) *api_v1alpha3.Headers {
	var translatedHeaderManipulation *api_v1alpha3.Headers
	headerManipulation := trafficPolicy.Spec.GetHeaderManipulation()
	if headerManipulation != nil {
		translatedHeaderManipulation = &api_v1alpha3.Headers{
			Request: &api_v1alpha3.Headers_HeaderOperations{
				Add:    headerManipulation.GetAppendRequestHeaders(),
				Remove: headerManipulation.GetRemoveRequestHeaders(),
			},
			Response: &api_v1alpha3.Headers_HeaderOperations{
				Add:    headerManipulation.GetAppendResponseHeaders(),
				Remove: headerManipulation.GetRemoveResponseHeaders(),
			},
		}
	}
	return translatedHeaderManipulation
}

func (i *istioTrafficPolicyTranslator) translateCorsPolicy(trafficPolicy *networking_v1alpha1.TrafficPolicy) (*api_v1alpha3.CorsPolicy, error) {
	var translatedCorsPolicy *api_v1alpha3.CorsPolicy
	corsPolicy := trafficPolicy.Spec.GetCorsPolicy()
	if corsPolicy != nil {
		var allowOrigins []*api_v1alpha3.StringMatch
		for i, allowOrigin := range corsPolicy.GetAllowOrigins() {
			var stringMatch *api_v1alpha3.StringMatch
			switch matchType := allowOrigin.GetMatchType().(type) {
			case *networking_types.TrafficPolicySpec_StringMatch_Exact:
				stringMatch = &api_v1alpha3.StringMatch{MatchType: &api_v1alpha3.StringMatch_Exact{Exact: allowOrigin.GetExact()}}
			case *networking_types.TrafficPolicySpec_StringMatch_Prefix:
				stringMatch = &api_v1alpha3.StringMatch{MatchType: &api_v1alpha3.StringMatch_Prefix{Prefix: allowOrigin.GetPrefix()}}
			case *networking_types.TrafficPolicySpec_StringMatch_Regex:
				stringMatch = &api_v1alpha3.StringMatch{MatchType: &api_v1alpha3.StringMatch_Regex{Regex: allowOrigin.GetRegex()}}
			default:
				return nil, eris.Errorf("AllowOrigins[%d].MatchType has unexpected type %T", i, matchType)
			}
			allowOrigins = append(allowOrigins, stringMatch)
		}
		translatedCorsPolicy = &api_v1alpha3.CorsPolicy{
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

func (i *istioTrafficPolicyTranslator) errorToStatus(err error) *networking_types.TrafficPolicyStatus_TranslatorError {
	return &networking_types.TrafficPolicyStatus_TranslatorError{
		TranslatorId: TranslatorId,
		ErrorMessage: err.Error(),
	}
}

func (i *istioTrafficPolicyTranslator) getClusterNameForMeshService(
	ctx context.Context,
	meshService *discovery_v1alpha1.MeshService,
) (string, error) {
	mesh, err := i.meshClient.Get(ctx, clients.ResourceRefToObjectKey(meshService.Spec.GetMesh()))
	if err != nil {
		return "", err
	}
	return mesh.Spec.GetCluster().GetName(), nil
}

// If destination is in the same namespace as k8s Service, return k8s Service name.namespace
// Else, return k8s Service multicluster DNS name
func (i *istioTrafficPolicyTranslator) getHostnameForKubeService(
	ctx context.Context,
	meshService *discovery_v1alpha1.MeshService,
	destination *core_types.ResourceRef,
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

func buildServiceHostname(meshService *discovery_v1alpha1.MeshService) string {
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
