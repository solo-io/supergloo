package mirror

import (
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	discoveryv1sets "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1/sets"
	v1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/hostutils"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/trafficpolicyutils"
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
	appliedPolicy *v1.AppliedTrafficPolicy,
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
	// An empty sourceClusterName indicates translation for VirtualService local to Destination
	if sourceClusterName == "" {
		sourceClusterName = destination.Spec.GetKubeService().GetRef().GetClusterName()
	}

	return trafficpolicyutils.TranslateMirror(mirror, sourceClusterName, d.clusterDomains, d.destinations)
}
