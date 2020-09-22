package trafficshift

import (
	"reflect"
	"sort"
	"strings"

	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/kubeutils"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	discoveryv1alpha2sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2/sets"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	v1alpha2sets "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2/sets"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/failoverserviceutils"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/hostutils"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/trafficpolicyutils"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/traffictargetutils"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/solo-io/skv2/pkg/ezkube"
	networkingv1alpha3spec "istio.io/api/networking/v1alpha3"
)

const (
	decoratorName = "traffic-shift"
)

func init() {
	decorators.Register(decoratorConstructor)
}

func decoratorConstructor(params decorators.Parameters) decorators.Decorator {
	return NewTrafficShiftDecorator(
		params.ClusterDomains,
		params.Snapshot.TrafficTargets(),
		params.Snapshot.FailoverServices(),
	)
}

// handles setting Weighted Destinations on a VirtualService
type trafficShiftDecorator struct {
	clusterDomains   hostutils.ClusterDomainRegistry
	trafficTargets   discoveryv1alpha2sets.TrafficTargetSet
	failoverServices v1alpha2sets.FailoverServiceSet
}

var _ decorators.TrafficPolicyVirtualServiceDecorator = &trafficShiftDecorator{}

func NewTrafficShiftDecorator(
	clusterDomains hostutils.ClusterDomainRegistry,
	trafficTargets discoveryv1alpha2sets.TrafficTargetSet,
	failoverServices v1alpha2sets.FailoverServiceSet,
) *trafficShiftDecorator {
	return &trafficShiftDecorator{
		clusterDomains:   clusterDomains,
		trafficTargets:   trafficTargets,
		failoverServices: failoverServices,
	}
}

func (d *trafficShiftDecorator) DecoratorName() string {
	return decoratorName
}

func (d *trafficShiftDecorator) ApplyTrafficPolicyToVirtualService(
	appliedPolicy *discoveryv1alpha2.TrafficTargetStatus_AppliedTrafficPolicy,
	service *discoveryv1alpha2.TrafficTarget,
	output *networkingv1alpha3spec.HTTPRoute,
	registerField decorators.RegisterField,
) error {
	trafficShiftDestinations, err := d.translateTrafficShift(service, appliedPolicy.Spec)
	if err != nil {
		return err
	}
	if trafficShiftDestinations != nil {
		if err := registerField(&output.Route, trafficShiftDestinations); err != nil {
			return err
		}
		output.Route = trafficShiftDestinations
	}
	return nil
}

func (d *trafficShiftDecorator) translateTrafficShift(
	trafficTarget *discoveryv1alpha2.TrafficTarget,
	trafficPolicy *v1alpha2.TrafficPolicySpec,
) ([]*networkingv1alpha3spec.HTTPRouteDestination, error) {
	trafficShift := trafficPolicy.GetTrafficShift()
	if trafficShift == nil {
		return nil, nil
	}

	var shiftedDestinations []*networkingv1alpha3spec.HTTPRouteDestination
	for _, destination := range trafficShift.Destinations {
		if destination.DestinationType == nil {
			return nil, eris.Errorf("must set a destination type on traffic shift destination")
		}
		var trafficShiftDestination *networkingv1alpha3spec.HTTPRouteDestination
		switch destinationType := destination.DestinationType.(type) {
		case *v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination_KubeService:
			var err error
			trafficShiftDestination, err = d.buildKubeTrafficShiftDestination(
				destinationType.KubeService,
				trafficTarget,
				destination.Weight,
			)
			if err != nil {
				return nil, err
			}
		case *v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination_FailoverService:
			var err error
			trafficShiftDestination, err = d.buildFailoverServiceDestination(
				destinationType.FailoverService,
				destination.Weight,
			)
			if err != nil {
				return nil, err
			}
		default:
			return nil, eris.Errorf("unsupported traffic shift destination type: %T", destination.DestinationType)
		}
		shiftedDestinations = append(shiftedDestinations, trafficShiftDestination)

	}

	return shiftedDestinations, nil
}

func (d *trafficShiftDecorator) buildKubeTrafficShiftDestination(
	kubeDest *v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination_KubeDestination,
	originalService *discoveryv1alpha2.TrafficTarget,
	weight uint32,
) (*networkingv1alpha3spec.HTTPRouteDestination, error) {
	originalKubeService := originalService.Spec.GetKubeService()

	if originalKubeService == nil {
		return nil, eris.Errorf("traffic shift only supported for kube traffic targets")
	}
	if kubeDest == nil {
		return nil, eris.Errorf("nil kube destination on traffic shift")
	}

	svcRef := &v1.ClusterObjectRef{
		Name:        kubeDest.Name,
		Namespace:   kubeDest.Namespace,
		ClusterName: kubeDest.ClusterName,
	}

	// validate destination service is a known traffictarget
	trafficShiftService, err := traffictargetutils.FindTrafficTargetForKubeService(d.trafficTargets.List(), svcRef)
	if err != nil {
		return nil, eris.Wrapf(err, "invalid traffic shift destination %s", sets.Key(svcRef))
	}
	trafficShiftKubeService := trafficShiftService.Spec.GetKubeService()

	sourceCluster := originalKubeService.Ref.ClusterName
	destinationHost := d.clusterDomains.GetDestinationServiceFQDN(sourceCluster, svcRef)

	var destinationPort *networkingv1alpha3spec.PortSelector
	if port := kubeDest.GetPort(); port != 0 {
		if !trafficpolicyutils.ContainsPort(trafficShiftKubeService.Ports, port) {
			return nil, eris.Errorf("specified port %d does not exist for traffic shift destination service %v", port, sets.Key(trafficShiftKubeService.Ref))
		}
		destinationPort = &networkingv1alpha3spec.PortSelector{
			Number: port,
		}
	} else {
		// validate that traffic target only has one port
		if numPorts := len(trafficShiftKubeService.Ports); numPorts > 1 {
			return nil, eris.Errorf("must provide port for traffic shift destination service %v with multiple ports (%v) defined", sets.Key(trafficShiftKubeService.Ref), numPorts)
		}
	}

	httpRouteDestination := &networkingv1alpha3spec.HTTPRouteDestination{
		Destination: &networkingv1alpha3spec.Destination{
			Host: destinationHost,
			Port: destinationPort,
		},
		Weight: int32(weight),
	}

	if kubeDest.Subset != nil {
		// Use the canonical SMH unique name for this subset.
		httpRouteDestination.Destination.Subset = subsetName(kubeDest.Subset)
	}

	return httpRouteDestination, nil
}

