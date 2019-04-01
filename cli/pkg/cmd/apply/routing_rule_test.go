package apply_test

import (
	"fmt"
	"strings"
	"time"

	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	skclients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/cli/test/utils"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

var _ = Describe("RoutingRule", func() {
	type destination struct {
		core.ResourceRef
		weight uint32
	}
	trafficShiftingArgs := func(name string, dests []destination) string {
		args := fmt.Sprintf("apply routingrule trafficshifting --name=%v --target-mesh=my.mesh ", name)
		for _, dest := range dests {
			args += fmt.Sprintf("--destination=%v.%v:%v ", dest.Namespace, dest.Name, dest.weight)
		}
		return args
	}
	faultInjectionArgs := func(name, mainType, subType, percentage string, extraArgs ...string) string {
		args := fmt.Sprintf("apply routingrule fi %v %v --name=%v"+
			" --target-mesh=my.mesh -p %v %v", mainType, subType, name, percentage, strings.Join(extraArgs, " "))
		return args
	}

	BeforeEach(func() {
		clients.UseMemoryClients()
		_, _ = clients.MustMeshClient().Write(&v1.Mesh{Metadata: core.Metadata{Namespace: "my", Name: "mesh"}}, skclients.WriteOpts{})
	})

	getRoutingRule := func(name string) *v1.RoutingRule {
		rr, err := clients.MustRoutingRuleClient().Read("supergloo-system", name, skclients.ReadOpts{})
		ExpectWithOffset(1, err).NotTo(HaveOccurred())
		return rr
	}

	Context("zero destinations provided", func() {
		It("errors", func() {
			name := "ts-rr"

			args := trafficShiftingArgs(name, nil)

			err := utils.Supergloo(args)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("must provide at least 1 destination"))
		})
	})
	Context("no target mesh", func() {
		It("returns an error", func() {
			err := utils.Supergloo("apply routingrule trafficshifting --name foo --destination=my.upstream:5")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("target mesh must be specified, provide with --target-mesh flag"))
		})
	})
	Context("nonexistant target mesh", func() {
		It("returns an error", func() {
			err := utils.Supergloo("apply routingrule trafficshifting --name foo --destination=my.upstream:5 --target-mesh notmy.mesh")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("notmy.mesh does not exist"))
		})
	})
	Context("fault injection", func() {
		It("fails with an invalid percentage", func() {
			name := "fi-rr"
			args := faultInjectionArgs(name, "a", "http", "101", "--status", "404")
			err := utils.Supergloo(args)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid value 101: percentage must be (0-100)"))
		})
		Context("abort", func() {
			It("can create an http abort fault rule", func() {
				name := "fi-rr"
				args := faultInjectionArgs(name, "a", "http", "50", "--status", "404")
				err := utils.Supergloo(args)
				Expect(err).NotTo(HaveOccurred())
				rr := getRoutingRule(name)
				faultType, ok := rr.Spec.RuleType.(*v1.RoutingRuleSpec_FaultInjection)
				Expect(ok).To(BeTrue())
				Expect(faultType.FaultInjection.Percentage).To(Equal(float64(50)))
				abortType, ok := faultType.FaultInjection.FaultInjectionType.(*v1.FaultInjection_Abort_)
				Expect(ok).To(BeTrue())
				abortHttpType, ok := abortType.Abort.ErrorType.(*v1.FaultInjection_Abort_HttpStatus)
				Expect(ok).To(BeTrue())
				Expect(abortHttpType.HttpStatus).To(Equal(int32(404)))
			})
			It("fails to create an abort rule with an invalid status code", func() {
				name := "fi-rr"
				args := faultInjectionArgs(name, "a", "http", "50", "--status", "600")
				err := utils.Supergloo(args)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid value 600: must be valid http status code"))
			})
		})
		Context("delay", func() {
			It("can create an fixed delay fault rule", func() {
				name := "fi-rr"
				args := faultInjectionArgs(name, "d", "fixed", "50", "-d", "100ns")
				err := utils.Supergloo(args)
				Expect(err).NotTo(HaveOccurred())
				rr := getRoutingRule(name)
				faultType, ok := rr.Spec.RuleType.(*v1.RoutingRuleSpec_FaultInjection)
				Expect(ok).To(BeTrue())
				Expect(faultType.FaultInjection.Percentage).To(Equal(float64(50)))
				delayType, ok := faultType.FaultInjection.FaultInjectionType.(*v1.FaultInjection_Delay_)
				Expect(ok).To(BeTrue())
				Expect(delayType.Delay.DelayType).To(Equal(v1.FaultInjection_Delay_fixed))

				Expect(delayType.Delay.Duration).To(Equal(time.Nanosecond * 100))
			})
			It("fails to create an abort rule with an invalid status code", func() {
				name := "fi-rr"
				args := faultInjectionArgs(name, "d", "fixed", "50", "-d", "10xyz")
				err := utils.Supergloo(args)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("unknown unit xyz in duration 10xyz"))
			})
		})
	})
	Context("selector tests", func() {
		selectorTest := func(extraArgs ...string) (*v1.RoutingRule, error) {
			dests := []destination{{core.ResourceRef{"a", "a"}, 5}}
			name := "ts-rr"

			args := trafficShiftingArgs(name, dests) + strings.Join(extraArgs, " ")

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

			args := trafficShiftingArgs(name, dests)

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
	Context("overwrite previous rule", func() {
		It("updates an existing rule with the same name", func() {

			t := func(dests []destination) {
				name := "ts-rr"

				args := trafficShiftingArgs(name, dests)

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
			}
			t([]destination{
				{core.ResourceRef{"do", "a"}, 5},
			})
			t([]destination{
				{core.ResourceRef{"do", "a"}, 5},
				{core.ResourceRef{"barrel", "roll"}, 55},
			})
		})
	})
	Context("--crd flag", func() {
		It("prints the kubernetes yaml", func() {
			dests := []destination{
				{core.ResourceRef{"do", "a"}, 5},
				{core.ResourceRef{"barrel", "roll"}, 55},
			}
			name := "ts-rr"

			args := trafficShiftingArgs(name, dests)
			args += " --dryrun"

			out, err := utils.SuperglooOut(args)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring(`apiVersion: supergloo.solo.io/v1
kind: RoutingRule
metadata:
  creationTimestamp: null
  name: ts-rr
  namespace: supergloo-system
spec:
  spec:
    trafficShifting:
      destinations:
        destinations:
        - destination:
            upstream:
              name: do
              namespace: a
          weight: 5
        - destination:
            upstream:
              name: barrel
              namespace: roll
          weight: 55
  targetMesh:
    name: mesh
    namespace: my
status: {}
`))
		})
	})
	Context("request matcher tests", func() {
		rmTest := func(extraArgs ...string) (*v1.RoutingRule, error) {
			dests := []destination{{core.ResourceRef{"a", "a"}, 5}}
			name := "ts-rr"

			args := trafficShiftingArgs(name, dests) + strings.Join(extraArgs, " ")

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
