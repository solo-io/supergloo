package trafficshift

import (
	"github.com/rotisserie/eris"
	discoveryv1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	discoveryv1alpha1sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/sets"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	"github.com/solo-io/skv2/pkg/ezkube"
	"github.com/solo-io/smh/pkg/mesh-networking/translator/istio/destinationrule"
	"github.com/solo-io/smh/pkg/mesh-networking/translator/utils/hostutils"
	istiov1alpha3spec "istio.io/api/networking/v1alpha3"
	"reflect"
)

const (
	pluginName = "traffic-shift"
)

var (
	MultiClusterSubsetsNotSupportedErr = func(dest ezkube.ResourceId) error {
		return eris.Errorf("Multi cluster subsets are currently not supported, found one on destination: %v", sets.Key(dest))
	}
)

// handles setting Weighted Destinations on a VirtualService
type trafficShiftPlugin struct {
	clusterDomains hostutils.ClusterDomainRegistry
	meshServices   discoveryv1alpha1sets.MeshServiceSet
}

func NewTrafficShiftPlugin(
	clusterDomains hostutils.ClusterDomainRegistry,
	meshServices discoveryv1alpha1sets.MeshServiceSet,
) *trafficShiftPlugin {
	return &trafficShiftPlugin{
		clusterDomains: clusterDomains,
		meshServices:   meshServices,
	}
}

func (p *trafficShiftPlugin) PluginName() string {
	return pluginName
}

func (p *trafficShiftPlugin) ProcessTrafficPolicy(trafficPolicySpec *v1alpha1.TrafficPolicySpec, meshService *discoveryv1alpha1.MeshService, output *istiov1alpha3spec.HTTPRoute) error {
	trafficShiftDestinations, err := p.translateTrafficShift(meshService, trafficPolicySpec)
	if err != nil {
		return err
	}
	if trafficShiftDestinations != nil {
		if output.Route != nil && !reflect.DeepEqual(output.Route, trafficShiftDestinations) {
			return eris.Errorf("destinations already defined by a previous traffic policy")
		}
		output.Route = trafficShiftDestinations
	}
	return nil
}

func (p *trafficShiftPlugin) translateTrafficShift(
	meshService *discoveryv1alpha1.MeshService,
	trafficPolicy *v1alpha1.TrafficPolicySpec,
) ([]*istiov1alpha3spec.HTTPRouteDestination, error) {
	trafficShift := trafficPolicy.GetTrafficShift()
	if trafficShift == nil {
		return nil, nil
	}

	var shiftedDestinations []*istiov1alpha3spec.HTTPRouteDestination
	for _, destination := range trafficShift.Destinations {
		destinationRef := destination.Destination
		if destinationRef == nil {
			return nil, eris.Errorf("must provide destination ref on traffic shift")
		}

		sourceCluster := meshService.Spec.KubeService.Ref.ClusterName
		destinationHost := p.clusterDomains.GetDestinationServiceFQDN(sourceCluster, destinationRef)

		var destinationPort *istiov1alpha3spec.PortSelector
		if port := destination.GetPort(); port != 0 {
			destinationPort = &istiov1alpha3spec.PortSelector{
				Number: port,
			}
		} else {
			// validate that mesh service only has one port
			if numPorts := len(meshService.Spec.KubeService.Ports); numPorts > 1 {
				return nil, eris.Errorf("must provide port for traffic shift destination service %v with multiple ports (%v) defined", sets.Key(meshService.Spec.KubeService.Ref), numPorts)
			}
		}

		httpRouteDestination := &istiov1alpha3spec.HTTPRouteDestination{
			Destination: &istiov1alpha3spec.Destination{
				Host: destinationHost,
				Port: destinationPort,
			},
			Weight: int32(destination.GetWeight()),
		}

		if destination.Subset != nil {
			// cross-cluster subsets are currently unsupported, so return an error on the traffic policy
			if destinationRef.ClusterName != sourceCluster {
				return nil, MultiClusterSubsetsNotSupportedErr(destinationRef)
			}

			// Use the canonical SMH unique name for this subset.
			httpRouteDestination.Destination.Subset = destinationrule.SubsetName(destination.GetSubset())
		}
		shiftedDestinations = append(shiftedDestinations, httpRouteDestination)
	}

	return shiftedDestinations, nil
}
