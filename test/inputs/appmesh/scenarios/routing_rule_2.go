package scenarios

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/translator/appmesh"
	appmeshInputs "github.com/solo-io/supergloo/test/inputs/appmesh"

	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
)

type routingRuleScenario2 struct {
	meshName     string
	allResources appmeshInputs.TestResourceSet
}

// In this scenario, we have two routing rules that:
// - have the same traffic shifting configuration (same targets, same weights)
// - have same source and destination selectors
// - match different path prefixes
// the result should be the same a applying a single rule with two matchers
func RoutingRule2() AppMeshTestScenario {
	return &routingRuleScenario2{
		meshName:     MeshName,
		allResources: GetAllResources(),
	}
}

func (s *routingRuleScenario2) GetMeshName() string {
	return s.meshName
}

func (s *routingRuleScenario2) GetResources() appmeshInputs.TestResourceSet {
	return s.allResources
}

func (s *routingRuleScenario2) GetRoutingRules() v1.RoutingRuleList {
	return s.getRoutingRules()
}

func (s *routingRuleScenario2) VerifyExpectations(configuration appmesh.AwsAppMeshConfiguration) {
	// This is expected to produce the same result as the routing_rule_1 scenario
	routingRuleScenario1Expectations(configuration)
}

func (s *routingRuleScenario2) getRoutingRules() v1.RoutingRuleList {
	spec := &v1.RoutingRuleSpec{
		RuleType: &v1.RoutingRuleSpec_TrafficShifting{
			TrafficShifting: &v1.TrafficShifting{
				Destinations: &gloov1.MultiDestination{
					Destinations: []*gloov1.WeightedDestination{
						{
							Weight: 80,
							Destination: &gloov1.Destination{
								DestinationType: &gloov1.Destination_Upstream{
									Upstream: &core.ResourceRef{
										Namespace: "gloo-system",
										Name:      "default-reviews-9080",
									},
								},
							},
						},
						{
							Weight: 10,
							Destination: &gloov1.Destination{
								DestinationType: &gloov1.Destination_Upstream{
									Upstream: &core.ResourceRef{
										Namespace: "gloo-system",
										Name:      "default-reviews-v2-9080",
									},
								},
							},
						},
						{
							Weight: 10,
							Destination: &gloov1.Destination{
								DestinationType: &gloov1.Destination_Upstream{
									Upstream: &core.ResourceRef{
										Namespace: "gloo-system",
										Name:      "default-reviews-v3-9080",
									},
								},
							},
						},
					},
				},
			},
		},
	}
	dest := &v1.PodSelector{
		SelectorType: &v1.PodSelector_UpstreamSelector_{
			UpstreamSelector: &v1.PodSelector_UpstreamSelector{
				Upstreams: []core.ResourceRef{
					{
						Namespace: "gloo-system",
						Name:      "default-reviews-9080",
					},
				},
			},
		},
	}
	return v1.RoutingRuleList{
		{
			Metadata: core.Metadata{
				Name:      "reviews-test-path",
				Namespace: "supergloo-system",
			},
			Spec: spec,
			RequestMatchers: []*gloov1.Matcher{

				{
					PathSpecifier: &gloov1.Matcher_Prefix{
						Prefix: "/test",
					},
				},
			},
			DestinationSelector: dest,
		},
		{
			Metadata: core.Metadata{
				Name:      "reviews-base-path",
				Namespace: "supergloo-system",
			},
			Spec: spec,
			RequestMatchers: []*gloov1.Matcher{
				{
					PathSpecifier: &gloov1.Matcher_Prefix{
						Prefix: "/",
					},
				},
			},
			DestinationSelector: dest,
		},
	}
}
