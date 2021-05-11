package trafficpolicyutils

import (
	"github.com/gogo/protobuf/types"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/rotisserie/eris"
	v1sets "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1/sets"
	v1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/destinationutils"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/hostutils"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	istiov1alpha3 "istio.io/api/networking/v1alpha3"
	networkingv1alpha3spec "istio.io/api/networking/v1alpha3"
)

func TranslateHeaderManipulation(
	headerManipulation *v1.HeaderManipulation,
) *networkingv1alpha3spec.Headers {
	if headerManipulation == nil {
		return nil
	}
	return &networkingv1alpha3spec.Headers{
		Request: &networkingv1alpha3spec.Headers_HeaderOperations{
			Add:    headerManipulation.GetAppendRequestHeaders(),
			Remove: headerManipulation.GetRemoveRequestHeaders(),
		},
		Response: &networkingv1alpha3spec.Headers_HeaderOperations{
			Add:    headerManipulation.GetAppendResponseHeaders(),
			Remove: headerManipulation.GetRemoveResponseHeaders(),
		},
	}
}

func TranslateTimeout(
	timeout *duration.Duration,
) *types.Duration {
	if timeout == nil {
		return nil
	}
	return &types.Duration{
		Seconds: timeout.Seconds,
		Nanos: timeout.Nanos,
	}
}

func TranslateRetries(
	retries *v1.TrafficPolicySpec_Policy_RetryPolicy,
) *istiov1alpha3.HTTPRetry {
	if retries == nil {
		return nil
	}
	return &istiov1alpha3.HTTPRetry{
		PerTryTimeout: translateDuration(retries.PerTryTimeout),
		Attempts: retries.Attempts,
	}
}

func TranslateFault(
	fault *v1.TrafficPolicySpec_Policy_FaultInjection,
) *istiov1alpha3.HTTPFaultInjection {
	if fault == nil {
		return nil
	}

	switch fault.FaultInjectionType.(type) {
		case *v1.TrafficPolicySpec_Policy_FaultInjection_FixedDelay:
			transformDelay, _ := fault.FaultInjectionType.(*v1.TrafficPolicySpec_Policy_FaultInjection_FixedDelay)
			return &istiov1alpha3.HTTPFaultInjection{
				Delay: &istiov1alpha3.HTTPFaultInjection_Delay{
					HttpDelayType: &istiov1alpha3.HTTPFaultInjection_Delay_FixedDelay{
						FixedDelay: translateDuration(transformDelay.FixedDelay),
					},
					Percentage: &istiov1alpha3.Percent{ Value: fault.Percentage },
				},
			}

		case *v1.TrafficPolicySpec_Policy_FaultInjection_Abort_:
			transformAbort, _ := fault.FaultInjectionType.(*v1.TrafficPolicySpec_Policy_FaultInjection_Abort_)
			return &istiov1alpha3.HTTPFaultInjection{
				Abort: &istiov1alpha3.HTTPFaultInjection_Abort{
					ErrorType: &istiov1alpha3.HTTPFaultInjection_Abort_HttpStatus{
						HttpStatus: transformAbort.Abort.HttpStatus,
					},
					Percentage: &istiov1alpha3.Percent{ Value: fault.Percentage },
				},
			}
	default:
		return &istiov1alpha3.HTTPFaultInjection{}
	}
}

func TranslateMirror(
	mirrorconfig *v1.TrafficPolicySpec_Policy_Mirror,
	clusterDomains hostutils.ClusterDomainRegistry,
	sourceCluster string,
	destinations v1sets.DestinationSet,
) (*istiov1alpha3.Destination, *istiov1alpha3.Percent, error) {
	if mirrorconfig == nil {
		return nil, nil, nil
	}

	percent := &istiov1alpha3.Percent{ Value: mirrorconfig.Percentage }
	switch mirrorconfig.DestinationType.(type) {
		case *v1.TrafficPolicySpec_Policy_Mirror_KubeService:
			mirrorDest, _ := mirrorconfig.DestinationType.(*v1.TrafficPolicySpec_Policy_Mirror_KubeService)
			dest, err := MakeKubeDestinationMirror(mirrorDest, mirrorconfig.Port, sourceCluster, destinations, clusterDomains)
			if err != nil {
				return nil, nil, eris.Wrapf(err, "invalid mirror destination")
			}
			return dest, percent, nil
		default:
			return &istiov1alpha3.Destination{
				Port: &istiov1alpha3.PortSelector{
					Number: mirrorconfig.Port,
				},
			}, percent, nil
	}
}

