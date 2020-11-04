package timeout_test

import (
	"github.com/gogo/protobuf/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/testutils"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators/timeout"
	"istio.io/api/networking/v1alpha3"
)

var _ = Describe("TimeoutDecorator", func() {
	var (
		timeoutDecorator decorators.TrafficPolicyVirtualServiceDecorator
		output           *v1alpha3.HTTPRoute
	)

	BeforeEach(func() {
		timeoutDecorator = timeout.NewTimeoutDecorator()
		output = &v1alpha3.HTTPRoute{}
	})

	It("should set retries", func() {
		registerField := func(fieldPtr, val interface{}) error {
			return nil
		}
		appliedPolicy := &discoveryv1alpha2.TrafficTargetStatus_AppliedTrafficPolicy{
			Spec: &v1alpha2.TrafficPolicySpec{
				RequestTimeout: &types.Duration{Seconds: 5},
			},
		}
		expectedTimeout := &types.Duration{Seconds: 5}
		err := timeoutDecorator.ApplyTrafficPolicyToVirtualService(appliedPolicy, nil, nil, output, registerField)
		Expect(err).ToNot(HaveOccurred())
		Expect(output.Timeout).To(Equal(expectedTimeout))
	})

	It("should not set retries if error during field registration", func() {
		testErr := eris.New("registration error")
		registerField := func(fieldPtr, val interface{}) error {
			return testErr
		}
		appliedPolicy := &discoveryv1alpha2.TrafficTargetStatus_AppliedTrafficPolicy{
			Spec: &v1alpha2.TrafficPolicySpec{
				RequestTimeout: &types.Duration{Seconds: 5},
			},
		}
		err := timeoutDecorator.ApplyTrafficPolicyToVirtualService(appliedPolicy, nil, nil, output, registerField)
		Expect(err).To(testutils.HaveInErrorChain(testErr))
	})
})
