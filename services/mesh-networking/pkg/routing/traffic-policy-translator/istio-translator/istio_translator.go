package istio_translator

import (
	"context"

	"github.com/rotisserie/eris"
	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	discovery_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	networking_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	networking_types "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	istio_networking "github.com/solo-io/mesh-projects/pkg/clients/istio/networking"
	zephyr_discovery "github.com/solo-io/mesh-projects/pkg/clients/zephyr/discovery"
	"github.com/solo-io/mesh-projects/services/common"
	mc_manager "github.com/solo-io/mesh-projects/services/common/multicluster/manager"
	traffic_policy_translator "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/routing/traffic-policy-translator"
	api_v1beta1 "istio.io/api/networking/v1beta1"
	client_v1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	TranslatorId = "istio-translator"
)

type IstioTranslator traffic_policy_translator.TrafficPolicyMeshTranslator

func DefaultIstioTrafficPolicyTranslator(
	dynamicClientGetter mc_manager.DynamicClientGetter,
	meshClient zephyr_discovery.MeshClient,
	meshServiceClient zephyr_discovery.MeshServiceClient,
	virtualServiceClientFactory istio_networking.VirtualServiceClientFactory,
) IstioTranslator {
	return &IstioTrafficPolicyTranslator{
		dynamicClientGetter:         dynamicClientGetter,
		meshClient:                  meshClient,
		meshServiceClient:           meshServiceClient,
		virtualServiceClientFactory: virtualServiceClientFactory,
	}
}

// visible for testing
func NewIstioTrafficPolicyTranslator(
	dynamicClientGetter mc_manager.DynamicClientGetter,
	meshClient zephyr_discovery.MeshClient,
	meshServiceClient zephyr_discovery.MeshServiceClient,
	virtualServiceClientFactory istio_networking.VirtualServiceClientFactory,
) *IstioTrafficPolicyTranslator {
	return &IstioTrafficPolicyTranslator{
		dynamicClientGetter:         dynamicClientGetter,
		meshClient:                  meshClient,
		meshServiceClient:           meshServiceClient,
		virtualServiceClientFactory: virtualServiceClientFactory,
	}
}

// visible for testing
type IstioTrafficPolicyTranslator struct {
	dynamicClientGetter         mc_manager.DynamicClientGetter
	meshClient                  zephyr_discovery.MeshClient
	meshServiceClient           zephyr_discovery.MeshServiceClient
	virtualServiceClientFactory istio_networking.VirtualServiceClientFactory
}

func (i *IstioTrafficPolicyTranslator) TranslateTrafficPolicy(
	ctx context.Context,
	meshService *discovery_v1alpha1.MeshService,
	mesh *discovery_v1alpha1.Mesh,
	mergedTrafficPolicies []*networking_v1alpha1.TrafficPolicy,
) *networking_types.TrafficPolicyStatus_TranslatorError {
	if mesh.Spec.GetIstio() == nil {
		return nil
	}
	computedVirtualService, err := i.TranslateIntoVirtualService(meshService, mergedTrafficPolicies)
	if err != nil {
		return i.errorToStatus(err)
	}
	// Upsert computed VirtualService
	virtualServiceClient, err := i.fetchVirtualServiceClientForMeshService(ctx, meshService)
	if err != nil {
		return i.errorToStatus(err)
	}
	err = virtualServiceClient.Upsert(ctx, computedVirtualService)
	if err != nil {
		return i.errorToStatus(err)
	}
	return nil
}

func (i *IstioTrafficPolicyTranslator) fetchVirtualServiceClientForMeshService(
	ctx context.Context,
	meshService *discovery_v1alpha1.MeshService,
) (istio_networking.VirtualServiceClient, error) {
	clusterName, err := i.getClusterNameForMeshService(ctx, meshService)
	if err != nil {
		return nil, err
	}
	dynamicClient, ok := i.dynamicClientGetter.GetClientForCluster(clusterName)
	if !ok {
		return nil, eris.Errorf("Client not found for cluster with name: %s", clusterName)
	}
	return i.virtualServiceClientFactory(dynamicClient), nil
}

