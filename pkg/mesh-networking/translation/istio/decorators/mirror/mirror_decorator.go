package mirror

import (
	"github.com/rotisserie/eris"
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	discoveryv1sets "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1/sets"
	v1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/destinationutils"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/hostutils"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/trafficpolicyutils"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	networkingv1alpha3spec "istio.io/api/networking/v1alpha3"
)

const (
	decoratorName = "mirror"
)

func init() {
	decorators.Register(decoratorConstructor)
}

func decoratorConstructor(params decorators.Parameters) decorators.Decorator {
	return NewMirrorDecorator(params.ClusterDomains, params.Snapshot.Destinations())
}

// handles setting Mirror on a VirtualService
type mirrorDecorator struct {
	clusterDomains hostutils.ClusterDomainRegistry
	destinations   discoveryv1sets.DestinationSet
}

var _ decorators.TrafficPolicyVirtualServiceDecorator = &mirrorDecorator{}

func NewMirrorDecorator(
	clusterDomains hostutils.ClusterDomainRegistry,
	destinations discoveryv1sets.DestinationSet,
) *mirrorDecorator {
	return &mirrorDecorator{
		clusterDomains: clusterDomains,
		destinations:   destinations,
	}
}

func (d *mirrorDecorator) DecoratorName() string {
	return decoratorName
}

func (d *mirrorDecorator) ApplyTrafficPolicyToVirtualService(
	appliedPolicy *discoveryv1.DestinationStatus_AppliedTrafficPolicy,
	destination *discoveryv1.Destination,
	sourceMeshInstallation *discoveryv1.MeshInstallation,
	output *networkingv1alpha3spec.HTTPRoute,
	registerField decorators.RegisterField,
) error {
	mirror, percentage, err := d.translateMirror(destination, appliedPolicy.Spec, sourceMeshInstallation.GetCluster())
	if err != nil {
		return err
	}
	if mirror != nil {
		if err := registerField(&output.Mirror, mirror); err != nil {
			return err
		}
		output.Mirror = mirror
		output.MirrorPercentage = percentage
	}
	return nil
}

// If federatedClusterName is non-empty, it indicates translation for a federated VirtualService, so use it as the source cluster name.
func (d *mirrorDecorator) translateMirror(
	destination *discoveryv1.Destination,
	trafficPolicy *v1.TrafficPolicySpec,
	sourceClusterName string,
) (*networkingv1alpha3spec.Destination, *networkingv1alpha3spec.Percent, error) {
	mirror := trafficPolicy.GetPolicy().GetMirror()
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
		translatedMirror, err = d.makeKubeDestinationMirror(
			destinationType,
			mirror.Port,
			destination,
			sourceClusterName,
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

func (d *mirrorDecorator) makeKubeDestinationMirror(
	mirrorDest *v1.TrafficPolicySpec_Policy_Mirror_KubeService,
	port uint32,
	destination *discoveryv1.Destination,
	sourceClusterName string,
) (*networkingv1alpha3spec.Destination, error) {
	destinationRef := mirrorDest.KubeService
	mirrorService, err := destinationutils.FindDestinationForKubeService(d.destinations.List(), destinationRef)
	if err != nil {
		return nil, eris.Wrapf(err, "invalid mirror destination")
	}
	mirrorKubeService := mirrorService.Spec.GetKubeService()

	// TODO(ilackarms): support other types of Destination destinations, e.g. via ServiceEntries

	// An empty sourceClusterName indicates translation for VirtualService local to Destination
	if sourceClusterName == "" {
		sourceClusterName = destination.Spec.GetKubeService().GetRef().GetClusterName()
	}

	destinationHostname := d.clusterDomains.GetDestinationFQDN(
		sourceClusterName,
		destinationRef,
	)

	translatedMirror := &networkingv1alpha3spec.Destination{
		Host: destinationHostname,
	}

	if port != 0 {
		if !trafficpolicyutils.ContainsPort(mirrorKubeService.Ports, port) {
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
