package trafficpolicyutils

import (
	"github.com/gogo/protobuf/types"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/rotisserie/eris"
	v1sets "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1/sets"
	v1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/destinationutils"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/gogoutils"
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
	return gogoutils.DurationProtoToGogo(timeout)
}

func TranslateRetries(
	retries *v1.TrafficPolicySpec_Policy_RetryPolicy,
) *istiov1alpha3.HTTPRetry {
	if retries == nil {
		return nil
	}
	return &networkingv1alpha3spec.HTTPRetry{
		Attempts:      retries.GetAttempts(),
		PerTryTimeout: gogoutils.DurationProtoToGogo(retries.GetPerTryTimeout()),
	}
}

func TranslateFault(faultInjection *v1.TrafficPolicySpec_Policy_FaultInjection) (*networkingv1alpha3spec.HTTPFaultInjection, error) {
	if faultInjection == nil {
		return nil, nil
	}
	if faultInjection.GetFaultInjectionType() == nil {
		return nil, eris.New("FaultInjection type must be specified.")
	}
	var translatedFaultInjection *networkingv1alpha3spec.HTTPFaultInjection
	switch injectionType := faultInjection.GetFaultInjectionType().(type) {
	case *v1.TrafficPolicySpec_Policy_FaultInjection_Abort_:
		translatedFaultInjection = &networkingv1alpha3spec.HTTPFaultInjection{
			Abort: &networkingv1alpha3spec.HTTPFaultInjection_Abort{
				ErrorType: &networkingv1alpha3spec.HTTPFaultInjection_Abort_HttpStatus{
					HttpStatus: faultInjection.GetAbort().GetHttpStatus(),
				},
				Percentage: &networkingv1alpha3spec.Percent{Value: faultInjection.GetPercentage()},
			},
		}
	case *v1.TrafficPolicySpec_Policy_FaultInjection_FixedDelay:
		translatedFaultInjection = &networkingv1alpha3spec.HTTPFaultInjection{
			Delay: &networkingv1alpha3spec.HTTPFaultInjection_Delay{
				HttpDelayType: &networkingv1alpha3spec.HTTPFaultInjection_Delay_FixedDelay{
					FixedDelay: gogoutils.DurationProtoToGogo(faultInjection.GetFixedDelay()),
				},
				Percentage: &networkingv1alpha3spec.Percent{Value: faultInjection.GetPercentage()},
			},
		}
	default:
		return nil, eris.Errorf("FaultInjection.FaultInjectionType has unexpected type %T", injectionType)
	}
	return translatedFaultInjection, nil
}

func TranslateCorsPolicy(
	corsPolicy *v1.TrafficPolicySpec_Policy_CorsPolicy,
) (*istiov1alpha3.CorsPolicy, error) {
	if corsPolicy == nil {
		return nil, nil
	}
	var allowOrigins []*networkingv1alpha3spec.StringMatch
	for i, allowOrigin := range corsPolicy.GetAllowOrigins() {
		var stringMatch *networkingv1alpha3spec.StringMatch
		switch matchType := allowOrigin.GetMatchType().(type) {
		case *v1.StringMatch_Exact:
			stringMatch = &networkingv1alpha3spec.StringMatch{MatchType: &networkingv1alpha3spec.StringMatch_Exact{Exact: allowOrigin.GetExact()}}
		case *v1.StringMatch_Prefix:
			stringMatch = &networkingv1alpha3spec.StringMatch{MatchType: &networkingv1alpha3spec.StringMatch_Prefix{Prefix: allowOrigin.GetPrefix()}}
		case *v1.StringMatch_Regex:
			stringMatch = &networkingv1alpha3spec.StringMatch{MatchType: &networkingv1alpha3spec.StringMatch_Regex{Regex: allowOrigin.GetRegex()}}
		default:
			return nil, eris.Errorf("AllowOrigins[%d].MatchType has unexpected type %T", i, matchType)
		}
		allowOrigins = append(allowOrigins, stringMatch)
	}
	translatedCorsPolicy := &networkingv1alpha3spec.CorsPolicy{
		AllowOrigins:     allowOrigins,
		AllowMethods:     corsPolicy.GetAllowMethods(),
		AllowHeaders:     corsPolicy.GetAllowHeaders(),
		ExposeHeaders:    corsPolicy.GetExposeHeaders(),
		MaxAge:           gogoutils.DurationProtoToGogo(corsPolicy.GetMaxAge()),
		AllowCredentials: gogoutils.BoolProtoToGogo(corsPolicy.GetAllowCredentials()),
	}

	return translatedCorsPolicy, nil
}

// If federatedClusterName is non-empty, it indicates translation for a federated VirtualService, so use it as the source cluster name.
func TranslateMirror(
	mirror *v1.TrafficPolicySpec_Policy_Mirror,
	sourceClusterName string,
	clusterDomains hostutils.ClusterDomainRegistry,
	destinations v1sets.DestinationSet,
) (*networkingv1alpha3spec.Destination, *networkingv1alpha3spec.Percent, error) {
	if mirror == nil {
		return nil, nil, nil
	}
	if mirror.DestinationType == nil {
		return nil, nil, eris.Errorf("must provide mirror destination")
	}

	var translatedMirror *networkingv1alpha3spec.Destination
	switch destinationType := mirror.DestinationType.(type) {
	case *v1.TrafficPolicySpec_Policy_Mirror_KubeService:
		var err error
		translatedMirror, err = makeKubeDestinationMirror(
			destinationType,
			mirror.Port,
			sourceClusterName,
			clusterDomains,
			destinations,
		)
		if err != nil {
			return nil, nil, err
		}
	}

	mirrorPercentage := &networkingv1alpha3spec.Percent{
		Value: mirror.GetPercentage(),
	}

	return translatedMirror, mirrorPercentage, nil
}

func makeKubeDestinationMirror(
	mirrorDest *v1.TrafficPolicySpec_Policy_Mirror_KubeService,
	port uint32,
	sourceClusterName string,
	clusterDomains hostutils.ClusterDomainRegistry,
	destinations v1sets.DestinationSet,
) (*networkingv1alpha3spec.Destination, error) {
	destinationRef := mirrorDest.KubeService
	mirrorService, err := destinationutils.FindDestinationForKubeService(destinations.List(), destinationRef)
	if err != nil {
		return nil, eris.Wrapf(err, "invalid mirror destination")
	}
	mirrorKubeService := mirrorService.Spec.GetKubeService()

	// TODO(ilackarms): support other types of Destination destinations, e.g. via ServiceEntries

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