func (i *IstioTrafficPolicyTranslator) buildObjectMeta(meshService *discovery_v1alpha1.MeshService) v1.ObjectMeta {
	return v1.ObjectMeta{
		Name:      meshService.Spec.GetKubeService().GetRef().GetName(),
		Namespace: meshService.Spec.GetKubeService().GetRef().GetNamespace(),
	}
}

func (i *IstioTrafficPolicyTranslator) TranslateIntoVirtualService(
	meshService *discovery_v1alpha1.MeshService,
	trafficPolicies []*networking_v1alpha1.TrafficPolicy,
) (*client_v1beta1.VirtualService, error) {
	virtualService := &client_v1beta1.VirtualService{
		ObjectMeta: i.buildObjectMeta(meshService),
		Spec: api_v1beta1.VirtualService{
			Hosts: []string{meshService.GetName() + "." + meshService.GetNamespace()},
			Http:  []*api_v1beta1.HTTPRoute{},
		},
	}
	for _, trafficPolicy := range trafficPolicies {
		fault, err := i.TranslateFaultInjection(trafficPolicy)
		if err != nil {
			return nil, err
		}
		corsPolicy, err := i.TranslateCorsPolicy(trafficPolicy)
		if err != nil {
			return nil, err
		}
		requestMatchers, err := i.TranslateRequestMatchers(trafficPolicy)
		if err != nil {
			return nil, err
		}
		var mirrorPercentage *api_v1beta1.Percent
		if trafficPolicy.Spec.GetMirror() != nil {
			mirrorPercentage = &api_v1beta1.Percent{Value: trafficPolicy.Spec.GetMirror().GetPercentage()}
		}
		mirror, err := i.TranslateMirror(meshService, trafficPolicy)
		if err != nil {
			return nil, err
		}
		trafficShift, err := i.TranslateTrafficShift(meshService, trafficPolicy)
		if err != nil {
			return nil, err
		}
		virtualService.Spec.Http = append(virtualService.Spec.GetHttp(), &api_v1beta1.HTTPRoute{
			Match:            requestMatchers,
			Route:            trafficShift,
			Timeout:          trafficPolicy.Spec.GetRequestTimeout(),
			Fault:            fault,
			CorsPolicy:       corsPolicy,
			Retries:          i.TranslateRetries(trafficPolicy),
			MirrorPercentage: mirrorPercentage,
			Mirror:           mirror,
			Headers:          i.TranslateHeaderManipulation(trafficPolicy),
		})
	}
	return virtualService, nil
}

func (i *IstioTrafficPolicyTranslator) TranslateRequestMatchers(trafficPolicy *networking_v1alpha1.TrafficPolicy) ([]*api_v1beta1.HTTPMatchRequest, error) {
	var translatedRequestMatcher []*api_v1beta1.HTTPMatchRequest
	matchers := trafficPolicy.Spec.GetHttpRequestMatchers()
	if matchers != nil {
		translatedRequestMatcher = []*api_v1beta1.HTTPMatchRequest{}
		for _, matcher := range matchers {
			headerMatchers, inverseHeaderMatchers := i.TranslateRequestMatcherHeaders(matcher.GetHeaders())
			uriMatcher, err := i.TranslateRequestMatcherPathSpecifier(matcher)
			if err != nil {
				return nil, err
			}
			matchRequest := &api_v1beta1.HTTPMatchRequest{
				Uri:            uriMatcher,
				Method:         &api_v1beta1.StringMatch{MatchType: &api_v1beta1.StringMatch_Exact{Exact: matcher.GetMethod().String()}},
				QueryParams:    i.TranslateRequestMatcherQueryParams(matcher.GetQueryParameters()),
				Headers:        headerMatchers,
				WithoutHeaders: inverseHeaderMatchers,
			}
			translatedRequestMatcher = append(translatedRequestMatcher, matchRequest)
		}
	}
	return translatedRequestMatcher, nil
}

