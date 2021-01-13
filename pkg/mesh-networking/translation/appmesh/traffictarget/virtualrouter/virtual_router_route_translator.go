package virtualrouter

import (
	"fmt"

	appmeshv1beta2 "github.com/aws/aws-app-mesh-controller-for-k8s/apis/appmesh/v1beta2"
	discoveryv1alpha2 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2"
	discoveryv1alpha2sets "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2/sets"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1alpha2"
	v1alpha2types "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1alpha2/types"
	"github.com/solo-io/gloo-mesh/pkg/mesh-discovery/utils/workloadutils"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/traffictargetutils"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
)

type routeTranslator struct {
	reporter       reporting.Reporter
	trafficTargets discoveryv1alpha2sets.TrafficTargetSet
	workloads      discoveryv1alpha2sets.WorkloadSet
}

func newRouteTranslator(reporter reporting.Reporter, trafficTargets discoveryv1alpha2sets.TrafficTargetSet, workloads discoveryv1alpha2sets.WorkloadSet) *routeTranslator {
	return &routeTranslator{
		reporter:       reporter,
		trafficTargets: trafficTargets,
		workloads:      workloads,
	}
}

func (r *routeTranslator) getRoutes(trafficTarget *discoveryv1alpha2.TrafficTarget) []appmeshv1beta2.Route {
	var routes []appmeshv1beta2.Route
	for _, tp := range trafficTarget.Status.GetAppliedTrafficPolicies() {
		routes = append(routes, r.getTrafficPolicyRoutes(trafficTarget, tp.Ref, tp.Spec)...)
	}
	return routes
}

func (r *routeTranslator) getTrafficPolicyRoutes(trafficTarget *discoveryv1alpha2.TrafficTarget, trafficPolicyRef *v1.ObjectRef, trafficPolicy *v1alpha2.TrafficPolicySpec) []appmeshv1beta2.Route {
	getMatches := func(networkingMatchers []*v1alpha2.TrafficPolicySpec_HttpMatcher) []appmeshv1beta2.HTTPRouteMatch {
		if len(networkingMatchers) == 0 {
			// If there are no networking matchers, insert a * matcher.
			networkingMatchers = append(networkingMatchers, &v1alpha2.TrafficPolicySpec_HttpMatcher{
				PathSpecifier: &v1alpha2.TrafficPolicySpec_HttpMatcher_Prefix{},
			})
		}

		var httpRouteMatches []appmeshv1beta2.HTTPRouteMatch
		for _, nm := range networkingMatchers {
			// TODO report any non-prefix matchers as they're not supported by app mesh
			prefix := nm.GetPrefix()
			if prefix == "" {
				prefix = "/"
			}

			httpRouteMatches = append(httpRouteMatches, appmeshv1beta2.HTTPRouteMatch{
				Headers: convertHeaders(nm.Headers),
				Method:  convertMethod(nm.Method),
				Prefix:  prefix,
			})
		}

		return httpRouteMatches
	}

	getRouteAction := func() appmeshv1beta2.HTTPRouteAction {
		// If there is no traffic shift, split traffic among the virtual nodes backing this traffic target.
		if trafficPolicy.GetTrafficShift() == nil {
			var weightedTargets []appmeshv1beta2.WeightedTarget
			for _, workload := range workloadutils.FindBackingWorkloads(trafficTarget.Spec.GetKubeService(), r.workloads) {
				if workload.Spec.AppMesh == nil {
					// TODO joekelley report and error out
				}

				weightedTargets = append(weightedTargets, appmeshv1beta2.WeightedTarget{
					VirtualNodeARN: &workload.Spec.AppMesh.VirtualNodeArn,
					Weight:         1,
				})
			}

			return appmeshv1beta2.HTTPRouteAction{WeightedTargets: weightedTargets}
		}

		var weightedTargets []appmeshv1beta2.WeightedTarget
		for _, destination := range trafficPolicy.GetTrafficShift().GetDestinations() {

			kubeServiceDestination := destination.GetKubeService()
			if kubeServiceDestination == nil {
				// TODO joekelley report on anything but kube service
			}

			destinationTrafficTarget, err := traffictargetutils.FindTrafficTargetForKubeService(r.trafficTargets.List(), kubeServiceDestination)
			if err != nil {
				// TODO joekelley here and below
			}

			// Route traffic to one backing workload for the provided service
			// TODO split traffic among all backing workloads
			backingWorkloads := workloadutils.FindBackingWorkloads(destinationTrafficTarget.Spec.GetKubeService(), r.workloads)
			if len(backingWorkloads) == 0 {
				// TODO
			}
			workload := backingWorkloads[0]

			arn := workload.Spec.AppMesh.VirtualNodeArn
			if arn == "" {
				// TODO joekelley
			}

			weightedTargets = append(weightedTargets, appmeshv1beta2.WeightedTarget{
				VirtualNodeARN: &arn,
				Weight:         int64(destination.Weight),
			})
		}

		return appmeshv1beta2.HTTPRouteAction{
			WeightedTargets: weightedTargets,
		}
	}

	getRetryPolicy := func() *appmeshv1beta2.HTTPRetryPolicy {
		if trafficPolicy.Retries == nil {
			return nil
		}

		var perRetryTimeout appmeshv1beta2.Duration
		if trafficPolicy.Retries.PerTryTimeout != nil {
			perRetryTimeout.Value = trafficPolicy.Retries.PerTryTimeout.Seconds
			perRetryTimeout.Unit = appmeshv1beta2.DurationUnitS
		}

		// Use all supported HTTP and TCP retry events.
		return &appmeshv1beta2.HTTPRetryPolicy{
			HTTPRetryEvents: []appmeshv1beta2.HTTPRetryPolicyEvent{"server-error", "gateway-error", "client-error", "stream-error"},
			TCPRetryEvents:  []appmeshv1beta2.TCPRetryPolicyEvent{"connection-error"},
			MaxRetries:      int64(trafficPolicy.Retries.Attempts),
			PerRetryTimeout: perRetryTimeout,
		}
	}

	getTimeoutPolicy := func() *appmeshv1beta2.HTTPTimeout {
		if trafficPolicy.RequestTimeout == nil {
			return nil
		}

		return &appmeshv1beta2.HTTPTimeout{
			PerRequest: &appmeshv1beta2.Duration{
				Unit:  appmeshv1beta2.DurationUnitS,
				Value: trafficPolicy.RequestTimeout.Seconds,
			},
		}
	}

	var routes []appmeshv1beta2.Route
	for i, routeMatch := range getMatches(trafficPolicy.HttpRequestMatchers) {
		routes = append(routes, appmeshv1beta2.Route{
			Name: fmt.Sprintf("%s-%s-%d", trafficPolicyRef.Namespace, trafficPolicyRef.Name, i),
			HTTPRoute: &appmeshv1beta2.HTTPRoute{
				Match:       routeMatch,
				Action:      getRouteAction(),
				RetryPolicy: getRetryPolicy(),
				Timeout:     getTimeoutPolicy(),
			},
		})
	}

	return routes
}

