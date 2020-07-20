package subsets

import (
	"reflect"

	discoveryv1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/decorators"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/istio/decorators/trafficpolicy"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/istio/decorators/trafficpolicy/trafficshift"
	istiov1alpha3spec "istio.io/api/networking/v1alpha3"
)

const (
	decoratorName = "subsets"
)

func init() {
	decorators.Register(decoratorConstructor)
}

func decoratorConstructor(_ decorators.Parameters) decorators.Decorator {
	return NewSubsetsDecorator()
}

// Handles setting subsets on a DestinationRule.
type subsetsDecorator struct{}

var _ trafficpolicy.AggregatingDestinationRuleDecorator = &subsetsDecorator{}

func NewSubsetsDecorator() *subsetsDecorator {
	return &subsetsDecorator{}
}

func (s *subsetsDecorator) DecoratorName() string {
	return decoratorName
}

func (s *subsetsDecorator) ApplyAllTrafficPolicies(
	appliedPolicies []*discoveryv1alpha1.MeshServiceStatus_AppliedTrafficPolicy,
	output *istiov1alpha3spec.DestinationRule,
	registerField decorators.RegisterField,
) error {
	subsets := s.translateSubset(appliedPolicies)
	if subsets != nil {
		if err := registerField(&output.Subsets, subsets); err != nil {
			return err
		}
		output.Subsets = subsets
	}
	return nil
}

func (s *subsetsDecorator) translateSubset(
	appliedPolicies []*discoveryv1alpha1.MeshServiceStatus_AppliedTrafficPolicy,
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
			Name:   trafficshift.SubsetName(subsetLabels),
			Labels: subsetLabels,
		})
	}

	return subsets
}
