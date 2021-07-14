package timeout_test

import (
	"github.com/gogo/protobuf/types"
	"github.com/golang/protobuf/ptypes/duration"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	v1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators/timeout"
	"github.com/solo-io/go-utils/testutils"
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
		appliedPolicy := &v1.AppliedTrafficPolicy{
			Spec: &v1.TrafficPolicySpec{
				Policy: &v1.TrafficPolicySpec_Policy{
					RequestTimeout: &duration.Duration{Seconds: 5},
				},
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
		appliedPolicy := &v1.AppliedTrafficPolicy{
			Spec: &v1.TrafficPolicySpec{
				Policy: &v1.TrafficPolicySpec_Policy{
					RequestTimeout: &duration.Duration{Seconds: 5},
				},
			},
		}
		err := timeoutDecorator.ApplyTrafficPolicyToVirtualService(appliedPolicy, nil, nil, output, registerField)
		Expect(err).To(testutils.HaveInErrorChain(testErr))
	})
})
