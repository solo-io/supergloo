package routeutils

import (
	"reflect"
	"sort"
	"strings"

	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/destinationutils"

	"github.com/rotisserie/eris"
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	discoveryv1sets "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1/sets"
	v1beta1sets "github.com/solo-io/gloo-mesh/pkg/api/networking.enterprise.mesh.gloo.solo.io/v1beta1/sets"
	networkingv1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/hostutils"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/trafficpolicyutils"
	"github.com/solo-io/k8s-utils/kubeutils"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	skv2corev1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/solo-io/skv2/pkg/ezkube"
	networkingv1alpha3spec "istio.io/api/networking/v1alpha3"
)

func TranslateWeightedDestination(
	weightedDest *networkingv1.WeightedDestination,
	sourceCluster string,
	destinations discoveryv1sets.DestinationSet,
	virtualDestinations v1beta1sets.VirtualDestinationSet,
	clusterDomains hostutils.ClusterDomainRegistry,
) (*networkingv1alpha3spec.HTTPRouteDestination, error) {
	if weightedDest.DestinationType == nil {
		return nil, eris.Errorf("must set a destination type on weighted destination")
	}

	destinationHost, destinationPort, subsetName, err := resolveHostPortSubset(
		weightedDest,
		destinations,
		virtualDestinations,
		sourceCluster,
		clusterDomains,
	)
	if err != nil {
		return nil, err
	}

	return &networkingv1alpha3spec.HTTPRouteDestination{
		Destination: &networkingv1alpha3spec.Destination{
			Host:   destinationHost,
			Port:   destinationPort,
			Subset: subsetName,
		},
		Weight:  int32(weightedDest.Weight),
		Headers: trafficpolicyutils.TranslateHeaderManipulation(weightedDest.Options.GetHeaderManipulation()),
	}, nil
}

func resolveHostPortSubset(
	weightedDest *networkingv1.WeightedDestination,
	destinations discoveryv1sets.DestinationSet,
	virtualDestinations v1beta1sets.VirtualDestinationSet,
	sourceCluster string,
	clusterDomains hostutils.ClusterDomainRegistry,
) (string, *networkingv1alpha3spec.PortSelector, string, error) {
	switch destinationType := weightedDest.DestinationType.(type) {
	case *networkingv1.WeightedDestination_KubeService:
		return resolveHostPortSubsetKubeService(
			destinationType.KubeService,
			destinations,
			sourceCluster,
			clusterDomains,
		)
	case *networkingv1.WeightedDestination_VirtualDestination_:
		return resolveHostPortSubsetVirtualDestination(
			destinationType.VirtualDestination,
			virtualDestinations,
		)
	// TODO: Static
	// TODO: ClusterHeader
	default:
		return "", nil, "", eris.Errorf("unsupported destination type %T", destinationType)
	}
}

func resolveHostPortSubsetKubeService(
	kubeDest *networkingv1.WeightedDestination_KubeDestination,
	destinations discoveryv1sets.DestinationSet,
	sourceCluster string,
	clusterDomains hostutils.ClusterDomainRegistry,
) (string, *networkingv1alpha3spec.PortSelector, string, error) {

	svcRef := &skv2corev1.ClusterObjectRef{
		Name:        kubeDest.Name,
		Namespace:   kubeDest.Namespace,
		ClusterName: kubeDest.ClusterName,
	}

	// validate destination service is a known destination
	targetedDestination, err := destinationutils.FindDestinationForKubeService(destinations.List(), svcRef)
	if err != nil {
		return "", nil, "", eris.Wrapf(err, "invalid traffic shift destination %s", sets.Key(svcRef))
	}
	trafficShiftKubeService := targetedDestination.Spec.GetKubeService()

	destinationHost := clusterDomains.GetDestinationFQDN(sourceCluster, svcRef)

	var destinationPort *networkingv1alpha3spec.PortSelector
	if port := kubeDest.GetPort(); port != 0 {
		if !trafficpolicyutils.ContainsPort(trafficShiftKubeService.Ports, port) {
			return "", nil, "", eris.Errorf("specified port %d does not exist for traffic shift destination service %v", port, sets.Key(trafficShiftKubeService.Ref))
		}
		destinationPort = &networkingv1alpha3spec.PortSelector{
			Number: port,
		}
	} else {
		// validate that Destination only has one port
		if numPorts := len(trafficShiftKubeService.Ports); numPorts > 1 {
			return "", nil, "", eris.Errorf("must provide port for traffic shift destination service %v with multiple ports (%v) defined", sets.Key(trafficShiftKubeService.Ref), numPorts)
		}
	}

	var subsetName string
	if kubeDest.GetSubset() != nil {
		subsetName = SubsetName(kubeDest.GetSubset())
	}

	return destinationHost, destinationPort, subsetName, nil
}

func resolveHostPortSubsetVirtualDestination(
	virtualDest *networkingv1.WeightedDestination_VirtualDestination,
	virtualDestinations v1beta1sets.VirtualDestinationSet,
) (string, *networkingv1alpha3spec.PortSelector, string, error) {
	virtualDestination, err := virtualDestinations.Find(ezkube.MakeObjectRef(virtualDest))
	if err != nil {
		return "", nil, "", eris.Wrapf(err, "invalid traffic shift destination %s, VirtualDestination not found", sets.Key(virtualDest))
	}

	hostname := virtualDestination.Spec.Hostname
	port := &networkingv1alpha3spec.PortSelector{
		Number: virtualDestination.Spec.GetPort().GetNumber(),
	}
	var subsetName string
	if virtualDest.GetSubset() != nil {
		subsetName = SubsetName(virtualDest.GetSubset())
	}

	return hostname, port, subsetName, nil
}

// used in DestinationRule translator as well
func SubsetName(labels map[string]string) string {
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

// clusterName corresponds to the cluster name for the federated Destination.
//
// NOTE(ilackarms): we use these labels to support federated subsets.
// the values don't actually matter; but the subset names should
// match those on the DestinationRule for the Destination in the
// remote cluster.
// based on: https://istio.io/latest/blog/2019/multicluster-version-routing/#create-a-destination-rule-on-both-clusters-for-the-local-reviews-service
// exported for use in enterprise
func MakeFederatedSubsetLabel(clusterName string) map[string]string {
	return map[string]string{
		"cluster": clusterName,
	}
}

func makeDestinationRuleSubsets(
	requiredSubsets []*discoveryv1.DestinationStatus_RequiredSubsets,
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

	for _, requiredSubset := range requiredSubsets {
		for _, dest := range requiredSubset.TrafficShift.Destinations {
			switch destType := dest.DestinationType.(type) {
			case *networkingv1.WeightedDestination_KubeService:
				appendUniqueSubset(destType.KubeService.Subset)
			case *networkingv1.WeightedDestination_VirtualDestination_:
				appendUniqueSubset(destType.VirtualDestination.Subset)
			}
		}
	}

	var subsets []*networkingv1alpha3spec.Subset
	for _, subsetLabels := range uniqueSubsets {
		subsets = append(subsets, &networkingv1alpha3spec.Subset{
			Name:   SubsetName(subsetLabels),
			Labels: subsetLabels,
		})
	}

	return subsets
}
