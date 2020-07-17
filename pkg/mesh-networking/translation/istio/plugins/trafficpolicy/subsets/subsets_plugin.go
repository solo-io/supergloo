package subsets

import (
	"reflect"

	discoveryv1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	"github.com/solo-io/smh/pkg/mesh-networking/plugins"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/istio/plugins/trafficpolicy"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/istio/plugins/trafficpolicy/trafficshift"
	istiov1alpha3spec "istio.io/api/networking/v1alpha3"
)

const (
	pluginName = "subsets"
)

func init() {
	plugins.Register(pluginConstructor)
}

func pluginConstructor(_ plugins.Parameters) plugins.Plugin {
	return NewSubsetsPlugin()
}

// Handles setting subsets on a DestinationRule.
type subsetsPlugin struct{}

var _ trafficpolicy.DestinationRuleDecorator = &subsetsPlugin{}

func NewSubsetsPlugin() *subsetsPlugin {
	return &subsetsPlugin{}
}

func (s *subsetsPlugin) PluginName() string {
	return pluginName
}

func (s *subsetsPlugin) DecorateDestinationRule(
	appliedPolicy *discoveryv1alpha1.MeshServiceStatus_AppliedTrafficPolicy,
	_ *discoveryv1alpha1.MeshService,
	output *istiov1alpha3spec.DestinationRule,
	registerField plugins.RegisterField,
) error {
	subsets := s.translateSubset(appliedPolicy.GetSpec())
	if subsets != nil {
		if err := registerField(&output.Subsets, subsets); err != nil {
			return err
		}
		output.Subsets = subsets
	}
	return nil
}

func (s *subsetsPlugin) translateSubset(
	trafficPolicy *v1alpha1.TrafficPolicySpec,
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

	for _, destination := range trafficPolicy.GetTrafficShift().GetDestinations() {
		if subsetLabels := destination.GetKubeService().GetSubset(); len(subsetLabels) > 0 {
			appendUniqueSubset(subsetLabels)
		}
	}

	var subsets []*istiov1alpha3spec.Subset
	for _, subsetLabels := range uniqueSubsets {
		subsets = append(subsets, &istiov1alpha3spec.Subset{
			Name:   trafficshift.SubsetName(subsetLabels),
			Labels: subsetLabels,
		})
	}
	return subsets
}
