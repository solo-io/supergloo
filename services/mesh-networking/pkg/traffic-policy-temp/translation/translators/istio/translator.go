package istio_translator

import (
	"sort"
	"strings"

	"github.com/rotisserie/eris"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	zephyr_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/clients"
	"github.com/solo-io/service-mesh-hub/pkg/constants"
	"github.com/solo-io/service-mesh-hub/pkg/selector"
	mesh_translation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/translation/translators"
	istio_networking_types "istio.io/api/networking/v1alpha3"
	istio_client_networking_types "istio.io/client-go/pkg/apis/networking/v1alpha3"
)

const (
	TranslatorId = "istio_translator"
)

func NewIstioTrafficPolicyTranslator(
	resourceSelector selector.ResourceSelector,
) mesh_translation.IstioTranslator {
	return &istioTrafficPolicyTranslator{
		resourceSelector: resourceSelector,
	}
}

type istioTrafficPolicyTranslator struct {
	resourceSelector selector.ResourceSelector
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
func (i *istioTrafficPolicyTranslator) Translate(
	meshService *zephyr_discovery.MeshService,
	allMeshServices []*zephyr_discovery.MeshService,
	mesh *zephyr_discovery.Mesh,
	trafficPolicies []*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy,
) (*mesh_translation.IstioTranslationOutput, []*mesh_translation.TranslationError) {
	if mesh.Spec.GetIstio() == nil {
		return nil, nil
	}

	destinationRule := i.buildDestinationRule(meshService, allMeshServices)

	virtualService, translationErrors := i.buildVirtualService(meshService, allMeshServices, trafficPolicies)
	if len(translationErrors) > 0 {
		return nil, translationErrors
	}

	var virtualServices []*istio_client_networking_types.VirtualService
	if virtualService != nil {
		virtualServices = []*istio_client_networking_types.VirtualService{virtualService}
	}

	return &mesh_translation.IstioTranslationOutput{
		VirtualServices:  virtualServices,
		DestinationRules: []*istio_client_networking_types.DestinationRule{destinationRule},
	}, nil
}

func (i *istioTrafficPolicyTranslator) GetTranslationErrors(
	meshService *zephyr_discovery.MeshService,
	allMeshServices []*zephyr_discovery.MeshService,
	mesh *zephyr_discovery.Mesh,
	trafficPolicies []*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy,
) []*mesh_translation.TranslationError {
	_, errors := i.Translate(meshService, allMeshServices, mesh, trafficPolicies)
	return errors
}

func (*istioTrafficPolicyTranslator) GetTranslationLabels() map[string]string {
	return map[string]string{
		constants.ManagedByLabel: constants.ServiceMeshHubApplicationName,
	}
}

func (i *istioTrafficPolicyTranslator) buildDestinationRule(
	meshService *zephyr_discovery.MeshService,
	allMeshServices []*zephyr_discovery.MeshService,
) *istio_client_networking_types.DestinationRule {
	return &istio_client_networking_types.DestinationRule{
		ObjectMeta: clients.ResourceRefToObjectMeta(meshService.Spec.GetKubeService().GetRef()),
		Spec: istio_networking_types.DestinationRule{
			Host: i.buildServiceHostname(meshService),
			TrafficPolicy: &istio_networking_types.TrafficPolicy{
				Tls: &istio_networking_types.TLSSettings{
					Mode: istio_networking_types.TLSSettings_ISTIO_MUTUAL,
				},
			},
			Subsets: i.findReferencedSubsetsForService(meshService, allMeshServices),
		},
	}
}

// find all the subsets that are referenced in TrafficShift specs of other validated traffic policies
func (i *istioTrafficPolicyTranslator) findReferencedSubsetsForService(
	serviceBeingShiftedTo *zephyr_discovery.MeshService,
	allMeshServices []*zephyr_discovery.MeshService,
) (subsets []*istio_networking_types.Subset) {
	serviceBeingShiftedToList := []*zephyr_discovery.MeshService{serviceBeingShiftedTo}

	for _, service := range allMeshServices {
		for _, validatedTrafficPolicy := range service.Status.GetValidatedTrafficPolicies() {
			for _, destination := range validatedTrafficPolicy.TrafficPolicySpec.GetTrafficShift().GetDestinations() {
				result := i.resourceSelector.FindMeshServiceByRefSelector(
					serviceBeingShiftedToList,
					destination.GetDestination().GetName(),
					destination.GetDestination().GetNamespace(),
					destination.GetDestination().GetCluster(),
				)
				if result == nil {
					continue
				}

				// our service being shifted to is referenced in this traffic shift; record all the subsets
				subsets = append(subsets, &istio_networking_types.Subset{
					Name:   i.buildUniqueSubsetName(destination.Subset),
					Labels: destination.Subset,
				})
			}
		}
	}

	return subsets
}

func (i *istioTrafficPolicyTranslator) buildVirtualService(
	meshService *zephyr_discovery.MeshService,
	allMeshServices []*zephyr_discovery.MeshService,
	trafficPolicies []*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy,
) (*istio_client_networking_types.VirtualService, []*mesh_translation.TranslationError) {
	computedVirtualService, translationErrors := i.translateIntoVirtualService(meshService, allMeshServices, trafficPolicies)
	if len(translationErrors) > 0 {
		return nil, translationErrors
	}
	// The translator will attempt to create a virtual service for every MeshService.
	// However, a mesh service with no Http, Tcp, or Tls Routes is invalid, so as we only support Http here,
	// we simply return nil. This will not be true for MeshServices which have been configured by TrafficPolicies
	if len(computedVirtualService.Spec.GetHttp()) == 0 {
		// TODO: Is this a bug? Why are we only checking GetHttp()?
		return nil, nil
	}

	return computedVirtualService, nil
}

func (i *istioTrafficPolicyTranslator) translateIntoVirtualService(
	meshService *zephyr_discovery.MeshService,
	allMeshServices []*zephyr_discovery.MeshService,
	validatedPolicies []*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy,
) (*istio_client_networking_types.VirtualService, []*mesh_translation.TranslationError) {
	virtualService := &istio_client_networking_types.VirtualService{
		ObjectMeta: clients.ResourceRefToObjectMeta(meshService.Spec.GetKubeService().GetRef()),
		Spec: istio_networking_types.VirtualService{
			Hosts: []string{i.buildServiceHostname(meshService)},
		},
	}
	var allHttpRoutes []*istio_networking_types.HTTPRoute
	var translationErrors []*mesh_translation.TranslationError
	for _, validatedPolicy := range validatedPolicies {
		httpRoutes, err := i.translateIntoHTTPRoutes(meshService, allMeshServices, validatedPolicy)
		if err != nil {
			translationErrors = append(translationErrors, &mesh_translation.TranslationError{
				Policy: validatedPolicy,
				TranslatorErrors: []*zephyr_networking_types.TrafficPolicyStatus_TranslatorError{{
					TranslatorId: TranslatorId,
					ErrorMessage: err.Error(),
				}},
			})
			continue
		}
		allHttpRoutes = append(allHttpRoutes, httpRoutes...)
	}

	if len(translationErrors) > 0 {
		return nil, translationErrors
	}

	sort.Sort(SpecificitySortableRoutes(allHttpRoutes))
	virtualService.Spec.Http = allHttpRoutes
	return virtualService, nil
}

func (i *istioTrafficPolicyTranslator) translateIntoHTTPRoutes(
	meshService *zephyr_discovery.MeshService,
	allMeshServices []*zephyr_discovery.MeshService,
	validatedPolicy *zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy,
) ([]*istio_networking_types.HTTPRoute, error) {
	var err error
	var faultInjection *istio_networking_types.HTTPFaultInjection
	if faultInjection, err = i.translateFaultInjection(validatedPolicy); err != nil {
		return nil, err
	}

	var corsPolicy *istio_networking_types.CorsPolicy
	if corsPolicy, err = i.translateCorsPolicy(validatedPolicy); err != nil {
		return nil, err
	}

	var requestMatchers []*istio_networking_types.HTTPMatchRequest
	if requestMatchers, err = i.translateRequestMatchers(validatedPolicy); err != nil {
		return nil, err
	}

	var mirrorPercentage *istio_networking_types.Percent
	if validatedPolicy.TrafficPolicySpec.GetMirror() != nil {
		mirrorPercentage = &istio_networking_types.Percent{Value: validatedPolicy.TrafficPolicySpec.GetMirror().GetPercentage()}
	}

	var mirror *istio_networking_types.Destination
	if mirror, err = i.translateMirror(meshService, allMeshServices, validatedPolicy); err != nil {
		return nil, err
	}

	var trafficShift []*istio_networking_types.HTTPRouteDestination
	if trafficShift, err = i.translateDestinationRoutes(meshService, allMeshServices, validatedPolicy); err != nil {
		return nil, err
	}
	retries := i.translateRetries(validatedPolicy)
	headerManipulation := i.translateHeaderManipulation(validatedPolicy)
	var httpRoutes []*istio_networking_types.HTTPRoute

	if len(requestMatchers) == 0 {
		// If no matchers are present return a single route with no matchers
		httpRoutes = append(httpRoutes, &istio_networking_types.HTTPRoute{
			Route:            trafficShift,
			Timeout:          validatedPolicy.TrafficPolicySpec.GetRequestTimeout(),
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
				Timeout:          validatedPolicy.TrafficPolicySpec.GetRequestTimeout(),
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
	validatedPolicy *zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy,
) ([]*istio_networking_types.HTTPMatchRequest, error) {
	// Generate HttpMatchRequests for SourceSelector, one per namespace.
	var sourceMatchers []*istio_networking_types.HTTPMatchRequest
	// Set SourceNamespace and SourceLabels.
	if len(validatedPolicy.TrafficPolicySpec.GetSourceSelector().GetLabels()) > 0 ||
		len(validatedPolicy.TrafficPolicySpec.GetSourceSelector().GetNamespaces()) > 0 {
		if len(validatedPolicy.TrafficPolicySpec.GetSourceSelector().GetNamespaces()) > 0 {
			for _, namespace := range validatedPolicy.TrafficPolicySpec.GetSourceSelector().GetNamespaces() {
				matchRequest := &istio_networking_types.HTTPMatchRequest{
					SourceNamespace: namespace,
					SourceLabels:    validatedPolicy.TrafficPolicySpec.GetSourceSelector().GetLabels(),
				}
				sourceMatchers = append(sourceMatchers, matchRequest)
			}
		} else {
			sourceMatchers = append(sourceMatchers, &istio_networking_types.HTTPMatchRequest{
				SourceLabels: validatedPolicy.TrafficPolicySpec.GetSourceSelector().GetLabels(),
			})
		}
	}
	if validatedPolicy.TrafficPolicySpec.GetHttpRequestMatchers() == nil {
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
		for _, matcher := range validatedPolicy.TrafficPolicySpec.GetHttpRequestMatchers() {
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

// For each Destination, create an Istio HTTPRouteDestination
func (i *istioTrafficPolicyTranslator) translateDestinationRoutes(
	meshService *zephyr_discovery.MeshService,
	allMeshServices []*zephyr_discovery.MeshService,
	validatedPolicy *zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy,
) ([]*istio_networking_types.HTTPRouteDestination, error) {
	var translatedRouteDestinations []*istio_networking_types.HTTPRouteDestination
	trafficShift := validatedPolicy.TrafficPolicySpec.GetTrafficShift()
	if trafficShift != nil {
		for _, destination := range trafficShift.GetDestinations() {
			hostnameForKubeService, isMulticluster, err := i.getHostnameForKubeService(
				meshService,
				allMeshServices,
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

				// Build a deterministic, unique name for this subset.
				// It's fine if that subset name doesn't exist on that DestinationRule when this particular line executes;
				// since this method is deterministic, we'll build the same name when we eventually process that Destination
				// Rule's relevant MeshService
				httpRouteDestination.Destination.Subset = i.buildUniqueSubsetName(destination.GetSubset())
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
					Host: i.buildServiceHostname(meshService),
					Port: &istio_networking_types.PortSelector{
						Number: defaultServicePort.Port,
					},
				},
			},
		}
	}
	return translatedRouteDestinations, nil
}

func (i *istioTrafficPolicyTranslator) translateRetries(validatedPolicy *zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy) *istio_networking_types.HTTPRetry {
	var translatedRetries *istio_networking_types.HTTPRetry
	retries := validatedPolicy.TrafficPolicySpec.GetRetries()
	if retries != nil {
		translatedRetries = &istio_networking_types.HTTPRetry{
			Attempts:      retries.GetAttempts(),
			PerTryTimeout: retries.GetPerTryTimeout(),
		}
	}
	return translatedRetries
}

func (i *istioTrafficPolicyTranslator) translateFaultInjection(validatedPolicy *zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy) (*istio_networking_types.HTTPFaultInjection, error) {
	var translatedFaultInjection *istio_networking_types.HTTPFaultInjection
	faultInjection := validatedPolicy.TrafficPolicySpec.GetFaultInjection()
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
	meshService *zephyr_discovery.MeshService,
	allMeshServices []*zephyr_discovery.MeshService,
	validatedPolicy *zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy,
) (*istio_networking_types.Destination, error) {
	var mirror *istio_networking_types.Destination
	if validatedPolicy.TrafficPolicySpec.GetMirror() != nil {
		hostnameForKubeService, _, err := i.getHostnameForKubeService(
			meshService,
			allMeshServices,
			validatedPolicy.TrafficPolicySpec.GetMirror().GetDestination(),
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
		if validatedPolicy.TrafficPolicySpec.GetMirror().GetPort() != 0 {
			mirror.Port = &istio_networking_types.PortSelector{
				Number: validatedPolicy.TrafficPolicySpec.GetMirror().GetPort(),
			}
		}
	}
	return mirror, nil
}

func (i *istioTrafficPolicyTranslator) translateHeaderManipulation(validatedPolicy *zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy) *istio_networking_types.Headers {
	var translatedHeaderManipulation *istio_networking_types.Headers
	headerManipulation := validatedPolicy.TrafficPolicySpec.GetHeaderManipulation()
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

func (i *istioTrafficPolicyTranslator) translateCorsPolicy(validatedPolicy *zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy) (*istio_networking_types.CorsPolicy, error) {
	var translatedCorsPolicy *istio_networking_types.CorsPolicy
	corsPolicy := validatedPolicy.TrafficPolicySpec.GetCorsPolicy()
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

// If destination is in the same namespace as k8s Service, return k8s Service name.namespace
// Else, return k8s Service multicluster DNS name
func (i *istioTrafficPolicyTranslator) getHostnameForKubeService(
	meshService *zephyr_discovery.MeshService,
	allMeshServices []*zephyr_discovery.MeshService,
	destination *zephyr_core_types.ResourceRef,
) (hostname string, isMulticluster bool, err error) {
	destinationMeshService := i.resourceSelector.FindMeshServiceByRefSelector(
		allMeshServices,
		destination.GetName(),
		destination.GetNamespace(),
		destination.GetCluster(),
	)

	if destinationMeshService == nil {
		return "", false, selector.MeshServiceNotFound(destination.GetName(), destination.GetNamespace(), destination.GetCluster())
	}

	if destination.GetCluster() == meshService.Spec.GetKubeService().GetRef().GetCluster() {
		// destination is on the same cluster as the MeshService's k8s Service
		return i.buildServiceHostname(destinationMeshService), false, nil
	} else {
		// destination is on a remote cluster to the MeshService's k8s Service
		return destinationMeshService.Spec.GetFederation().GetMulticlusterDnsName(), true, nil
	}
}

func (*istioTrafficPolicyTranslator) buildServiceHostname(meshService *zephyr_discovery.MeshService) string {
	// we write the destination rule to the same namespace as the k8s service, so it's fine to just use the service name
	// TODO: Is this always right?
	return meshService.Spec.GetKubeService().GetRef().GetName()
}

// sort the label keys, then in order concatenate keys-values
func (*istioTrafficPolicyTranslator) buildUniqueSubsetName(selectors map[string]string) string {
	var keys []string
	for key, val := range selectors {
		keys = append(keys, key+"-"+val)
	}
	sort.Strings(keys)
	return strings.Join(keys, "_")
}
