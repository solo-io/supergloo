package mirror

import (
	"github.com/rotisserie/eris"
	discoveryv1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	discoveryv1alpha1sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/sets"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	"github.com/solo-io/smh/pkg/mesh-networking/translator/utils/hostutils"
	"github.com/solo-io/smh/pkg/mesh-networking/translator/utils/meshserviceutils"
	istiov1alpha3spec "istio.io/api/networking/v1alpha3"
	"reflect"
)

const (
	pluginName = "mirror"
)

// handles setting Mirror on a VirtualService
type mirrorPlugin struct {
	clusterDomains hostutils.ClusterDomainRegistry
	meshServices   discoveryv1alpha1sets.MeshServiceSet
}

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

func (p *mirrorPlugin) ProcessTrafficPolicy(trafficPolicySpec *v1alpha1.TrafficPolicySpec, meshService *discoveryv1alpha1.MeshService, output *istiov1alpha3spec.HTTPRoute) error {
	mirror, percentage, err := p.translateMirror(meshService, trafficPolicySpec)
	if err != nil {
		return err
	}
	if mirror != nil {
		if output.Mirror != nil && !reflect.DeepEqual(output.MirrorPercentage, mirror) {
			return eris.Errorf("mirroring was already defined by a previous traffic policy")
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
	mirror := trafficPolicy.GetMirror()
	if mirror == nil {
		return nil, nil, nil
	}
	destination := mirror.GetDestination()
	if destination == nil {
		return nil, nil, eris.Errorf("must provide mirror destination")
	}
	if _, err := meshserviceutils.FindMeshServiceForKubeService(p.meshServices.List(), destination); err != nil {
		return nil, nil, eris.Wrapf(err, "invalid mirror destination")
	}

	localCluster := meshService.Spec.GetKubeService().GetRef().GetClusterName()
	destinationHostname := p.clusterDomains.GetDestinationServiceFQDN(
		localCluster,
		destination,
	)

	translatedMirror := &istiov1alpha3spec.Destination{
		Host: destinationHostname,
	}

	if port := mirror.GetPort(); port != 0 {
		translatedMirror.Port = &istiov1alpha3spec.PortSelector{
			Number: port,
		}
	} else {
		// validate that mesh service only has one port
		if numPorts := len(meshService.Spec.KubeService.Ports); numPorts > 1 {
			return nil, nil, eris.Errorf("must provide port for mirror destination service %v with multiple ports (%v) defined", sets.Key(meshService.Spec.KubeService.Ref), numPorts)
		}
	}

	mirrorPercentage := &istiov1alpha3spec.Percent{
		Value: mirror.GetPercentage(),
	}

	return translatedMirror, mirrorPercentage, nil
}