func convertHeaders(in []*v1alpha2.TrafficPolicySpec_HeaderMatcher) []appmeshv1beta2.HTTPRouteHeader {
	var output []appmeshv1beta2.HTTPRouteHeader
	for _, headerMatcher := range in {
		headerValue := headerMatcher.Value
		invert := headerMatcher.InvertMatch

		matchMethod := &appmeshv1beta2.HeaderMatchMethod{}
		if headerMatcher.Regex {
			matchMethod.Regex = &headerValue
		} else {
			matchMethod.Exact = &headerValue
		}

		output = append(output, appmeshv1beta2.HTTPRouteHeader{
			Name:   headerMatcher.Name,
			Match:  matchMethod,
			Invert: &invert,
		})
	}
	return output
}

func convertMethod(in *v1alpha2.TrafficPolicySpec_HttpMethod) *string {
	var str string

	if in == nil {
		return nil
	}

	switch in.Method {
	case v1alpha2types.HttpMethodValue_GET:
		str = "GET"
	case v1alpha2types.HttpMethodValue_POST:
		str = "POST"
	case v1alpha2types.HttpMethodValue_PUT:
		str = "PUT"
	case v1alpha2types.HttpMethodValue_DELETE:
		str = "DELETE"
	case v1alpha2types.HttpMethodValue_HEAD:
		str = "HEAD"
	case v1alpha2types.HttpMethodValue_CONNECT:
		str = "CONNECT"
	case v1alpha2types.HttpMethodValue_OPTIONS:
		str = "OPTIONS"
	case v1alpha2types.HttpMethodValue_TRACE:
		str = "TRACE"
	case v1alpha2types.HttpMethodValue_PATCH:
		str = "PATCH"
	default:
		return nil
	}
	return &str
}
