package inputs

import (
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

func BookInfoRoutingRules(namespace string) v1.RoutingRuleList {
	return v1.RoutingRuleList{
		{
			Metadata: core.Metadata{
				Namespace: namespace,
				Name:      "rule-applied-to-reviews",
			},
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