func (d *trafficShiftDecorator) buildFailoverServiceDestination(
	failoverServiceDest *v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination_FailoverServiceDestination,
	weight uint32,
) (*networkingv1alpha3spec.HTTPRouteDestination, error) {
	failoverService, err := d.failoverServices.Find(ezkube.MakeObjectRef(failoverServiceDest))
	if err != nil {
		return nil, eris.Wrapf(err, "invalid traffic shift destination %s, FailoverService not found", sets.Key(failoverServiceDest))
	}

	httpRouteDestination := &networkingv1alpha3spec.HTTPRouteDestination{
		Destination: &networkingv1alpha3spec.Destination{
			Host: failoverService.Spec.Hostname,
			Port: &networkingv1alpha3spec.PortSelector{
				Number: failoverService.Spec.Port.Number,
			},
		},
		Weight: int32(weight),
	}

	if failoverServiceDest.Subset != nil {
		// Use the canonical SMH unique name for this subset.
		httpRouteDestination.Destination.Subset = subsetName(failoverServiceDest.Subset)
	}

	return httpRouteDestination, nil
}

// make all the necessary subsets for the destination rule for the given FailoverService.
// traverses all the applied traffic policies to find subsets matching this FailoverService
func MakeDestinationRuleSubsetsForFailoverService(
	failoverService *discoveryv1alpha2.MeshStatus_AppliedFailoverService,
	allTrafficTargets discoveryv1alpha2sets.TrafficTargetSet,
) []*networkingv1alpha3spec.Subset {
	return makeDestinationRuleSubsets(
		allTrafficTargets,
		func(weightedDestination *v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination) bool {
			failoverDestination := weightedDestination.GetFailoverService()
			if failoverDestination == nil {
				return false
			}
			return ezkube.RefsMatch(failoverService.Ref, failoverDestination)
		},
	)
}

// make all the necessary subsets for the destination rule for the given traffictarget.
// traverses all the applied traffic policies to find subsets matching this traffictarget
func MakeDestinationRuleSubsetsForTrafficTarget(
	trafficTarget *discoveryv1alpha2.TrafficTarget,
	allTrafficTargets discoveryv1alpha2sets.TrafficTargetSet,
	failoverServices v1alpha2sets.FailoverServiceSet,
) []*networkingv1alpha3spec.Subset {
	return makeDestinationRuleSubsets(
		allTrafficTargets,
		func(weightedDestination *v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination) bool {
			switch destType := weightedDestination.DestinationType.(type) {
			case *v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination_KubeService:
				return traffictargetutils.IsTrafficTargetForKubeService(trafficTarget, destType.KubeService)
			case *v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination_FailoverService:
				failoverService, err := failoverServices.Find(destType.FailoverService)
				if err != nil {
					return false
				}
				return failoverserviceutils.ContainsTrafficTarget(failoverService, trafficTarget)
			}
			return false
		},
	)
}

func makeDestinationRuleSubsets(
	allTrafficTargets discoveryv1alpha2sets.TrafficTargetSet,
	destinationMatchFunc func(weightedDestination *v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination) bool,
) []*networkingv1alpha3spec.Subset {
	var uniqueSubsets []map[string]string
	appendUniqueSubset := func(subsetLabels map[string]string) {
		if len(subsetLabels) == 0 {
			return
		}
		for _, subset := range uniqueSubsets {
			if reflect.DeepEqual(subset, subsetLabels) {
				return
			}
		}
		uniqueSubsets = append(uniqueSubsets, subsetLabels)
	}

	allTrafficTargets.List(func(service *discoveryv1alpha2.TrafficTarget) bool {
		for _, policy := range service.Status.AppliedTrafficPolicies {
			trafficShiftDestinations := policy.Spec.GetTrafficShift().GetDestinations()
			for _, dest := range trafficShiftDestinations {
				if destinationMatchFunc(dest) {
					switch destType := dest.DestinationType.(type) {
					case *v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination_KubeService:
						appendUniqueSubset(destType.KubeService.Subset)
					case *v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination_FailoverService:
						appendUniqueSubset(destType.FailoverService.Subset)
					}
				}
			}
		}
		return true
	})

	var subsets []*networkingv1alpha3spec.Subset
	for _, subsetLabels := range uniqueSubsets {
		subsets = append(subsets, &networkingv1alpha3spec.Subset{
			Name:   subsetName(subsetLabels),
			Labels: subsetLabels,
		})
	}

	return subsets
}

// used in DestinationRule translator as well
func subsetName(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}
	var keys []string
	for key, val := range labels {
		keys = append(keys, key+"-"+val)
	}
	sort.Strings(keys)
	return kubeutils.SanitizeNameV2(strings.Join(keys, "_"))
}
