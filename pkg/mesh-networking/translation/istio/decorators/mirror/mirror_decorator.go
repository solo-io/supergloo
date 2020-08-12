package mirror

import (
	"github.com/rotisserie/eris"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	discoveryv1alpha2sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2/sets"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/hostutils"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/meshserviceutils"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/trafficpolicyutils"
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
	return NewMirrorDecorator(params.ClusterDomains, params.Snapshot.MeshServices())
}

// handles setting Mirror on a VirtualService
type mirrorDecorator struct {
	clusterDomains hostutils.ClusterDomainRegistry
	meshServices   discoveryv1alpha2sets.MeshServiceSet
}

var _ decorators.TrafficPolicyVirtualServiceDecorator = &mirrorDecorator{}

func NewMirrorDecorator(
	clusterDomains hostutils.ClusterDomainRegistry,
	meshServices discoveryv1alpha2sets.MeshServiceSet,
) *mirrorDecorator {
	return &mirrorDecorator{
		clusterDomains: clusterDomains,
		meshServices:   meshServices,
	}
}

func (d *mirrorDecorator) DecoratorName() string {
	return decoratorName
}

func (d *mirrorDecorator) ApplyTrafficPolicyToVirtualService(
	appliedPolicy *discoveryv1alpha2.MeshServiceStatus_AppliedTrafficPolicy,
	service *discoveryv1alpha2.MeshService,
	output *networkingv1alpha3spec.HTTPRoute,
	registerField decorators.RegisterField,
) error {
	mirror, percentage, err := d.translateMirror(service, appliedPolicy.Spec)
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

func (d *mirrorDecorator) translateMirror(
	meshService *discoveryv1alpha2.MeshService,
	trafficPolicy *v1alpha2.TrafficPolicySpec,
) (*networkingv1alpha3spec.Destination, *networkingv1alpha3spec.Percent, error) {
	mirror := trafficPolicy.Mirror
	if mirror == nil {
		return nil, nil, nil
	}
	if mirror.DestinationType == nil {
		return nil, nil, eris.Errorf("must provide mirror destination")
	}

	var translatedMirror *networkingv1alpha3spec.Destination
	switch destinationType := mirror.DestinationType.(type) {
	case *v1alpha2.TrafficPolicySpec_Mirror_KubeService:
		var err error
		translatedMirror, err = d.makeKubeDestinationMirror(
			destinationType,
			mirror.Port,
			meshService,
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
	destination *v1alpha2.TrafficPolicySpec_Mirror_KubeService,
	port uint32,
	originalService *discoveryv1alpha2.MeshService,
) (*networkingv1alpha3spec.Destination, error) {
	destinationRef := destination.KubeService
	mirrorService, err := meshserviceutils.FindMeshServiceForKubeService(d.meshServices.List(), destinationRef)
	if err != nil {
		return nil, eris.Wrapf(err, "invalid mirror destination")
	}
	mirrorKubeService := mirrorService.Spec.GetKubeService()

	// TODO(ilackarms): support other types of MeshService destinations, e.g. via ServiceEntries
	localCluster := originalService.Spec.GetKubeService().GetRef().GetClusterName()
	destinationHostname := d.clusterDomains.GetDestinationServiceFQDN(
		localCluster,
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
		// validate that mesh service only has one port
		if numPorts := len(mirrorKubeService.GetPorts()); numPorts > 1 {
			return nil, eris.Errorf("must provide port for mirror destination service %v with multiple ports (%v) defined", sets.Key(mirrorKubeService.GetRef()), numPorts)
		}
	}

	return translatedMirror, nil
}
