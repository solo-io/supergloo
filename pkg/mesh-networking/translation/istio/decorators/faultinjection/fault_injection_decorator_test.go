package faultinjection_test

import (
	"github.com/gogo/protobuf/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/testutils"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators/faultinjection"
	"istio.io/api/networking/v1alpha3"
)

var _ = Describe("FaultInjectionDecorator", func() {
	var (
		faulInjectionDecorator decorators.TrafficPolicyVirtualServiceDecorator
		output                 *v1alpha3.HTTPRoute
	)

	BeforeEach(func() {
		faulInjectionDecorator = faultinjection.NewFaultInjectionDecorator()
		output = &v1alpha3.HTTPRoute{}
	})

	It("should set fault injection of type abort", func() {
		registerField := func(fieldPtr, val interface{}) error {
			return nil
		}
		appliedPolicy := &discoveryv1alpha2.TrafficTargetStatus_AppliedTrafficPolicy{
			Spec: &v1alpha2.TrafficPolicySpec{
				FaultInjection: &v1alpha2.TrafficPolicySpec_FaultInjection{
					FaultInjectionType: &v1alpha2.TrafficPolicySpec_FaultInjection_Abort_{
						Abort: &v1alpha2.TrafficPolicySpec_FaultInjection_Abort{
							HttpStatus: 404,
						},
					},
					Percentage: 50,
				},
			},
		}
		expectedFaultInjection := &v1alpha3.HTTPFaultInjection{
			Abort: &v1alpha3.HTTPFaultInjection_Abort{
				ErrorType:  &v1alpha3.HTTPFaultInjection_Abort_HttpStatus{HttpStatus: 404},
				Percentage: &v1alpha3.Percent{Value: 50},
			},
		}
		err := faulInjectionDecorator.ApplyTrafficPolicyToVirtualService(
			appliedPolicy,
			nil,
			output,
			registerField,
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(output.Fault).To(Equal(expectedFaultInjection))
	})

	It("should set fault injection of type fixed delay", func() {
		registerField := func(fieldPtr, val interface{}) error {
			return nil
		}
		appliedPolicy := &discoveryv1alpha2.TrafficTargetStatus_AppliedTrafficPolicy{
			Spec: &v1alpha2.TrafficPolicySpec{
				FaultInjection: &v1alpha2.TrafficPolicySpec_FaultInjection{
					FaultInjectionType: &v1alpha2.TrafficPolicySpec_FaultInjection_FixedDelay{
						FixedDelay: &types.Duration{Seconds: 2},
					},
					Percentage: 50,
				},
			},
		}
		expectedFaultInjection := &v1alpha3.HTTPFaultInjection{
			Delay: &v1alpha3.HTTPFaultInjection_Delay{
				HttpDelayType: &v1alpha3.HTTPFaultInjection_Delay_FixedDelay{FixedDelay: &types.Duration{Seconds: 2}},
				Percentage:    &v1alpha3.Percent{Value: 50},
			},
		}
		err := faulInjectionDecorator.ApplyTrafficPolicyToVirtualService(
			appliedPolicy,
			nil,
			output,
			registerField,
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(output.Fault).To(Equal(expectedFaultInjection))
	})

	It("should set fault injection of type exponential delay", func() {
		registerField := func(fieldPtr, val interface{}) error {
			return nil
		}
		appliedPolicy := &discoveryv1alpha2.TrafficTargetStatus_AppliedTrafficPolicy{
			Spec: &v1alpha2.TrafficPolicySpec{
				FaultInjection: &v1alpha2.TrafficPolicySpec_FaultInjection{
					FaultInjectionType: &v1alpha2.TrafficPolicySpec_FaultInjection_ExponentialDelay{
						ExponentialDelay: &types.Duration{Seconds: 2},
					},
					Percentage: 50,
				},
			},
		}
		expectedFaultInjection := &v1alpha3.HTTPFaultInjection{
			Delay: &v1alpha3.HTTPFaultInjection_Delay{
				HttpDelayType: &v1alpha3.HTTPFaultInjection_Delay_ExponentialDelay{ExponentialDelay: &types.Duration{Seconds: 2}},
				Percentage:    &v1alpha3.Percent{Value: 50},
			},
		}
		err := faulInjectionDecorator.ApplyTrafficPolicyToVirtualService(
			appliedPolicy,
			nil,
			output,
			registerField,
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(output.Fault).To(Equal(expectedFaultInjection))
	})

	It("should not set fault injection if error during field registration", func() {
		testErr := eris.New("registration error")
		registerField := func(fieldPtr, val interface{}) error {
			return testErr
		}
		appliedPolicy := &discoveryv1alpha2.TrafficTargetStatus_AppliedTrafficPolicy{
			Spec: &v1alpha2.TrafficPolicySpec{
				FaultInjection: &v1alpha2.TrafficPolicySpec_FaultInjection{
					FaultInjectionType: &v1alpha2.TrafficPolicySpec_FaultInjection_ExponentialDelay{
						ExponentialDelay: &types.Duration{Seconds: 2},
					},
				},
			},
		}
		err := faulInjectionDecorator.ApplyTrafficPolicyToVirtualService(
			appliedPolicy,
			nil,
			output,
			registerField,
		)
		Expect(err).To(testutils.HaveInErrorChain(testErr))
		Expect(output.Fault).To(BeNil())
	})

	It("should return error if fault injection type not specified", func() {
		registerField := func(fieldPtr, val interface{}) error {
			return nil
		}
		appliedPolicy := &discoveryv1alpha2.TrafficTargetStatus_AppliedTrafficPolicy{
			Spec: &v1alpha2.TrafficPolicySpec{
				FaultInjection: &v1alpha2.TrafficPolicySpec_FaultInjection{
					Percentage: 50,
				},
			},
		}
		err := faulInjectionDecorator.ApplyTrafficPolicyToVirtualService(
			appliedPolicy,
			nil,
			output,
			registerField,
		)
		Expect(err.Error()).To(ContainSubstring("FaultInjection type must be specified"))
		Expect(output.Fault).To(BeNil())
	})
})
