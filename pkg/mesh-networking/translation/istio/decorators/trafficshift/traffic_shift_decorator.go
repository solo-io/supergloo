package trafficshift

import (
	"reflect"
	"sort"
	"strings"

	"github.com/rotisserie/eris"
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	discoveryv1sets "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1/sets"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.enterprise.mesh.gloo.solo.io/v1beta1"
	v1beta1sets "github.com/solo-io/gloo-mesh/pkg/api/networking.enterprise.mesh.gloo.solo.io/v1beta1/sets"
	networkingv1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/destinationutils"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/hostutils"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/trafficpolicyutils"
	"github.com/solo-io/k8s-utils/kubeutils"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	skv2corev1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
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
		params.Snapshot.Destinations(),
		params.Snapshot.VirtualDestinations(),
	)
}

// handles setting Weighted Destinations on a VirtualService
type trafficShiftDecorator struct {
	clusterDomains      hostutils.ClusterDomainRegistry
	virtualDestinations v1beta1sets.VirtualDestinationSet
	destinations        discoveryv1sets.DestinationSet
}

var _ decorators.TrafficPolicyVirtualServiceDecorator = &trafficShiftDecorator{}

func NewTrafficShiftDecorator(
	clusterDomains hostutils.ClusterDomainRegistry,
	destinations discoveryv1sets.DestinationSet,
	virtualDestinations v1beta1sets.VirtualDestinationSet,
) *trafficShiftDecorator {
	return &trafficShiftDecorator{
		clusterDomains:      clusterDomains,
		destinations:        destinations,
		virtualDestinations: virtualDestinations,
	}
}

func (d *trafficShiftDecorator) DecoratorName() string {
	return decoratorName
}

func (d *trafficShiftDecorator) ApplyTrafficPolicyToVirtualService(
	appliedPolicy *discoveryv1.DestinationStatus_AppliedTrafficPolicy,
	destination *discoveryv1.Destination,
	sourceMeshInstallation *discoveryv1.MeshSpec_MeshInstallation,
	output *networkingv1alpha3spec.HTTPRoute,
	registerField decorators.RegisterField,
) error {
	trafficShiftDestinations, err := d.translateTrafficShift(destination, appliedPolicy.Spec, sourceMeshInstallation.GetCluster())
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
	destination *discoveryv1.Destination,
	trafficPolicy *networkingv1.TrafficPolicySpec,
	sourceClusterName string,
) ([]*networkingv1alpha3spec.HTTPRouteDestination, error) {
	trafficShift := trafficPolicy.GetPolicy().GetTrafficShift()
	if trafficShift == nil {
		return nil, nil
	}

	var shiftedDestinations []*networkingv1alpha3spec.HTTPRouteDestination
	for _, weightedDest := range trafficShift.Destinations {
		if weightedDest.DestinationType == nil {
			return nil, eris.Errorf("must set a destination type on traffic shift destination")
		}
		var trafficShiftDestination *networkingv1alpha3spec.HTTPRouteDestination
		switch destinationType := weightedDest.DestinationType.(type) {
		case *networkingv1.TrafficPolicySpec_Policy_MultiDestination_WeightedDestination_KubeService:
			var err error
			trafficShiftDestination, err = d.buildKubeTrafficShiftDestination(
				destinationType.KubeService,
				destination,
				weightedDest.Weight,
				sourceClusterName,
			)
			if err != nil {
				return nil, err
			}
		case *networkingv1.TrafficPolicySpec_Policy_MultiDestination_WeightedDestination_VirtualDestination:
			var err error
			trafficShiftDestination, err = d.buildVirtualDestinationDestination(
				destinationType.VirtualDestination,
				weightedDest.Weight,
			)
			if err != nil {
				return nil, err
			}
		default:
			return nil, eris.Errorf("unsupported traffic shift destination type: %T", weightedDest.DestinationType)
		}
		shiftedDestinations = append(shiftedDestinations, trafficShiftDestination)

	}

	return shiftedDestinations, nil
}

