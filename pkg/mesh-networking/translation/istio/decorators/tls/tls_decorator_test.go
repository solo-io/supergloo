package tls_test

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators/tls"
	"istio.io/api/networking/v1alpha3"
)

var _ = Describe("TlsDecorator", func() {
	var (
		tlsDecorator decorators.TrafficPolicyDestinationRuleDecorator
		output       *v1alpha3.DestinationRule
		ctrl         *gomock.Controller
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		tlsDecorator = tls.NewTlsDecorator()
		output = &v1alpha3.DestinationRule{
			TrafficPolicy: &v1alpha3.TrafficPolicy{},
		}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should set mTLS settings if specified", func() {
		registerField := func(fieldPtr, val interface{}) error {
			return nil
		}
		appliedPolicy := &discoveryv1alpha2.TrafficTargetStatus_AppliedTrafficPolicy{
			Spec: &v1alpha2.TrafficPolicySpec{
				Mtls: &v1alpha2.TrafficPolicySpec_MTLS{
					Istio: &v1alpha2.TrafficPolicySpec_MTLS_Istio{
						TlsMode: v1alpha2.TrafficPolicySpec_MTLS_Istio_DISABLE,
					},
				},
			},
		}
		expectedClientTlsSettings := &v1alpha3.ClientTLSSettings{
			Mode: v1alpha3.ClientTLSSettings_DISABLE,
		}
		err := tlsDecorator.ApplyTrafficPolicyToDestinationRule(
			appliedPolicy,
			nil,
			output,
			registerField,
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(output.TrafficPolicy.Tls).To(Equal(expectedClientTlsSettings))
	})

	It("should return nil if mTLS settings not specified", func() {
		registerField := func(fieldPtr, val interface{}) error {
			return nil
		}
		appliedPolicy := &discoveryv1alpha2.TrafficTargetStatus_AppliedTrafficPolicy{
			Spec: &v1alpha2.TrafficPolicySpec{},
		}
		err := tlsDecorator.ApplyTrafficPolicyToDestinationRule(
			appliedPolicy,
			nil,
			output,
			registerField,
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(output.TrafficPolicy.Tls).To(BeNil())
	})
})
