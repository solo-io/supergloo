package inputs

import (
	"time"

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
					RequestTimeout: func() *time.Duration {
						dur := time.Second * 10
						return &dur
					}(),
				},
			},
		},
	}
}

func AdvancedBookInfoRoutingRules(namespace string, targetMesh *core.ResourceRef) v1.RoutingRuleList {
	return v1.RoutingRuleList{
		{
			Metadata: core.Metadata{
				Namespace: namespace,
				Name:      "timeouts-productpage",
			},
			TargetMesh: targetMesh,
			SourceSelector: &v1.PodSelector{
				SelectorType: &v1.PodSelector_LabelSelector_{
					LabelSelector: &v1.PodSelector_LabelSelector{
						LabelsToMatch: map[string]string{
							"app": "productpage",
						},
					},
				},
			},
			Spec: &v1.RoutingRuleSpec{
				RuleType: &v1.RoutingRuleSpec_RequestTimeout{
					RequestTimeout: func() *time.Duration {
						dur := time.Second * 10
						return &dur
					}(),
				},
			},
		},
		{
			Metadata: core.Metadata{
				Namespace: namespace,
				Name:      "trafficshifting-productpage",
			},
			TargetMesh: targetMesh,
			SourceSelector: &v1.PodSelector{
				SelectorType: &v1.PodSelector_LabelSelector_{
					LabelSelector: &v1.PodSelector_LabelSelector{
						LabelsToMatch: map[string]string{
							"app": "productpage",
						},
					},
				},
			},
			DestinationSelector: &v1.PodSelector{
				SelectorType: &v1.PodSelector_LabelSelector_{
					LabelSelector: &v1.PodSelector_LabelSelector{
						LabelsToMatch: map[string]string{
							"app": "reviews",
						},
					},
				},
			},
			Spec: TrafficShiftingRuleSpec(core.ResourceRef{Namespace: namespace, Name: namespace + "-reviews-v1-9080"}),
		},
		{
			// added for smi tests
			Metadata: core.Metadata{
				Namespace: namespace,
				Name:      "trafficshifting-reviews-50-50",
			},
			TargetMesh: targetMesh,
			DestinationSelector: &v1.PodSelector{
				SelectorType: &v1.PodSelector_LabelSelector_{
					LabelSelector: &v1.PodSelector_LabelSelector{
						LabelsToMatch: map[string]string{
							"app": "reviews",
						},
					},
				},
			},
			Spec: TrafficShiftingRuleSpec(
				core.ResourceRef{Namespace: namespace, Name: namespace + "-reviews-v1-9080"},
				core.ResourceRef{Namespace: namespace, Name: namespace + "-reviews-v2-9080"},
			),
		},
		{
			Metadata: core.Metadata{
				Namespace: namespace,
				Name:      "faultinjection-productpage",
			},
			TargetMesh: targetMesh,
			SourceSelector: &v1.PodSelector{
				SelectorType: &v1.PodSelector_LabelSelector_{
					LabelSelector: &v1.PodSelector_LabelSelector{
						LabelsToMatch: map[string]string{
							"app": "productpage",
						},
					},
				},
			},
			DestinationSelector: &v1.PodSelector{
				SelectorType: &v1.PodSelector_LabelSelector_{
					LabelSelector: &v1.PodSelector_LabelSelector{
						LabelsToMatch: map[string]string{
							"app": "details",
						},
					},
				},
			},
			Spec: SampleFaultInjectionRuleSpec(),
		},
		{
			Metadata: core.Metadata{
				Namespace: namespace,
				Name:      "retries-ratings",
			},
			TargetMesh: targetMesh,
			DestinationSelector: &v1.PodSelector{
				SelectorType: &v1.PodSelector_LabelSelector_{
					LabelSelector: &v1.PodSelector_LabelSelector{
						LabelsToMatch: map[string]string{
							"app": "ratings",
						},
					},
				},
			},
			Spec: RetryRuleSpec(),
		},
		{
			Metadata: core.Metadata{
				Namespace: namespace,
				Name:      "retries-reviews",
			},
			TargetMesh: targetMesh,
			RequestMatchers: []*gloov1.Matcher{
				{
					PathSpecifier: &gloov1.Matcher_Prefix{Prefix: "/users/"},
					Methods:       []string{"GET", "POST"},
				},
				{
					PathSpecifier: &gloov1.Matcher_Prefix{Prefix: "/posts/"},
					Methods:       []string{"GET", "POST"},
				},
			},
			DestinationSelector: &v1.PodSelector{
				SelectorType: &v1.PodSelector_LabelSelector_{
					LabelSelector: &v1.PodSelector_LabelSelector{
						LabelsToMatch: map[string]string{
							"app": "reviews",
						},
					},
				},
			},
			Spec: RetryRuleSpec(),
		},
	}
}

func TrafficShiftingRuleSpec(destinations ...core.ResourceRef) *v1.RoutingRuleSpec {
	var dests []*gloov1.WeightedDestination
	for i, d := range destinations {
		d := d
		dests = append(dests, &gloov1.WeightedDestination{
			Destination: &gloov1.Destination{
				DestinationType: &gloov1.Destination_Upstream{
					Upstream: &d,
				},
			},
			Weight: uint32(i + 1),
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

func FaultInjectionRuleSpec(fi *v1.FaultInjection) *v1.RoutingRuleSpec {
	return &v1.RoutingRuleSpec{
		RuleType: &v1.RoutingRuleSpec_FaultInjection{
			FaultInjection: fi,
		},
	}
}

func SampleFaultInjectionRuleSpec() *v1.RoutingRuleSpec {
	return &v1.RoutingRuleSpec{
		RuleType: &v1.RoutingRuleSpec_FaultInjection{
			FaultInjection: &v1.FaultInjection{
				Percentage: 50,
				FaultInjectionType: &v1.FaultInjection_Abort_{
					Abort: &v1.FaultInjection_Abort{
						ErrorType: &v1.FaultInjection_Abort_HttpStatus{
							HttpStatus: 418,
						},
					},
				},
			},
		},
	}
}

func RetryRuleSpec() *v1.RoutingRuleSpec {
	return &v1.RoutingRuleSpec{
		RuleType: &v1.RoutingRuleSpec_Retries{
			Retries: &v1.RetryPolicy{
				RetryBudget: &v1.RetryBudget{
					MinRetriesPerSecond: 5,
				},
			},
		},
	}
}
