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
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/hostutils"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/meshserviceutils"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/trafficpolicyutils"
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
	return NewTrafficShiftDecorator(params.ClusterDomains, params.Snapshot.MeshServices())
}

var (
	MultiClusterSubsetsNotSupportedErr = func(dest ezkube.ResourceId) error {
		return eris.Errorf("Multi cluster subsets are currently not supported, found one on destination: %v", sets.Key(dest))
	}
)

// handles setting Weighted Destinations on a VirtualService
type trafficShiftDecorator struct {
	clusterDomains hostutils.ClusterDomainRegistry
	meshServices   discoveryv1alpha2sets.MeshServiceSet
}

var _ decorators.TrafficPolicyVirtualServiceDecorator = &trafficShiftDecorator{}

func NewTrafficShiftDecorator(
	clusterDomains hostutils.ClusterDomainRegistry,
	meshServices discoveryv1alpha2sets.MeshServiceSet,
) *trafficShiftDecorator {
	return &trafficShiftDecorator{
		clusterDomains: clusterDomains,
		meshServices:   meshServices,
	}
}

func (d *trafficShiftDecorator) DecoratorName() string {
	return decoratorName
}

func (d *trafficShiftDecorator) ApplyTrafficPolicyToVirtualService(
	appliedPolicy *discoveryv1alpha2.MeshServiceStatus_AppliedTrafficPolicy,
	service *discoveryv1alpha2.MeshService,
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
	meshService *discoveryv1alpha2.MeshService,
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
				meshService,
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
	originalService *discoveryv1alpha2.MeshService,
	weight uint32,
) (*networkingv1alpha3spec.HTTPRouteDestination, error) {
	originalKubeService := originalService.Spec.GetKubeService()

	if originalKubeService == nil {
		return nil, eris.Errorf("traffic shift only supported for kube mesh services")
	}
	if kubeDest == nil {
		return nil, eris.Errorf("nil kube destination on traffic shift")
	}

	svcRef := &v1.ClusterObjectRef{
		Name:        kubeDest.Name,
		Namespace:   kubeDest.Namespace,
		ClusterName: kubeDest.Cluster,
	}

	// validate destination service is a known meshservice
	trafficShiftService, err := meshserviceutils.FindMeshServiceForKubeService(d.meshServices.List(), svcRef)
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
		// validate that mesh service only has one port
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
		// cross-cluster subsets are currently unsupported, so return an error on the traffic policy
		if kubeDest.Cluster != sourceCluster {
			return nil, MultiClusterSubsetsNotSupportedErr(kubeDest)
		}

		// Use the canonical SMH unique name for this subset.
		httpRouteDestination.Destination.Subset = subsetName(kubeDest.Subset)
	}

	return httpRouteDestination, nil
}

// exposed for use in translators that initialize DestinationRules
func MakeDestinationRuleSubsets(
	appliedPolicies []*discoveryv1alpha2.MeshServiceStatus_AppliedTrafficPolicy,
) []*networkingv1alpha3spec.Subset {
	var uniqueSubsets []map[string]string
	appendUniqueSubset := func(subsetLabels map[string]string) {
		for _, subset := range uniqueSubsets {
			if reflect.DeepEqual(subset, subsetLabels) {
				return
			}
		}
		uniqueSubsets = append(uniqueSubsets, subsetLabels)
	}

	for _, policy := range appliedPolicies {
		for _, destination := range policy.GetSpec().GetTrafficShift().GetDestinations() {
			if subsetLabels := destination.GetKubeService().GetSubset(); len(subsetLabels) > 0 {
				appendUniqueSubset(subsetLabels)
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
	return kubeutils.SanitizeNameV2(strings.Join(keys, "_"))
}