func (i *IstioTrafficPolicyTranslator) TranslateRequestMatcherPathSpecifier(matcher *networking_types.HttpMatcher) (*api_v1beta1.StringMatch, error) {
	if matcher != nil && matcher.GetPathSpecifier() != nil {
		switch pathSpecifierType := matcher.GetPathSpecifier().(type) {
		case *networking_types.HttpMatcher_Exact:
			return &api_v1beta1.StringMatch{MatchType: &api_v1beta1.StringMatch_Exact{Exact: matcher.GetExact()}}, nil
		case *networking_types.HttpMatcher_Prefix:
			return &api_v1beta1.StringMatch{MatchType: &api_v1beta1.StringMatch_Prefix{Prefix: matcher.GetPrefix()}}, nil
		case *networking_types.HttpMatcher_Regex:
			return &api_v1beta1.StringMatch{MatchType: &api_v1beta1.StringMatch_Regex{Regex: matcher.GetRegex()}}, nil
		default:
			return nil, eris.Errorf("RequestMatchers[].PathSpecifier has unexpected type %T", pathSpecifierType)
		}
	}
	return nil, nil
}

func (i *IstioTrafficPolicyTranslator) TranslateRequestMatcherQueryParams(matchers []*networking_types.QueryParameterMatcher) map[string]*api_v1beta1.StringMatch {
	var translatedQueryParamMatcher map[string]*api_v1beta1.StringMatch
	if matchers != nil {
		translatedQueryParamMatcher = map[string]*api_v1beta1.StringMatch{}
		for _, matcher := range matchers {
			if matcher.GetRegex() {
				translatedQueryParamMatcher[matcher.GetName()] = &api_v1beta1.StringMatch{
					MatchType: &api_v1beta1.StringMatch_Regex{Regex: matcher.GetValue()},
				}
			} else {
				translatedQueryParamMatcher[matcher.GetName()] = &api_v1beta1.StringMatch{
					MatchType: &api_v1beta1.StringMatch_Exact{Exact: matcher.GetValue()},
				}
			}
		}
	}
	return translatedQueryParamMatcher
}

