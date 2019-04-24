package appmesh_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/test/inputs/appmesh/scenarios"

	. "github.com/solo-io/supergloo/pkg/translator/appmesh"
)

var _ = Describe("RoutingRules", func() {

	var (
		scenario scenarios.AppMeshTestScenario
		config   AwsAppMeshConfiguration
		err      error
	)

	BeforeEach(func() {
		scenario = scenarios.InitializeOnly()
		config, err = NewAwsAppMeshConfiguration(
			scenario.GetMeshName(),
			scenario.GetResources().MustGetPodList(),
			scenario.GetResources().MustGetUpstreams())
		Expect(err).NotTo(HaveOccurred())
		Expect(config).NotTo(BeNil())
	})

	It("fails if the route has no matchers", func() {
		routingRules := v1.RoutingRuleList{
			{
				Metadata: core.Metadata{
					Namespace: "test",
					Name:      "rule",
				},
				Spec: &v1.RoutingRuleSpec{
					RuleType: &v1.RoutingRuleSpec_TrafficShifting{
						TrafficShifting: &v1.TrafficShifting{
							Destinations: nil,
						},
					},
				},
				DestinationSelector: &v1.PodSelector{
					SelectorType: &v1.PodSelector_NamespaceSelector_{
						NamespaceSelector: &v1.PodSelector_NamespaceSelector{
							Namespaces: []string{"test"},
						},
					},
				},
			},
		}
		err := config.ProcessRoutingRules(routingRules)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("routing rule test.rule has zero matchers. At least one matcher is required"))
	})

	It("fails if the matcher is not a prefix matcher", func() {
		routingRules := v1.RoutingRuleList{
			{
				Metadata: core.Metadata{
					Namespace: "test",
					Name:      "rule",
				},
				Spec: &v1.RoutingRuleSpec{
					RuleType: &v1.RoutingRuleSpec_TrafficShifting{
						TrafficShifting: &v1.TrafficShifting{
							Destinations: nil,
						},
					},
				},
				RequestMatchers: []*gloov1.Matcher{
					{
						PathSpecifier: &gloov1.Matcher_Exact{},
					},
				},
				DestinationSelector: &v1.PodSelector{
					SelectorType: &v1.PodSelector_NamespaceSelector_{
						NamespaceSelector: &v1.PodSelector_NamespaceSelector{
							Namespaces: []string{"test"},
						},
					},
				},
			},
		}
		err := config.ProcessRoutingRules(routingRules)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("unsupported matcher type *v1.Matcher_Exact on routing rule test.rule"))
	})

	It("fails if no destinations are provided", func() {
		routingRules := v1.RoutingRuleList{
			{
				Metadata: core.Metadata{
					Namespace: "test",
					Name:      "rule",
				},
				Spec: &v1.RoutingRuleSpec{
					RuleType: &v1.RoutingRuleSpec_TrafficShifting{
						TrafficShifting: &v1.TrafficShifting{
							Destinations: nil,
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
			},
		}
		err := config.ProcessRoutingRules(routingRules)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("traffic shifting destinations cannot be missing or empty"))
	})

	It("fails if the route has is of an unsupported type", func() {
		routingRules := v1.RoutingRuleList{
			{
				Metadata: core.Metadata{
					Namespace: "test",
					Name:      "rule",
				},
				RequestMatchers: []*gloov1.Matcher{
					{
						PathSpecifier: &gloov1.Matcher_Prefix{
							Prefix: "/",
						},
					},
				},
				Spec: &v1.RoutingRuleSpec{
					RuleType: &v1.RoutingRuleSpec_FaultInjection{},
				},
				DestinationSelector: &v1.PodSelector{
					SelectorType: &v1.PodSelector_NamespaceSelector_{
						NamespaceSelector: &v1.PodSelector_NamespaceSelector{
							Namespaces: []string{"test"},
						},
					},
				},
			},
		}
		err := config.ProcessRoutingRules(routingRules)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Currently only traffic shifting rules are supported for AWS AppMesh"))
	})

})