func (d *trafficShiftDecorator) buildKubeTrafficShiftDestination(
	kubeDest *networkingv1.TrafficPolicySpec_Policy_MultiDestination_WeightedDestination_KubeDestination,
	destination *discoveryv1.Destination,
	weight uint32,
	sourceClusterName string,
) (*networkingv1alpha3spec.HTTPRouteDestination, error) {
	originalKubeService := destination.Spec.GetKubeService()

	if originalKubeService == nil {
		return nil, eris.Errorf("traffic shift only supported for kube Destinations")
	}
	if kubeDest == nil {
		return nil, eris.Errorf("nil kube destination on traffic shift")
	}

	svcRef := &skv2corev1.ClusterObjectRef{
		Name:        kubeDest.Name,
		Namespace:   kubeDest.Namespace,
		ClusterName: kubeDest.ClusterName,
	}

	// validate destination service is a known destination
	trafficShiftService, err := destinationutils.FindDestinationForKubeService(d.destinations.List(), svcRef)
	if err != nil {
		return nil, eris.Wrapf(err, "invalid traffic shift destination %s", sets.Key(svcRef))
	}
	trafficShiftKubeService := trafficShiftService.Spec.GetKubeService()

	// An empty sourceClusterName indicates translation for VirtualService local to Destination
	if sourceClusterName == "" {
		sourceClusterName = destination.Spec.GetKubeService().GetRef().GetClusterName()
	}

	destinationHost := d.clusterDomains.GetDestinationFQDN(sourceClusterName, svcRef)

	var destinationPort *networkingv1alpha3spec.PortSelector
	if port := kubeDest.GetPort(); port != 0 {
		if !trafficpolicyutils.ContainsPort(trafficShiftKubeService.Ports, port) {
			return nil, eris.Errorf("specified port %d does not exist for traffic shift destination service %v", port, sets.Key(trafficShiftKubeService.Ref))
		}
		destinationPort = &networkingv1alpha3spec.PortSelector{
			Number: port,
		}
	} else {
		// validate that Destination only has one port
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
		// Use the canonical GlooMesh unique name for this subset.
		httpRouteDestination.Destination.Subset = subsetName(kubeDest.Subset)
	}

	return httpRouteDestination, nil
}

func (d *trafficShiftDecorator) buildVirtualDestinationDestination(
	virtualDestinationDest *networkingv1.TrafficPolicySpec_Policy_MultiDestination_WeightedDestination_VirtualDestinationReference,
	weight uint32,
) (*networkingv1alpha3spec.HTTPRouteDestination, error) {
	virtualDestination, err := d.virtualDestinations.Find(ezkube.MakeObjectRef(virtualDestinationDest))
	if err != nil {
		return nil, eris.Wrapf(err, "invalid traffic shift destination %s, VirtualDestination not found", sets.Key(virtualDestinationDest))
	}

	httpRouteDestination := d.buildHttpRouteDestination(
		virtualDestination.Spec.GetHostname(),
		virtualDestinationDest.GetSubset(),
		virtualDestination.Spec.GetPort().GetNumber(),
		weight,
	)

	return httpRouteDestination, nil
}

func (d *trafficShiftDecorator) buildHttpRouteDestination(
	hostname string,
	subset map[string]string,
	portNumber, weight uint32,
) *networkingv1alpha3spec.HTTPRouteDestination {
	httpRouteDestination := &networkingv1alpha3spec.HTTPRouteDestination{
		Destination: &networkingv1alpha3spec.Destination{
			Host: hostname,
			Port: &networkingv1alpha3spec.PortSelector{
				Number: portNumber,
			},
		},
		Weight: int32(weight),
	}

	if subset != nil {
		// Use the canonical GlooMesh unique name for this subset.
		httpRouteDestination.Destination.Subset = subsetName(subset)
	}

	return httpRouteDestination
}

