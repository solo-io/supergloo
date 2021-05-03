package trafficshift

import (
	"github.com/rotisserie/eris"
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	discoveryv1sets "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1/sets"
	v1beta1sets "github.com/solo-io/gloo-mesh/pkg/api/networking.enterprise.mesh.gloo.solo.io/v1beta1/sets"
	networkingv1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/hostutils"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/routeutils"
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

	originalKubeService := destination.Spec.GetKubeService()

	if originalKubeService == nil {
		return nil, eris.Errorf("traffic shift only supported for kube Destinations")
	}

	// An empty sourceClusterName indicates translation for VirtualService local to Destination
	if sourceClusterName == "" {
		sourceClusterName = destination.Spec.GetKubeService().GetRef().GetClusterName()
	}

	var shiftedDestinations []*networkingv1alpha3spec.HTTPRouteDestination
	for _, weightedDest := range trafficShift.Destinations {
		istioDestination, err := routeutils.TranslateWeightedDestination(
			weightedDest,
			sourceClusterName,
			d.destinations,
			d.virtualDestinations,
			d.clusterDomains,
		)
		if err != nil {
			return nil, err
		}
		shiftedDestinations = append(shiftedDestinations, istioDestination)

	}

	return shiftedDestinations, nil
}