func TranslateCorsPolicy(
	cors *v1.TrafficPolicySpec_Policy_CorsPolicy,
) *istiov1alpha3.CorsPolicy {
	if cors == nil {
		return nil
	}

	var allowedOrigins []*istiov1alpha3.StringMatch
	for _, origin := range cors.AllowOrigins {
		switch origin.MatchType.(type) {
			case *v1.TrafficPolicySpec_Policy_CorsPolicy_StringMatch_Exact:
				matchExact, _ := origin.MatchType.(*v1.TrafficPolicySpec_Policy_CorsPolicy_StringMatch_Exact)
				match := &istiov1alpha3.StringMatch{
					MatchType: &istiov1alpha3.StringMatch_Exact{
						Exact: matchExact.Exact,
					},
				}
				allowedOrigins = append(allowedOrigins, match)
			case *v1.TrafficPolicySpec_Policy_CorsPolicy_StringMatch_Prefix:
				matchPrefix, _ := origin.MatchType.(*v1.TrafficPolicySpec_Policy_CorsPolicy_StringMatch_Prefix)
				match := &istiov1alpha3.StringMatch{
					MatchType: &istiov1alpha3.StringMatch_Prefix{
						Prefix: matchPrefix.Prefix,
					},
				}
				allowedOrigins = append(allowedOrigins, match)
			case *v1.TrafficPolicySpec_Policy_CorsPolicy_StringMatch_Regex:
				matchRegex, _ := origin.MatchType.(*v1.TrafficPolicySpec_Policy_CorsPolicy_StringMatch_Regex)
				match := &istiov1alpha3.StringMatch{
					MatchType: &istiov1alpha3.StringMatch_Regex{
						Regex: matchRegex.Regex,
					},
				}
				allowedOrigins = append(allowedOrigins, match)
		}
	}

	return &istiov1alpha3.CorsPolicy{
		AllowOrigins: allowedOrigins,
		AllowMethods: cors.AllowMethods,
		AllowHeaders: cors.AllowHeaders,
		ExposeHeaders: cors.ExposeHeaders,
		MaxAge: translateDuration(cors.MaxAge),
		AllowCredentials: translateBoolValue(cors.AllowCredentials),
	}
}

func translateDuration(
	duration *duration.Duration,
) *types.Duration {
	if duration == nil {
		return nil
	}

	return &types.Duration{
		Seconds: duration.Seconds,
		Nanos: duration.Nanos,
	}
}

func translateBoolValue(
	boolValue  *wrappers.BoolValue,
) *types.BoolValue {
	if boolValue == nil {
		return nil
	}

	return &types.BoolValue{ Value: boolValue.Value }
}

func MakeKubeDestinationMirror(
	mirrorDest *v1.TrafficPolicySpec_Policy_Mirror_KubeService,
	port uint32,
	sourceClusterName string,
	destinations v1sets.DestinationSet,
	clusterDomains hostutils.ClusterDomainRegistry,
) (*networkingv1alpha3spec.Destination, error) {
	destinationRef := mirrorDest.KubeService
	mirrorService, err := destinationutils.FindDestinationForKubeService(destinations.List(), destinationRef)
	if err != nil {
		return nil, eris.Wrapf(err, "invalid mirror destination")
	}
	mirrorKubeService := mirrorService.Spec.GetKubeService()

	// TODO: empty sourceClusterName?
	destinationHostname := clusterDomains.GetDestinationFQDN(
		sourceClusterName,
		destinationRef,
	)

	translatedMirror := &networkingv1alpha3spec.Destination{
		Host: destinationHostname,
	}

	if port != 0 {
		if !ContainsPort(mirrorKubeService.Ports, port) {
			return nil, eris.Errorf("specified port %d does not exist for mirror destination service %v", port, sets.Key(mirrorKubeService.Ref))
		}
		translatedMirror.Port = &networkingv1alpha3spec.PortSelector{
			Number: port,
		}
	} else {
		// validate that Destination only has one port
		if numPorts := len(mirrorKubeService.GetPorts()); numPorts > 1 {
			return nil, eris.Errorf("must provide port for mirror destination service %v with multiple ports (%v) defined", sets.Key(mirrorKubeService.GetRef()), numPorts)
		}
	}

	return translatedMirror, nil
}