package inputs

import (
	"time"

	"github.com/gogo/protobuf/types"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

func BookInfoRoutingRules(namespace string, targetMesh *core.ResourceRef) v1.RoutingRuleList {
	return v1.RoutingRuleList{
		{
			Metadata: core.Metadata{
				Namespace: namespace,
				Name:      "rule-applied-to-reviews",
			},
			TargetMesh: targetMesh,
			SourceSelector: &v1.PodSelector{
				SelectorType: &v1.PodSelector_LabelSelector_{
					LabelSelector: &v1.PodSelector_LabelSelector{
						LabelsToMatch: map[string]string{
							"app": "reviews",
						},
					},
				},
			},
			Spec: &v1.RoutingRuleSpec{
				RuleType: &v1.RoutingRuleSpec_RequestTimeout{
					RequestTimeout: types.DurationProto(time.Second * 10),
				},
			},
		},
	}
}

func TrafficShiftingRuleSpec(ups ...core.ResourceRef) *v1.RoutingRuleSpec {
	var dests []*gloov1.WeightedDestination
	for i, us := range ups {
		dests = append(dests, &gloov1.WeightedDestination{
			Destination: &gloov1.Destination{Upstream: us},
			Weight:      uint32(i + 1),
		})
	}
	return &v1.RoutingRuleSpec{
		RuleType: &v1.RoutingRuleSpec_TrafficShifting{
			TrafficShifting: &v1.TrafficShifting{
				Destinations: &gloov1.MultiDestination{
					Destinations: dests,
				},
			},
		},
	}
}
