package trafficshift

import (
	"github.com/rotisserie/eris"
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	discoveryv1sets "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1/sets"
	v1beta1sets "github.com/solo-io/gloo-mesh/pkg/api/networking.enterprise.mesh.gloo.solo.io/v1beta1/sets"
	networkingv1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/destinationutils"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/hostutils"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/routeutils"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/trafficpolicyutils"
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
	appliedPolicy *networkingv1.AppliedTrafficPolicy,
	destination *discoveryv1.Destination,
	sourceMeshInstallation *discoveryv1.MeshInstallation,
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
		case *networkingv1.WeightedDestination_KubeService:
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
		case *networkingv1.WeightedDestination_VirtualDestination_:
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
	kubeDest *networkingv1.WeightedDestination_KubeDestination,
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
		httpRouteDestination.Destination.Subset = routeutils.SubsetName(kubeDest.Subset)
	}

	return httpRouteDestination, nil
}

func (d *trafficShiftDecorator) buildVirtualDestinationDestination(
	virtualDestinationDest *networkingv1.WeightedDestination_VirtualDestination,
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
		httpRouteDestination.Destination.Subset = routeutils.SubsetName(subset)
	}

	return httpRouteDestination
}
