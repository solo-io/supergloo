package appmesh

import (
	"github.com/aws/aws-sdk-go/service/appmesh"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"
	"github.com/solo-io/supergloo/pkg/api/custom/clients/kubernetes"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/webhook/test"
	"github.com/solo-io/supergloo/test/inputs"
)

var _ = Describe("config translator", func() {
	var (
		injectedPodList v1.PodList
		upstreamList    gloov1.UpstreamList
	)

	var defaultConfig = func() *awsAppMeshConfiguration {
		mesh := inputs.AppmeshMesh("")
		config, err := NewAwsAppMeshConfiguration(mesh, injectedPodList, upstreamList)
		Expect(err).NotTo(HaveOccurred())
		err = config.AllowAll()
		Expect(err).NotTo(HaveOccurred())
		typedConfig, ok := config.(*awsAppMeshConfiguration)
		Expect(ok).To(BeTrue())
		err = typedConfig.initialize()
		Expect(err).NotTo(HaveOccurred())
		return typedConfig
	}

	var defaultRoutingRule = func(destinations *gloov1.MultiDestination) *v1.RoutingRule {
		routingRule := &v1.RoutingRule{
			Metadata: core.Metadata{
				Name:      "one",
				Namespace: "supergloo-system",
			},
			Spec: &v1.RoutingRuleSpec{
				RuleType: &v1.RoutingRuleSpec_TrafficShifting{
					TrafficShifting: &v1.TrafficShifting{
						Destinations: destinations,
					},
				},
			},
			RequestMatchers: []*gloov1.Matcher{
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
								Namespace: "supergloo-system",
								Name:      "namespace-with-inject-reviews-9080",
							},
						},
					},
				},
			},
		}
		return routingRule
	}
	BeforeEach(func() {
		clients.UseMemoryClients()
		injectedPodList = test.MustGetInjectedPodList()
		upstreamList = test.MustGetUpstreamList()
	})

	Context("get pod info", func() {
		It("can get valid virtual node name", func() {
			kubePod := injectedPodList[0]
			pod, err := kubernetes.ToKubePod(kubePod)
			Expect(err).NotTo(HaveOccurred())
			info, err := getPodInfo(inputs.AppmeshMesh(""), pod)
			Expect(err).NotTo(HaveOccurred())
			Expect(info.virtualNodeName).To(Equal("productpage-v1"))
			Expect(info.ports).To(Equal([]uint32{9080}))
		})
		It("will return nil, if meshname doesn't match virtual node path", func() {
			kubePod := injectedPodList[0]
			pod, err := kubernetes.ToKubePod(kubePod)
			Expect(err).NotTo(HaveOccurred())
			info, err := getPodInfo(inputs.AppmeshMesh("123"), pod)
			Expect(err).NotTo(HaveOccurred())
			Expect(info).To(BeNil())
		})
	})
	Context("get pods for mesh", func() {
		It("can filter valid mesh pods", func() {
			_, podList, err := getPodsForMesh(inputs.AppmeshMesh(""), injectedPodList)
			Expect(err).NotTo(HaveOccurred())
			Expect(podList).To(HaveLen(6))
		})
	})

	Context("get upstreams for mesh", func() {
		It("can get valid upstreams for the mesh", func() {
			info, podList, err := getPodsForMesh(inputs.AppmeshMesh(""), injectedPodList)
			Expect(err).NotTo(HaveOccurred())
			_, usList, err := getUpstreamsForMesh(upstreamList, info, podList)
			Expect(err).NotTo(HaveOccurred())
			Expect(usList).To(HaveLen(9))
		})
	})
	Context("allow all", func() {
		It("can create the proper config for allow all (no routing rules)", func() {
			mesh := inputs.AppmeshMesh("")
			config, err := NewAwsAppMeshConfiguration(mesh, injectedPodList, upstreamList)
			Expect(err).NotTo(HaveOccurred())
			err = config.ProcessRoutingRules(nil)
			Expect(err).NotTo(HaveOccurred())
			err = config.AllowAll()
			Expect(err).NotTo(HaveOccurred())
			typedConfig, ok := config.(*awsAppMeshConfiguration)
			Expect(ok).To(BeTrue())
			Expect(typedConfig.MeshName).To(Equal(mesh.Metadata.Name))
		})
		It("can create the proper config for allow all (with routing rules)", func() {
			destinations := &gloov1.MultiDestination{
				Destinations: []*gloov1.WeightedDestination{
					{
						Weight: 50,
						Destination: &gloov1.Destination{
							Upstream: core.ResourceRef{
								Namespace: "supergloo-system",
								Name:      "namespace-with-inject-reviews-v3-9080",
							},
						},
					},
					{
						Weight: 50,
						Destination: &gloov1.Destination{
							Upstream: core.ResourceRef{
								Namespace: "supergloo-system",
								Name:      "namespace-with-inject-reviews-v2-9080",
							},
						},
					},
				},
			}
			rules := v1.RoutingRuleList{defaultRoutingRule(destinations)}
			mesh := inputs.AppmeshMesh("")
			config, err := NewAwsAppMeshConfiguration(mesh, injectedPodList, upstreamList)
			Expect(err).NotTo(HaveOccurred())
			err = config.ProcessRoutingRules(rules)
			Expect(err).NotTo(HaveOccurred())
			err = config.AllowAll()
			Expect(err).NotTo(HaveOccurred())
			typedConfig, ok := config.(*awsAppMeshConfiguration)
			Expect(ok).To(BeTrue())
			Expect(typedConfig.MeshName).To(Equal(mesh.Metadata.Name))
		})
	})

	Context("Routing Rules", func() {
		Context("matchers", func() {
			It("cannot have 0 matchers", func() {
				routingRules := v1.RoutingRuleList{
					{
						Spec: &v1.RoutingRuleSpec{
							RuleType: &v1.RoutingRuleSpec_FaultInjection{},
						},
					},
				}
				typedConfig := defaultConfig()
				err := typedConfig.ProcessRoutingRules(routingRules)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("appmesh requires exactly one matcher, 0 found"))
			})

			It("Cannot have > 2 matchers", func() {
				routingRules := v1.RoutingRuleList{
					{
						Spec: &v1.RoutingRuleSpec{
							RuleType: &v1.RoutingRuleSpec_FaultInjection{},
						},
						RequestMatchers: []*gloov1.Matcher{
							{}, {},
						},
					},
				}
				typedConfig := defaultConfig()
				err := typedConfig.ProcessRoutingRules(routingRules)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("appmesh requires exactly one matcher, 2 found"))
			})
			It("must be a prefix matcher", func() {
				routingRules := v1.RoutingRuleList{
					{
						Spec: &v1.RoutingRuleSpec{
							RuleType: &v1.RoutingRuleSpec_FaultInjection{},
						},
						RequestMatchers: []*gloov1.Matcher{
							{
								PathSpecifier: &gloov1.Matcher_Exact{},
							},
						},
					},
				}
				typedConfig := defaultConfig()
				err := typedConfig.ProcessRoutingRules(routingRules)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("unsupported matcher type found"))
			})
		})

		Context("traffic shifting", func() {
			// This test actually checks the hostname, so 2 different upstreams with the same host will also fail
			It("Every upstream must have a unique host", func() {
				route := &appmesh.HttpRoute{}
				destinations := &gloov1.MultiDestination{
					Destinations: []*gloov1.WeightedDestination{
						{
							Weight: 50,
							Destination: &gloov1.Destination{
								Upstream: core.ResourceRef{
									Namespace: "supergloo-system",
									Name:      "namespace-with-inject-reviews-v3-9080",
								},
							},
						},
						{
							Weight: 50,
							Destination: &gloov1.Destination{
								Upstream: core.ResourceRef{
									Namespace: "supergloo-system",
									Name:      "namespace-with-inject-reviews-v3-9080",
								},
							},
						},
					},
				}
				routingRule := defaultRoutingRule(destinations)
				typedConfig := defaultConfig()
				err := processTrafficShiftingRule(typedConfig.upstreamList, typedConfig.VirtualNodes,
					routingRule.GetSpec().GetTrafficShifting(), route)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("all appmesh destinations must be unique services"))
			})
			It("only works with valid destinations", func() {
				route := &appmesh.HttpRoute{}
				destinations := &gloov1.MultiDestination{
					Destinations: []*gloov1.WeightedDestination{
						{
							Weight: 100,
							Destination: &gloov1.Destination{
								Upstream: core.ResourceRef{
									Namespace: "supergloo-system-1",
									Name:      "namespace-with-inject-reviews-v3-9080",
								},
							},
						},
					},
				}
				routingRule := defaultRoutingRule(destinations)
				typedConfig := defaultConfig()
				err := processTrafficShiftingRule(typedConfig.upstreamList, typedConfig.VirtualNodes,
					routingRule.GetSpec().GetTrafficShifting(), route)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not find upstream for destination:"))
			})
			It("can do the happy path", func() {

				route := &appmesh.HttpRoute{}
				destinations := &gloov1.MultiDestination{
					Destinations: []*gloov1.WeightedDestination{
						{
							Weight: 50,
							Destination: &gloov1.Destination{
								Upstream: core.ResourceRef{
									Namespace: "supergloo-system",
									Name:      "namespace-with-inject-reviews-v3-9080",
								},
							},
						},
						{
							Weight: 50,
							Destination: &gloov1.Destination{
								Upstream: core.ResourceRef{
									Namespace: "supergloo-system",
									Name:      "namespace-with-inject-reviews-v2-9080",
								},
							},
						},
					},
				}
				routingRule := defaultRoutingRule(destinations)
				typedConfig := defaultConfig()
				err := processTrafficShiftingRule(typedConfig.upstreamList, typedConfig.VirtualNodes,
					routingRule.GetSpec().GetTrafficShifting(), route)
				Expect(err).NotTo(HaveOccurred())
			})

		})

		Context("General errors", func() {
			It("can only handle traffic shifting", func() {
				routingRules := v1.RoutingRuleList{{
					Spec: &v1.RoutingRuleSpec{
						RuleType: &v1.RoutingRuleSpec_FaultInjection{},
					},
					RequestMatchers: []*gloov1.Matcher{
						{
							PathSpecifier: &gloov1.Matcher_Prefix{
								Prefix: "/",
							},
						},
					},
				}}
				typedConfig := defaultConfig()
				err := typedConfig.ProcessRoutingRules(routingRules)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("currently only traffic shifting rules are supported by appmesh"))
			})
			It("fails when no destinations are provided", func() {
				routingRules := v1.RoutingRuleList{
					defaultRoutingRule(nil),
				}
				typedConfig := defaultConfig()
				err := typedConfig.ProcessRoutingRules(routingRules)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("traffic shifting destinations cannot be missing or empty"))
			})
		})

	})
})
