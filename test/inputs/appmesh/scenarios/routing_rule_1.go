package scenarios

import (
	appmeshApi "github.com/aws/aws-sdk-go/service/appmesh"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/translator/appmesh"
	appmeshInputs "github.com/solo-io/supergloo/test/inputs/appmesh"

	"fmt"

	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
)

type routingRuleScenario1 struct {
	meshName     string
	allResources appmeshInputs.TestResourceSet
}

func RoutingRule1() AppMeshTestScenario {
	return &routingRuleScenario1{
		meshName:     MeshName,
		allResources: GetAllResources(),
	}
}

func (s *routingRuleScenario1) GetMeshName() string {
	return s.meshName
}

func (s *routingRuleScenario1) GetResources() appmeshInputs.TestResourceSet {
	return s.allResources
}

func (s *routingRuleScenario1) GetRoutingRules() v1.RoutingRuleList {
	return v1.RoutingRuleList{s.getRoutingRule()}
}

func (s *routingRuleScenario1) VerifyExpectations(configuration appmesh.AwsAppMeshConfiguration) {
	routingRuleScenario1Expectations(configuration)
}

func routingRuleScenario1Expectations(configuration appmesh.AwsAppMeshConfiguration) {
	config, ok := configuration.(*appmesh.AwsAppMeshConfigurationImpl)
	ExpectWithOffset(1, ok).To(BeTrue())

	// Verify virtual nodes
	ExpectWithOffset(1, config.VirtualNodes).To(HaveLen(6))
	for hostname, expectedVn := range routingRule1VirtualNodes() {

		vn, ok := config.VirtualNodes[hostname]
		ExpectWithOffset(1, ok).To(BeTrue())
		ExpectWithOffset(1, vn.MeshName).To(BeEquivalentTo(expectedVn.MeshName))
		ExpectWithOffset(1, vn.VirtualNodeName).To(BeEquivalentTo(expectedVn.VirtualNodeName))
		ExpectWithOffset(1, vn.Spec.Listeners).To(ConsistOf(expectedVn.Spec.Listeners))
		ExpectWithOffset(1, vn.Spec.ServiceDiscovery).To(BeEquivalentTo(expectedVn.Spec.ServiceDiscovery))
		ExpectWithOffset(1, vn.Spec.Backends).To(ConsistOf(expectedVn.Spec.Backends))
	}

	// Verify virtual services
	ExpectWithOffset(1, config.VirtualServices).To(HaveLen(1))
	for hostname, expectedVs := range routingRule1VirtualServices() {
		vs, ok := config.VirtualServices[hostname]
		ExpectWithOffset(1, ok).To(BeTrue())
		ExpectWithOffset(1, vs.VirtualServiceName).To(BeEquivalentTo(expectedVs.VirtualServiceName))
		ExpectWithOffset(1, vs.MeshName).To(BeEquivalentTo(expectedVs.MeshName))
		ExpectWithOffset(1, vs.Spec.Provider.VirtualNode).To(BeEquivalentTo(expectedVs.Spec.Provider.VirtualNode))
		ExpectWithOffset(1, vs.Spec.Provider.VirtualRouter).To(BeEquivalentTo(expectedVs.Spec.Provider.VirtualRouter))
	}

	// Verify virtual routers
	ExpectWithOffset(1, config.VirtualRouters).To(HaveLen(1))
	ExpectWithOffset(1, config.VirtualRouters).To(ConsistOf(routingRule1VirtualRouters()))

	// Verify routes
	routes := config.Routes
	expectedRoutes := routingRule1Routes()
	ExpectWithOffset(1, routes).To(BeEquivalentTo(expectedRoutes))
}

func routingRule1VirtualNodes() map[string]*appmeshApi.VirtualNodeData {
	return map[string]*appmeshApi.VirtualNodeData{
		productPageHostname: createVirtualNode(productPageVnName, productPageHostname, MeshName, "http", 9080, []string{reviewsV1Hostname}),
		detailsHostname:     createVirtualNode(detailsVnName, detailsHostname, MeshName, "http", 9080, []string{reviewsV1Hostname}),
		reviewsV1Hostname:   createVirtualNode(reviewsV1VnName, reviewsV1Hostname, MeshName, "http", 9080, nil),
		reviewsV2Hostname:   createVirtualNode(reviewsV2VnName, reviewsV2Hostname, MeshName, "http", 9080, []string{reviewsV1Hostname}),
		reviewsV3Hostname:   createVirtualNode(reviewsV3VnName, reviewsV3Hostname, MeshName, "http", 9080, []string{reviewsV1Hostname}),
		ratingsHostname:     createVirtualNode(ratingsVnName, ratingsHostname, MeshName, "http", 9080, []string{reviewsV1Hostname}),
	}
}

func routingRule1VirtualServices() map[string]*appmeshApi.VirtualServiceData {
	return map[string]*appmeshApi.VirtualServiceData{
		reviewsV1Hostname: createVirtualServiceWithVr(reviewsV1Hostname, MeshName, reviewsV1Hostname),
	}
}

func routingRule1VirtualRouters() []*appmeshApi.VirtualRouterData {
	return []*appmeshApi.VirtualRouterData{createVirtualRouter(MeshName, reviewsV1Hostname, 9080)}
}

func routingRule1Routes() []*appmeshApi.RouteData {
	action := createRouteAction([]vnWeightTuple{
		{
			virtualNode: reviewsV1VnName,
			weight:      80,
		},
		{
			virtualNode: reviewsV2VnName,
			weight:      10,
		},
		{
			virtualNode: reviewsV3VnName,
			weight:      10,
		},
	})

	return []*appmeshApi.RouteData{
		createRoute(MeshName, fmt.Sprintf("%s-%d", reviewsV1Hostname, 0), reviewsV1Hostname, "/test", action),
		createRoute(MeshName, fmt.Sprintf("%s-%d", reviewsV1Hostname, 1), reviewsV1Hostname, "/", action),
	}
}

func (s *routingRuleScenario1) getRoutingRule() *v1.RoutingRule {
	return &v1.RoutingRule{
		Metadata: core.Metadata{
			Name:      "reviews",
			Namespace: "supergloo-system",
		},
		Spec: &v1.RoutingRuleSpec{
			RuleType: &v1.RoutingRuleSpec_TrafficShifting{
				TrafficShifting: &v1.TrafficShifting{
					Destinations: &gloov1.MultiDestination{
						Destinations: []*gloov1.WeightedDestination{
							{
								Weight: 80,
								Destination: &gloov1.Destination{
									Upstream: core.ResourceRef{
										Namespace: "gloo-system",
										Name:      "default-reviews-9080",
									},
								},
							},
							{
								Weight: 10,
								Destination: &gloov1.Destination{
									Upstream: core.ResourceRef{
										Namespace: "gloo-system",
										Name:      "default-reviews-v2-9080",
									},
								},
							},
							{
								Weight: 10,
								Destination: &gloov1.Destination{
									Upstream: core.ResourceRef{
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
		RequestMatchers: []*gloov1.Matcher{

			{
				PathSpecifier: &gloov1.Matcher_Prefix{
					Prefix: "/test",
				},
			},
			{
				PathSpecifier: &gloov1.Matcher_Prefix{
					Prefix: "/",
				},
			},
		},
		DestinationSelector: &v1.PodSelector{
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
		},
	}
}
