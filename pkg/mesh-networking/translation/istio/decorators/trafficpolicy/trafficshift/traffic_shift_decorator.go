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
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/decorators"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators/trafficpolicy"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/hostutils"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/meshserviceutils"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/solo-io/skv2/pkg/ezkube"
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
	meshServices   discoveryv1alpha2sets.MeshServiceSet
}

var _ trafficpolicy.VirtualServiceDecorator = &trafficShiftDecorator{}
var _ trafficpolicy.AggregatingDestinationRuleDecorator = &trafficShiftDecorator{}

func NewTrafficShiftDecorator(
	clusterDomains hostutils.ClusterDomainRegistry,
	meshServices discoveryv1alpha2sets.MeshServiceSet,
) *trafficShiftDecorator {
	return &trafficShiftDecorator{
		clusterDomains: clusterDomains,
		meshServices:   meshServices,
	}
}

func (t *trafficShiftDecorator) DecoratorName() string {
	return decoratorName
}

func (t *trafficShiftDecorator) ApplyToVirtualService(
	appliedPolicy *discoveryv1alpha2.MeshServiceStatus_AppliedTrafficPolicy,
	service *discoveryv1alpha2.MeshService,
	output *istiov1alpha3spec.HTTPRoute,
	registerField decorators.RegisterField,
) error {
	trafficShiftDestinations, err := t.translateTrafficShift(service, appliedPolicy.Spec)
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

func (t *trafficShiftDecorator) ApplyAllToDestinationRule(
	appliedPolicies []*discoveryv1alpha2.MeshServiceStatus_AppliedTrafficPolicy,
	output *istiov1alpha3spec.DestinationRule,
	registerField decorators.RegisterField,
) error {
	subsets := t.translateSubset(appliedPolicies)
	if subsets != nil {
		if err := registerField(&output.Subsets, subsets); err != nil {
			return err
		}
		output.Subsets = subsets
	}
	return nil
}

func (t *trafficShiftDecorator) translateTrafficShift(
	meshService *discoveryv1alpha2.MeshService,
	trafficPolicy *v1alpha2.TrafficPolicySpec,
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
		case *v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination_KubeService:
			var err error
			trafficShiftDestination, err = t.buildKubeTrafficShiftDestination(
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

func (t *trafficShiftDecorator) buildKubeTrafficShiftDestination(
	kubeDest *v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination_KubeDestination,
	originalService *discoveryv1alpha2.MeshService,
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
	if _, err := meshserviceutils.FindMeshServiceForKubeService(t.meshServices.List(), svcRef); err != nil {
		return nil, eris.Wrapf(err, "invalid mirror destination")
	}

	sourceCluster := originalKubeService.Ref.ClusterName
	destinationHost := t.clusterDomains.GetDestinationServiceFQDN(sourceCluster, svcRef)

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
		httpRouteDestination.Destination.Subset = subsetName(kubeDest.Subset)
	}

	return httpRouteDestination, nil
}

func (d *trafficShiftDecorator) translateSubset(
	appliedPolicies []*discoveryv1alpha2.MeshServiceStatus_AppliedTrafficPolicy,
) []*istiov1alpha3spec.Subset {
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

	var subsets []*istiov1alpha3spec.Subset
	for _, subsetLabels := range uniqueSubsets {
		subsets = append(subsets, &istiov1alpha3spec.Subset{
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
