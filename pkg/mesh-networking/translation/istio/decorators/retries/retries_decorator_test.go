package retries_test

import (
	"github.com/gogo/protobuf/types"
	"github.com/golang/protobuf/ptypes/duration"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	v1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators/retries"
	"github.com/solo-io/go-utils/testutils"
	"istio.io/api/networking/v1alpha3"
)

var _ = Describe("RetriesDecorator", func() {
	var (
		retriesDecorator decorators.TrafficPolicyVirtualServiceDecorator
		output           *v1alpha3.HTTPRoute
	)

	BeforeEach(func() {
		retriesDecorator = retries.NewRetriesDecorator()
		output = &v1alpha3.HTTPRoute{}
	})

	It("should set retries", func() {
		registerField := func(fieldPtr, val interface{}) error {
			return nil
		}
		appliedPolicy := &v1.AppliedTrafficPolicy{
			Spec: &v1.TrafficPolicySpec{
				Policy: &v1.TrafficPolicySpec_Policy{
					Retries: &v1.TrafficPolicySpec_Policy_RetryPolicy{
						Attempts:      5,
						PerTryTimeout: &duration.Duration{Seconds: 2},
					},
				},
			},
		}
		expectedRetries := &v1alpha3.HTTPRetry{
			Attempts:      5,
			PerTryTimeout: &types.Duration{Seconds: 2},
		}
		err := retriesDecorator.ApplyTrafficPolicyToVirtualService(appliedPolicy, nil, nil, output, registerField)
		Expect(err).ToNot(HaveOccurred())
		Expect(output.Retries).To(Equal(expectedRetries))
	})

	It("should not set retries if error during field registration", func() {
		testErr := eris.New("registration error")
		registerField := func(fieldPtr, val interface{}) error {
			return testErr
		}
		appliedPolicy := &v1.AppliedTrafficPolicy{
			Spec: &v1.TrafficPolicySpec{
				Policy: &v1.TrafficPolicySpec_Policy{
					Retries: &v1.TrafficPolicySpec_Policy_RetryPolicy{
						Attempts:      5,
						PerTryTimeout: &duration.Duration{Seconds: 2},
					},
				},
			},
		}
		err := retriesDecorator.ApplyTrafficPolicyToVirtualService(appliedPolicy, nil, nil, output, registerField)
		Expect(err).To(testutils.HaveInErrorChain(testErr))
	})
})
