package plugins_test

import (
	"github.com/gogo/protobuf/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/pkg/api/external/istio/networking/v1alpha3"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	. "github.com/solo-io/supergloo/pkg/translator/istio/plugins"
	"github.com/solo-io/supergloo/pkg/translator/utils"
	"github.com/solo-io/supergloo/test/inputs"
)

var _ = Describe("IstioHttp", func() {
	Context("with RoutingRuleSpec_TrafficShifting", func() {
		Context("no upstream", func() {
			It("returns an error", func() {
				params := Params{
					Upstreams: inputs.BookInfoUpstreams("default"),
				}
				in := inputs.TrafficShiftingRuleSpec()
				out := &v1alpha3.HTTPRoute{}

				err := NewIstioHttpPlugin().ProcessRoute(
					params, *in, out)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("destinations cannot be missing or empty"))
			})
		})
		Context("invalid upstream", func() {
			It("returns an error", func() {
				params := Params{
					Upstreams: inputs.BookInfoUpstreams("default"),
				}
				in := inputs.TrafficShiftingRuleSpec(core.ResourceRef{Name: "happy", Namespace: "gilmore"})
				out := &v1alpha3.HTTPRoute{}

				err := NewIstioHttpPlugin().ProcessRoute(
					params, *in, out)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("did not find upstream"))
			})
		})
		Context("weight is 0", func() {
			It("returns an error", func() {
				params := Params{
					Upstreams: inputs.BookInfoUpstreams("default"),
				}
				in := inputs.TrafficShiftingRuleSpec(core.ResourceRef{Name: "default-reviews-v1-9080", Namespace: "default"})
				in.RuleType.(*v1.RoutingRuleSpec_TrafficShifting).TrafficShifting.Destinations.Destinations[0].Weight = 0
				out := &v1alpha3.HTTPRoute{}

				err := NewIstioHttpPlugin().ProcessRoute(
					params, *in, out)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("weight cannot be 0"))
			})
		})
		Context("valid config", func() {
			It("configures traffic shifting on the route", func() {

				params := Params{
					Upstreams: inputs.BookInfoUpstreams("default"),
				}
				in := inputs.TrafficShiftingRuleSpec(
					core.ResourceRef{Name: "default-reviews-v1-9080", Namespace: "default"},
					core.ResourceRef{Name: "default-reviews-v2-9080", Namespace: "default"},
					core.ResourceRef{Name: "default-reviews-v3-9080", Namespace: "default"},
					core.ResourceRef{Name: "default-details-v1-9080", Namespace: "default"},
				)
				out := &v1alpha3.HTTPRoute{}

				err := NewIstioHttpPlugin().ProcessRoute(
					params, *in, out)

				Expect(err).NotTo(HaveOccurred())

				Expect(out.Route).To(Equal([]*v1alpha3.HTTPRouteDestination{
					{
						Destination: &v1alpha3.Destination{
							Host:   "reviews.default.svc.cluster.local",
							Subset: "app-reviews-version-v1",
							Port: &v1alpha3.PortSelector{
								Port: &v1alpha3.PortSelector_Number{
									Number: 9080,
								},
							},
						},
						Weight: 10,
					},
					{
						Destination: &v1alpha3.Destination{
							Host:   "reviews.default.svc.cluster.local",
							Subset: "app-reviews-version-v2",
							Port: &v1alpha3.PortSelector{
								Port: &v1alpha3.PortSelector_Number{
									Number: 9080,
								},
							},
						},
						Weight: 20,
					},
					{
						Destination: &v1alpha3.Destination{
							Host:   "reviews.default.svc.cluster.local",
							Subset: "app-reviews-version-v3",
							Port: &v1alpha3.PortSelector{
								Port: &v1alpha3.PortSelector_Number{
									Number: 9080,
								},
							},
						},
						Weight: 30,
					},
					{
						Destination: &v1alpha3.Destination{
							Host:   "details.default.svc.cluster.local",
							Subset: "app-details-version-v1",
							Port: &v1alpha3.PortSelector{
								Port: &v1alpha3.PortSelector_Number{
									Number: 9080,
								},
							},
						},
						Weight: 40,
					},
				}))

			})
		})
		Context("valid config, rounding error", func() {
			It("configures traffic shifting on the route, balances weights to sum to 100", func() {

				params := Params{
					Upstreams: inputs.BookInfoUpstreams("default"),
				}
				in := inputs.TrafficShiftingRuleSpec(
					core.ResourceRef{Name: "default-reviews-v1-9080", Namespace: "default"},
					core.ResourceRef{Name: "default-reviews-v2-9080", Namespace: "default"},
					core.ResourceRef{Name: "default-reviews-v3-9080", Namespace: "default"},
				)
				out := &v1alpha3.HTTPRoute{}

				err := NewIstioHttpPlugin().ProcessRoute(
					params, *in, out)

				Expect(err).NotTo(HaveOccurred())
				Expect(out.Route).To(Equal([]*v1alpha3.HTTPRouteDestination{
					{
						Destination: &v1alpha3.Destination{
							Host:   "reviews.default.svc.cluster.local",
							Subset: "app-reviews-version-v1",
							Port: &v1alpha3.PortSelector{
								Port: &v1alpha3.PortSelector_Number{
									Number: 9080,
								},
							},
						},
						Weight: 17,
					},
					{
						Destination: &v1alpha3.Destination{
							Host:   "reviews.default.svc.cluster.local",
							Subset: "app-reviews-version-v2",
							Port: &v1alpha3.PortSelector{
								Port: &v1alpha3.PortSelector_Number{
									Number: 9080,
								},
							},
						},
						Weight: 33,
					},
					{
						Destination: &v1alpha3.Destination{
							Host:   "reviews.default.svc.cluster.local",
							Subset: "app-reviews-version-v3",
							Port: &v1alpha3.PortSelector{
								Port: &v1alpha3.PortSelector_Number{
									Number: 9080,
								},
							},
						},
						Weight: 50,
					},
				}))

			})
		})
	})

	Context("with RoutingRuleSpec_FaultInjection", func() {
		Context("validation errors", func() {
			It("errors with no rule present", func() {
				in := inputs.FaultInjectionRuleSpec(&v1.FaultInjection{
					Percentage: 50,
				})
				out := &v1alpha3.HTTPRoute{}
				err := NewIstioHttpPlugin().ProcessRoute(Params{}, *in, out)
				Expect(err).To(HaveOccurred())
			})

			It("errors with no delay type present present", func() {
				in := inputs.FaultInjectionRuleSpec(&v1.FaultInjection{
					Percentage: 50,
					FaultInjectionType: &v1.FaultInjection_Abort_{
						Abort: &v1.FaultInjection_Abort{},
					},
				})
				out := &v1alpha3.HTTPRoute{}
				err := NewIstioHttpPlugin().ProcessRoute(Params{}, *in, out)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("proper transformation", func() {
			var (
				percent float64 = 50
				status  int32   = 404
			)
			It("can transform fault delays", func() {
				duration := &types.Duration{
					Seconds: 1,
					Nanos:   1,
				}
				timeDuration, err := utils.DurationFromProto(duration)
				Expect(err).NotTo(HaveOccurred())
				in := inputs.FaultInjectionRuleSpec(&v1.FaultInjection{
					Percentage: percent,
					FaultInjectionType: &v1.FaultInjection_Delay_{
						Delay: &v1.FaultInjection_Delay{
							Duration:  timeDuration,
							DelayType: v1.FaultInjection_Delay_fixed,
						},
					},
				})
				out := &v1alpha3.HTTPRoute{}
				err = NewIstioHttpPlugin().ProcessRoute(Params{}, *in, out)
				Expect(err).NotTo(HaveOccurred())
				Expect(out.Fault).To(Equal(&v1alpha3.HTTPFaultInjection{
					Delay: &v1alpha3.HTTPFaultInjection_Delay{
						Percent: int32(percent),
						HttpDelayType: &v1alpha3.HTTPFaultInjection_Delay_FixedDelay{
							FixedDelay: duration,
						},
					},
				}))
			})
			It("can transform fault aborts	", func() {

				in := inputs.FaultInjectionRuleSpec(&v1.FaultInjection{
					Percentage: percent,
					FaultInjectionType: &v1.FaultInjection_Abort_{
						Abort: &v1.FaultInjection_Abort{
							ErrorType: &v1.FaultInjection_Abort_HttpStatus{
								HttpStatus: status,
							},
						},
					},
				})
				out := &v1alpha3.HTTPRoute{}
				err := NewIstioHttpPlugin().ProcessRoute(Params{}, *in, out)
				Expect(err).NotTo(HaveOccurred())
				Expect(out.Fault).To(Equal(&v1alpha3.HTTPFaultInjection{
					Abort: &v1alpha3.HTTPFaultInjection_Abort{
						Percent: int32(percent),
						ErrorType: &v1alpha3.HTTPFaultInjection_Abort_HttpStatus{
							HttpStatus: status,
						},
					},
				}))
			})
		})
	})
})
