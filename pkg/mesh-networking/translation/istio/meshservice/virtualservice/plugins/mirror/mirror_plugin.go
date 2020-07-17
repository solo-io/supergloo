package mirror

import (
	"github.com/rotisserie/eris"
	discoveryv1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	discoveryv1alpha1sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/sets"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	"github.com/solo-io/smh/pkg/mesh-networking/plugins"
	virtualserviceplugins "github.com/solo-io/smh/pkg/mesh-networking/translation/istio/meshservice/virtualservice/plugins"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/utils/hostutils"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/utils/meshserviceutils"
	istiov1alpha3spec "istio.io/api/networking/v1alpha3"
)

const (
	pluginName = "mirror"
)

func init() {
	plugins.Register(pluginConstructor)
}

func pluginConstructor(params plugins.Parameters) plugins.Plugin {
	return NewMirrorPlugin(params.ClusterDomains, params.Snapshot.MeshServices())
}

// handles setting Mirror on a VirtualService
type mirrorPlugin struct {
	clusterDomains hostutils.ClusterDomainRegistry
	meshServices   discoveryv1alpha1sets.MeshServiceSet
}

var _ virtualserviceplugins.TrafficPolicyPlugin = &mirrorPlugin{}

func NewMirrorPlugin(
	clusterDomains hostutils.ClusterDomainRegistry,
	meshServices discoveryv1alpha1sets.MeshServiceSet,
) *mirrorPlugin {
	return &mirrorPlugin{
		clusterDomains: clusterDomains,
		meshServices:   meshServices,
	}
}

func (p *mirrorPlugin) PluginName() string {
	return pluginName
}

func (p *mirrorPlugin) ProcessTrafficPolicy(
	appliedPolicy *discoveryv1alpha1.MeshServiceStatus_AppliedTrafficPolicy,
	service *discoveryv1alpha1.MeshService,
	output *istiov1alpha3spec.HTTPRoute,
	registerField plugins.RegisterField,
) error {
	mirror, percentage, err := p.translateMirror(service, appliedPolicy.GetSpec())
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

func (p *mirrorPlugin) translateMirror(
	meshService *discoveryv1alpha1.MeshService,
	trafficPolicy *v1alpha1.TrafficPolicySpec,
) (*istiov1alpha3spec.Destination, *istiov1alpha3spec.Percent, error) {
	mirror := trafficPolicy.Mirror
	if mirror == nil {
		return nil, nil, nil
	}
	if mirror.DestinationType == nil {
		return nil, nil, eris.Errorf("must provide mirror destination")
	}

	var translatedMirror *istiov1alpha3spec.Destination
	switch destinationType := mirror.DestinationType.(type) {
	case *v1alpha1.TrafficPolicySpec_Mirror_KubeService:
		var err error
		translatedMirror, err = p.makeKubeDestinationMirror(
			destinationType,
			mirror.Port,
			meshService,
		)
		if err != nil {
			return nil, nil, err
		}
	}

	mirrorPercentage := &istiov1alpha3spec.Percent{
		Value: mirror.GetPercentage(),
	}

	return translatedMirror, mirrorPercentage, nil
}

func (p *mirrorPlugin) makeKubeDestinationMirror(
	destination *v1alpha1.TrafficPolicySpec_Mirror_KubeService,
	port uint32,
	originalService *discoveryv1alpha1.MeshService,
) (*istiov1alpha3spec.Destination, error) {

	destinationRef := destination.KubeService
	if _, err := meshserviceutils.FindMeshServiceForKubeService(p.meshServices.List(), destinationRef); err != nil {
		return nil, eris.Wrapf(err, "invalid mirror destination")
	}

	// TODO(ilackarms): support other types of MeshService destinations, e.g. via ServiceEntries
	localCluster := originalService.Spec.GetKubeService().GetRef().GetClusterName()
	destinationHostname := p.clusterDomains.GetDestinationServiceFQDN(
		localCluster,
		destinationRef,
	)

	translatedMirror := &istiov1alpha3spec.Destination{
		Host: destinationHostname,
	}

	if port != 0 {
		translatedMirror.Port = &istiov1alpha3spec.PortSelector{
			Number: port,
		}
	} else {
		// validate that mesh service only has one port
		if numPorts := len(originalService.Spec.GetKubeService().GetPorts()); numPorts > 1 {
			return nil, eris.Errorf("must provide port for mirror destination service %v with multiple ports (%v) defined", sets.Key(originalService.Spec.GetKubeService().GetRef()), numPorts)
		}
	}

	return translatedMirror, nil
}