// make all the necessary subsets for the destination rule for the given VirtualDestination.
// traverses all the applied traffic policies to find subsets matching this VirtualDestination
func MakeDestinationRuleSubsetsForVirtualDestination(
	virtualDestination *v1beta1.VirtualDestination,
	allTrafficTargets discoveryv1sets.DestinationSet,
) []*networkingv1alpha3spec.Subset {
	return makeDestinationRuleSubsets(
		allTrafficTargets,
		func(weightedDestination *networkingv1.TrafficPolicySpec_Policy_MultiDestination_WeightedDestination) bool {
			virtualDestinationDest := weightedDestination.GetVirtualDestination()
			if virtualDestinationDest == nil {
				return false
			}
			return ezkube.RefsMatch(ezkube.MakeObjectRef(virtualDestination), virtualDestinationDest)
		},
	)
}

// make all the necessary subsets for the destination rule for the given destination.
// traverses all the applied traffic policies to find subsets matching this destination
func MakeDestinationRuleSubsetsForDestination(
	destination *discoveryv1.Destination,
	allDestinations discoveryv1sets.DestinationSet,
	sourceClusterName string,
) []*networkingv1alpha3spec.Subset {
	subsets := makeDestinationRuleSubsets(
		allDestinations,
		func(weightedDestination *networkingv1.TrafficPolicySpec_Policy_MultiDestination_WeightedDestination) bool {
			switch destType := weightedDestination.DestinationType.(type) {
			case *networkingv1.TrafficPolicySpec_Policy_MultiDestination_WeightedDestination_KubeService:
				return destinationutils.IsDestinationForKubeService(destination, destType.KubeService)
			}
			return false
		},
	)

	// NOTE(ilackarms): we make subsets here for the client-side destination rule for a federated Destination,
	// which contain all the matching subset names for the remote destination rule.
	// the labels for the subsets must match the labels on the ServiceEntry Endpoint(s).
	// Based on https://istio.io/latest/blog/2019/multicluster-version-routing/#create-a-destination-rule-on-both-clusters-for-the-local-reviews-service
	//
	// If flat-networking is enabled, we leave the subset info as there is no ingress involved
	if sourceClusterName != "" &&
		sourceClusterName != destination.ClusterName &&
		!destination.Status.GetAppliedFederation().GetFlatNetwork() {
		for _, subset := range subsets {
			// only the name of the subset matters here.
			// the labels must match those on the ServiceEntry's endpoints.
			subset.Labels = MakeFederatedSubsetLabel(destination.Spec.GetKubeService().Ref.ClusterName)
			// we also remove the TrafficPolicy, leaving
			// it to the server-side DestinationRule to enforce.
			subset.TrafficPolicy = nil
		}
	}

	return subsets
}

// clusterName corresponds to the cluster name for the federated Destination.
//
// NOTE(ilackarms): we use these labels to support federated subsets.
// the values don't actually matter; but the subset names should
// match those on the DestinationRule for the Destination in the
// remote cluster.
// based on: https://istio.io/latest/blog/2019/multicluster-version-routing/#create-a-destination-rule-on-both-clusters-for-the-local-reviews-service
func MakeFederatedSubsetLabel(clusterName string) map[string]string {
	return map[string]string{
		"cluster": clusterName,
	}
}

func makeDestinationRuleSubsets(
	allDestinations discoveryv1sets.DestinationSet,
	destinationMatchFunc func(weightedDestination *networkingv1.TrafficPolicySpec_Policy_MultiDestination_WeightedDestination) bool,
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

	// TODO(yuval-k): Once exposed, use a function that provides iteration without sorting
	// and coping
	for _, genericSvc := range allDestinations.Generic().Map() {
		service := genericSvc.(*discoveryv1.Destination)
		for _, policy := range service.Status.AppliedTrafficPolicies {
			trafficShiftDestinations := policy.Spec.GetPolicy().GetTrafficShift().GetDestinations()
			for _, dest := range trafficShiftDestinations {
				if destinationMatchFunc(dest) {
					switch destType := dest.DestinationType.(type) {
					case *networkingv1.TrafficPolicySpec_Policy_MultiDestination_WeightedDestination_KubeService:
						appendUniqueSubset(destType.KubeService.Subset)
					}
				}
			}
		}
	}

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
	return kubeutils.SanitizeNameV2(strings.Join(keys, "-"))
}
