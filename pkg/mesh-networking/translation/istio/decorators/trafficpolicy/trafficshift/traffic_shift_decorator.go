package trafficshift

import (
	"sort"
	"strings"

	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/kubeutils"
	discoveryv1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	discoveryv1alpha1sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/sets"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/solo-io/skv2/pkg/ezkube"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/decorators"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/istio/decorators/trafficpolicy"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/utils/hostutils"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/utils/meshserviceutils"
	istiov1alpha3spec "istio.io/api/networking/v1alpha3"
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
	meshServices   discoveryv1alpha1sets.MeshServiceSet
}

var _ trafficpolicy.VirtualServiceDecorator = &trafficShiftDecorator{}

func NewTrafficShiftDecorator(
	clusterDomains hostutils.ClusterDomainRegistry,
	meshServices discoveryv1alpha1sets.MeshServiceSet,
) *trafficShiftDecorator {
	return &trafficShiftDecorator{
		clusterDomains: clusterDomains,
		meshServices:   meshServices,
	}
}

func (p *trafficShiftDecorator) DecoratorName() string {
	return decoratorName
}

func (p *trafficShiftDecorator) DecorateVirtualService(
	appliedPolicy *discoveryv1alpha1.MeshServiceStatus_AppliedTrafficPolicy,
	service *discoveryv1alpha1.MeshService,
	output *istiov1alpha3spec.HTTPRoute,
	registerField decorators.RegisterField,
) error {
	trafficShiftDestinations, err := p.translateTrafficShift(service, appliedPolicy.Spec)
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

func (p *trafficShiftDecorator) translateTrafficShift(
	meshService *discoveryv1alpha1.MeshService,
	trafficPolicy *v1alpha1.TrafficPolicySpec,
) ([]*istiov1alpha3spec.HTTPRouteDestination, error) {
	trafficShift := trafficPolicy.GetTrafficShift()
	if trafficShift == nil {
		return nil, nil
	}

	var shiftedDestinations []*istiov1alpha3spec.HTTPRouteDestination
	for _, destination := range trafficShift.Destinations {
		if destination.DestinationType == nil {
			return nil, eris.Errorf("must set a destination type on traffic shift destination")
		}
		var trafficShiftDestination *istiov1alpha3spec.HTTPRouteDestination
		switch destinationType := destination.DestinationType.(type) {
		case *v1alpha1.TrafficPolicySpec_MultiDestination_WeightedDestination_KubeService:
			var err error
			trafficShiftDestination, err = p.buildKubeTrafficShiftDestination(
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

func (p *trafficShiftDecorator) buildKubeTrafficShiftDestination(
	kubeDest *v1alpha1.TrafficPolicySpec_MultiDestination_WeightedDestination_KubeDestination,
	originalService *discoveryv1alpha1.MeshService,
	weight uint32,
) (*istiov1alpha3spec.HTTPRouteDestination, error) {
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
	if _, err := meshserviceutils.FindMeshServiceForKubeService(p.meshServices.List(), svcRef); err != nil {
		return nil, eris.Wrapf(err, "invalid mirror destination")
	}

	sourceCluster := originalKubeService.Ref.ClusterName
	destinationHost := p.clusterDomains.GetDestinationServiceFQDN(sourceCluster, svcRef)

	var destinationPort *istiov1alpha3spec.PortSelector
	if port := kubeDest.GetPort(); port != 0 {
		destinationPort = &istiov1alpha3spec.PortSelector{
			Number: port,
		}
	} else {
		// validate that mesh service only has one port
		if numPorts := len(originalKubeService.Ports); numPorts > 1 {
			return nil, eris.Errorf("must provide port for traffic shift destination service %v with multiple ports (%v) defined", sets.Key(originalKubeService.Ref), numPorts)
		}
	}

	httpRouteDestination := &istiov1alpha3spec.HTTPRouteDestination{
		Destination: &istiov1alpha3spec.Destination{
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
		httpRouteDestination.Destination.Subset = SubsetName(kubeDest.Subset)
	}

	return httpRouteDestination, nil
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
	return kubeutils.SanitizeNameV2(strings.Join(keys, "_"))
}
