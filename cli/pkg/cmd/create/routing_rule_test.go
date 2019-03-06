package create_test

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/cli/pkg/helpers"
	"github.com/solo-io/supergloo/cli/test/utils"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

var _ = Describe("RoutingRule", func() {
	type destination struct {
		core.ResourceRef
		weight uint32
	}
	rrArgs := func(name string, dests []destination) string {
		args := fmt.Sprintf("create routingrule trafficshifting --name=%v ", name)
		for _, dest := range dests {
			args += fmt.Sprintf("--destination=%v.%v:%v ", dest.Namespace, dest.Name, dest.weight)
		}
		return args
	}

	BeforeEach(func() {
		helpers.UseMemoryClients()
	})

	getRoutingRule := func(name string) *v1.RoutingRule {
		rr, err := helpers.MustRoutingRuleClient().Read("supergloo-system", name, clients.ReadOpts{})
		ExpectWithOffset(1, err).NotTo(HaveOccurred())
		return rr
	}

	Context("zero destinations provided", func() {
		It("errors", func() {
			name := "ts-rr"

			args := rrArgs(name, nil)

			err := utils.Supergloo(args)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("must provide at least 1 destination"))
		})
	})
	Context("selector tests", func() {
		selectorTest := func(extraArgs ...string) (*v1.RoutingRule, error) {
			dests := []destination{{core.ResourceRef{"a", "a"}, 5}}
			name := "ts-rr"

			args := rrArgs(name, dests) + strings.Join(extraArgs, " ")

			err := utils.Supergloo(args)
			if err != nil {
				return nil, err
			}

			routingRule := getRoutingRule(name)

			return routingRule, nil
		}
		Context("no selector", func() {
			It("creates a rule with no selectors", func() {
				routingRule, err := selectorTest()
				Expect(err).NotTo(HaveOccurred())

				Expect(routingRule.SourceSelector).To(BeNil())
				Expect(routingRule.DestinationSelector).To(BeNil())
			})
		})
		Context("labels selector", func() {
			It("creates a rule with label selectors", func() {
				routingRule, err := selectorTest("--source-labels KEY1=VAL1",
					"--source-labels KEY2=VAL2",
					"--dest-labels KEY1=VAL1",
					"--dest-labels KEY2=VAL2")
				Expect(err).NotTo(HaveOccurred())

				expectedMap := map[string]string{"KEY1": "VAL1", "KEY2": "VAL2"}

				Expect(routingRule.SourceSelector).NotTo(BeNil())
				Expect(routingRule.SourceSelector.SelectorType).To(BeAssignableToTypeOf(&v1.PodSelector_LabelSelector_{}))
				ss := routingRule.SourceSelector.SelectorType.(*v1.PodSelector_LabelSelector_).LabelSelector
				Expect(ss.LabelsToMatch).To(Equal(expectedMap))

				Expect(routingRule.DestinationSelector).NotTo(BeNil())
				Expect(routingRule.DestinationSelector.SelectorType).To(BeAssignableToTypeOf(&v1.PodSelector_LabelSelector_{}))
				ds := routingRule.DestinationSelector.SelectorType.(*v1.PodSelector_LabelSelector_).LabelSelector
				Expect(ds.LabelsToMatch).To(Equal(expectedMap))
			})
		})
		Context("upstream selector", func() {
			It("creates a rule with upstream selectors", func() {
				routingRule, err := selectorTest("--source-upstreams ns1.us1",
					"--source-upstreams ns2.us2",
					"--dest-upstreams ns3.us3",
					"--dest-upstreams ns4.us4")
				Expect(err).NotTo(HaveOccurred())

				Expect(routingRule.SourceSelector).NotTo(BeNil())
				Expect(routingRule.SourceSelector.SelectorType).To(BeAssignableToTypeOf(&v1.PodSelector_UpstreamSelector_{}))
				ss := routingRule.SourceSelector.SelectorType.(*v1.PodSelector_UpstreamSelector_).UpstreamSelector
				Expect(ss.Upstreams).To(Equal([]core.ResourceRef{{"us1", "ns1"}, {"us2", "ns2"}}))

				Expect(routingRule.DestinationSelector).NotTo(BeNil())
				Expect(routingRule.DestinationSelector.SelectorType).To(BeAssignableToTypeOf(&v1.PodSelector_UpstreamSelector_{}))
				ds := routingRule.DestinationSelector.SelectorType.(*v1.PodSelector_UpstreamSelector_).UpstreamSelector
				Expect(ds.Upstreams).To(Equal([]core.ResourceRef{{"us3", "ns3"}, {"us4", "ns4"}}))
			})
		})
		Context("namespace selector", func() {
			It("creates a rule with namespace selectors", func() {
				routingRule, err := selectorTest("--source-namespaces ns1",
					"--source-namespaces ns2",
					"--dest-namespaces ns3",
					"--dest-namespaces ns4")
				Expect(err).NotTo(HaveOccurred())

				Expect(routingRule.SourceSelector).NotTo(BeNil())
				Expect(routingRule.SourceSelector.SelectorType).To(BeAssignableToTypeOf(&v1.PodSelector_NamespaceSelector_{}))
				ss := routingRule.SourceSelector.SelectorType.(*v1.PodSelector_NamespaceSelector_).NamespaceSelector
				Expect(ss.Namespaces).To(Equal([]string{"ns1", "ns2"}))

				Expect(routingRule.DestinationSelector).NotTo(BeNil())
				Expect(routingRule.DestinationSelector.SelectorType).To(BeAssignableToTypeOf(&v1.PodSelector_NamespaceSelector_{}))
				ds := routingRule.DestinationSelector.SelectorType.(*v1.PodSelector_NamespaceSelector_).NamespaceSelector
				Expect(ds.Namespaces).To(Equal([]string{"ns3", "ns4"}))
			})
		})
		Context("conflicting selector types", func() {
			It("returns an error", func() {
				_, err := selectorTest("--source-upstreams ns1.us1",
					"--source-namespaces ns2")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("you may only use one type of selector: upstreams, namespaces, or labels"))
			})
			It("returns an error", func() {
				_, err := selectorTest("--dest-upstreams ns3.us3",
					"--dest-namespaces ns4")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("you may only use one type of selector: upstreams, namespaces, or labels"))
			})
		})
	})
	Context("multiple destinations", func() {
		It("creates a rule with a resource ref for each destination and weight", func() {
			dests := []destination{
				{core.ResourceRef{"do", "a"}, 5},
				{core.ResourceRef{"barrel", "roll"}, 55},
			}
			name := "ts-rr"

			args := rrArgs(name, dests)

			err := utils.Supergloo(args)
			Expect(err).NotTo(HaveOccurred())

			routingRule := getRoutingRule(name)

			Expect(routingRule.Spec.RuleType).To(BeAssignableToTypeOf(&v1.RoutingRuleSpec_TrafficShifting{}))
			ts := routingRule.Spec.RuleType.(*v1.RoutingRuleSpec_TrafficShifting)
			Expect(ts.TrafficShifting.Destinations.Destinations).To(HaveLen(len(dests)))
			for i, dest := range dests {
				Expect(ts.TrafficShifting.Destinations.Destinations[i].Weight).To(Equal(dest.weight))
				Expect(ts.TrafficShifting.Destinations.Destinations[i].Destination.Upstream).To(Equal(core.ResourceRef{
					Namespace: dest.Namespace,
					Name:      dest.Name,
				}))
			}
		})
	})
	Context("request matcher tests", func() {
		rmTest := func(extraArgs ...string) (*v1.RoutingRule, error) {
			dests := []destination{{core.ResourceRef{"a", "a"}, 5}}
			name := "ts-rr"

			args := rrArgs(name, dests) + strings.Join(extraArgs, " ")

			err := utils.Supergloo(args)
			if err != nil {
				return nil, err
			}

			routingRule := getRoutingRule(name)

			return routingRule, nil
		}
		Context("no matcher", func() {
			It("creates a rule with no matchers", func() {
				routingRule, err := rmTest()
				Expect(err).NotTo(HaveOccurred())

				Expect(routingRule.RequestMatchers).To(BeEmpty())
			})
		})
		Context("invalid matcher: no path types defined", func() {
			It("returns error", func() {
				_, err := rmTest(
					`--request-matcher {"methods":["GET","POST"],"header_matchers":{"foo":"bar","baz":"qux"}} `,
				)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("must provide path prefix, path exact, or path regex for route matcher"))
			})
		})
		Context("invalid matcher: multiple path types defined", func() {
			It("returns error", func() {
				_, err := rmTest(
					`--request-matcher {"path_prefix":"/","path_exact":"/foo","methods":["GET","POST"],"header_matchers":{"foo":"bar","baz":"qux"}} `,
				)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("can only set one of path-regex, path-prefix, or path-exact"))
			})
		})
		Context("valid matchers", func() {
			It("creates a rule with a corresponding matcher for each provided matcher", func() {
				routingRule, err := rmTest(
					`--request-matcher {"path_prefix":"/","methods":["GET","POST"],"header_matchers":{"foo":"bar","baz":"qux"}} `,
					`--request-matcher {"path_exact":"/foo","methods":["POST"]} `,
				)
				Expect(err).NotTo(HaveOccurred())

				Expect(routingRule.RequestMatchers).To(HaveLen(2))
				Expect(routingRule.RequestMatchers[0]).To(Equal(&gloov1.Matcher{
					PathSpecifier: &gloov1.Matcher_Prefix{Prefix: "/"},
					Headers: []*gloov1.HeaderMatcher{
						{
							Name:  "baz",
							Value: "qux",
							Regex: true,
						},
						{
							Name:  "foo",
							Value: "bar",
							Regex: true,
						},
					},
					Methods: []string{"GET", "POST"},
				}))
				Expect(routingRule.RequestMatchers[1]).To(Equal(&gloov1.Matcher{
					PathSpecifier: &gloov1.Matcher_Exact{Exact: "/foo"},
					Methods:       []string{"POST"},
				}))
			})
		})
	})
})