func (i *IstioTrafficPolicyTranslator) TranslateRequestMatcherHeaders(matchers []*networking_types.HeaderMatcher) (
	map[string]*api_v1beta1.StringMatch, map[string]*api_v1beta1.StringMatch,
) {
	headerMatchers := map[string]*api_v1beta1.StringMatch{}
	inverseHeaderMatchers := map[string]*api_v1beta1.StringMatch{}
	var matcherMap map[string]*api_v1beta1.StringMatch
	if matchers != nil {
		for _, matcher := range matchers {
			matcherMap = headerMatchers
			if matcher.GetInvertMatch() {
				matcherMap = inverseHeaderMatchers
			}
			if matcher.GetRegex() {
				matcherMap[matcher.GetName()] = &api_v1beta1.StringMatch{
					MatchType: &api_v1beta1.StringMatch_Regex{Regex: matcher.GetValue()},
				}
			} else {
				matcherMap[matcher.GetName()] = &api_v1beta1.StringMatch{
					MatchType: &api_v1beta1.StringMatch_Exact{Exact: matcher.GetValue()},
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

// For each Destination, for each subset if any exist, create an Istio HTTPRouteDestination
func (i *IstioTrafficPolicyTranslator) TranslateTrafficShift(
	meshService *discovery_v1alpha1.MeshService,
	trafficPolicy *networking_v1alpha1.TrafficPolicy,
) ([]*api_v1beta1.HTTPRouteDestination, error) {
	var translatedTrafficShift []*api_v1beta1.HTTPRouteDestination
	trafficShift := trafficPolicy.Spec.GetTrafficShift()
	if trafficShift != nil {
		translatedTrafficShift = []*api_v1beta1.HTTPRouteDestination{}
		for _, destination := range trafficShift.GetDestinations() {
			if destination.GetSubset() != nil {
				//TODO: implement subsets
			} else {
				httpRouteDestination := &api_v1beta1.HTTPRouteDestination{
					Destination: &api_v1beta1.Destination{
						Host: getHostname(meshService, destination.GetDestination()),
					},
					Weight: int32(destination.GetWeight()),
				}
				translatedTrafficShift = append(translatedTrafficShift, httpRouteDestination)
			}
		}
	}
	return translatedTrafficShift, nil
}

func (i *IstioTrafficPolicyTranslator) TranslateRetries(trafficPolicy *networking_v1alpha1.TrafficPolicy) *api_v1beta1.HTTPRetry {
	var translatedRetries *api_v1beta1.HTTPRetry
	retries := trafficPolicy.Spec.GetRetries()
	if retries != nil {
		translatedRetries = &api_v1beta1.HTTPRetry{
			Attempts:      retries.GetAttempts(),
			PerTryTimeout: retries.GetPerTryTimeout(),
		}
	}
	return translatedRetries
}

func (i *IstioTrafficPolicyTranslator) TranslateFaultInjection(trafficPolicy *networking_v1alpha1.TrafficPolicy) (*api_v1beta1.HTTPFaultInjection, error) {
	var translatedFaultInjection *api_v1beta1.HTTPFaultInjection
	faultInjection := trafficPolicy.Spec.GetFaultInjection()
	if faultInjection != nil {
		switch injectionType := faultInjection.GetFaultInjectionType().(type) {
		case *networking_types.FaultInjection_Abort_:
			abort := faultInjection.GetAbort()
			switch abortType := abort.GetErrorType().(type) {
			case *networking_types.FaultInjection_Abort_HttpStatus:
				translatedFaultInjection = &api_v1beta1.HTTPFaultInjection{
					Abort: &api_v1beta1.HTTPFaultInjection_Abort{
						ErrorType:  &api_v1beta1.HTTPFaultInjection_Abort_HttpStatus{HttpStatus: abort.GetHttpStatus()},
						Percentage: &api_v1beta1.Percent{Value: faultInjection.GetPercentage()},
					}}
			default:
				return nil, eris.Errorf("Abort.ErrorType has unexpected type %T", abortType)
			}
		case *networking_types.FaultInjection_Delay_:
			delay := faultInjection.GetDelay()
			switch delayType := delay.GetHttpDelayType().(type) {
			case *networking_types.FaultInjection_Delay_FixedDelay:
				translatedFaultInjection = &api_v1beta1.HTTPFaultInjection{
					Delay: &api_v1beta1.HTTPFaultInjection_Delay{
						HttpDelayType: &api_v1beta1.HTTPFaultInjection_Delay_FixedDelay{FixedDelay: delay.GetFixedDelay()},
						Percentage:    &api_v1beta1.Percent{Value: faultInjection.GetPercentage()},
					}}
			case *networking_types.FaultInjection_Delay_ExponentialDelay:
				translatedFaultInjection = &api_v1beta1.HTTPFaultInjection{
					Delay: &api_v1beta1.HTTPFaultInjection_Delay{
						HttpDelayType: &api_v1beta1.HTTPFaultInjection_Delay_ExponentialDelay{ExponentialDelay: delay.GetExponentialDelay()},
						Percentage:    &api_v1beta1.Percent{Value: faultInjection.GetPercentage()},
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

func (i *IstioTrafficPolicyTranslator) TranslateMirror(
	meshService *discovery_v1alpha1.MeshService,
	trafficPolicy *networking_v1alpha1.TrafficPolicy,
) (*api_v1beta1.Destination, error) {
	var mirror *api_v1beta1.Destination
	if trafficPolicy.Spec.GetMirror() != nil {
		mirror = &api_v1beta1.Destination{
			Host: getHostname(meshService, trafficPolicy.Spec.GetMirror().GetDestination()),
		}
	}
	return mirror, nil
}

func (i *IstioTrafficPolicyTranslator) TranslateHeaderManipulation(trafficPolicy *networking_v1alpha1.TrafficPolicy) *api_v1beta1.Headers {
	var translatedHeaderManipulation *api_v1beta1.Headers
	headerManipulation := trafficPolicy.Spec.GetHeaderManipulation()
	if headerManipulation != nil {
		translatedHeaderManipulation = &api_v1beta1.Headers{
			Request: &api_v1beta1.Headers_HeaderOperations{
				Add:    headerManipulation.GetAppendRequestHeaders(),
				Remove: headerManipulation.GetRemoveRequestHeaders(),
			},
			Response: &api_v1beta1.Headers_HeaderOperations{
				Add:    headerManipulation.GetAppendResponseHeaders(),
				Remove: headerManipulation.GetRemoveResponseHeaders(),
			},
		}
	}
	return translatedHeaderManipulation
}

func (i *IstioTrafficPolicyTranslator) TranslateCorsPolicy(trafficPolicy *networking_v1alpha1.TrafficPolicy) (*api_v1beta1.CorsPolicy, error) {
	var translatedCorsPolicy *api_v1beta1.CorsPolicy
	corsPolicy := trafficPolicy.Spec.GetCorsPolicy()
	if corsPolicy != nil {
		var allowOrigins []*api_v1beta1.StringMatch
		for i, allowOrigin := range corsPolicy.GetAllowOrigins() {
			var stringMatch *api_v1beta1.StringMatch
			switch matchType := allowOrigin.GetMatchType().(type) {
			case *networking_types.StringMatch_Exact:
				stringMatch = &api_v1beta1.StringMatch{MatchType: &api_v1beta1.StringMatch_Exact{Exact: allowOrigin.GetExact()}}
			case *networking_types.StringMatch_Prefix:
				stringMatch = &api_v1beta1.StringMatch{MatchType: &api_v1beta1.StringMatch_Prefix{Prefix: allowOrigin.GetPrefix()}}
			case *networking_types.StringMatch_Regex:
				stringMatch = &api_v1beta1.StringMatch{MatchType: &api_v1beta1.StringMatch_Regex{Regex: allowOrigin.GetRegex()}}
			default:
				return nil, eris.Errorf("AllowOrigins[%d].MatchType has unexpected type %T", i, matchType)
			}
			allowOrigins = append(allowOrigins, stringMatch)
		}
		translatedCorsPolicy = &api_v1beta1.CorsPolicy{
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

func (i *IstioTrafficPolicyTranslator) errorToStatus(err error) *networking_types.TrafficPolicyStatus_TranslatorError {
	return &networking_types.TrafficPolicyStatus_TranslatorError{
		TranslatorId: TranslatorId,
		ErrorMessage: err.Error(),
	}
}

func (i *IstioTrafficPolicyTranslator) getClusterNameForMeshService(
	ctx context.Context,
	meshService *discovery_v1alpha1.MeshService,
) (string, error) {
	mesh, err := i.meshClient.Get(ctx, client.ObjectKey{
		Name:      meshService.Spec.GetMesh().GetName(),
		Namespace: meshService.Spec.GetMesh().GetNamespace(),
	})
	if err != nil {
		return "", err
	}
	return mesh.Spec.GetCluster().GetName(), nil
}

// If destination is in the same namespace as k8s Service, return k8s Service name.namespace
// Else, return k8s Service multicluster DNS name
func getHostname(meshService *discovery_v1alpha1.MeshService, destination *core_types.ResourceRef) string {
	if common.AreResourcesOnLocalCluster(destination, meshService.Spec.GetKubeService().GetRef()) {
		// destination is on the same cluster as the k8s Service
		return meshService.Spec.GetKubeService().GetRef().GetName() + "." + meshService.Spec.GetKubeService().GetRef().GetNamespace()
	} else {
		return meshService.Spec.GetFederation().GetMulticlusterDnsName()
	}
}
