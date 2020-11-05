package cors_test

import (
	"github.com/gogo/protobuf/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/testutils"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators/cors"
	"istio.io/api/networking/v1alpha3"
)

var _ = Describe("CorsDecorator", func() {
	var (
		corsDecorator decorators.TrafficPolicyVirtualServiceDecorator
		output        *v1alpha3.HTTPRoute
	)

	BeforeEach(func() {
		corsDecorator = cors.NewCorsDecorator()
		output = &v1alpha3.HTTPRoute{}
	})

	It("should set cors policy", func() {
		registerField := func(fieldPtr, val interface{}) error {
			return nil
		}
		appliedPolicy := &discoveryv1alpha2.TrafficTargetStatus_AppliedTrafficPolicy{
			Spec: &v1alpha2.TrafficPolicySpec{
				CorsPolicy: &v1alpha2.TrafficPolicySpec_CorsPolicy{
					AllowOrigins: []*v1alpha2.TrafficPolicySpec_StringMatch{
						{MatchType: &v1alpha2.TrafficPolicySpec_StringMatch_Exact{Exact: "exact"}},
						{MatchType: &v1alpha2.TrafficPolicySpec_StringMatch_Prefix{Prefix: "prefix"}},
						{MatchType: &v1alpha2.TrafficPolicySpec_StringMatch_Regex{Regex: "regex"}},
					},
					AllowMethods:     []string{"GET", "POST"},
					AllowHeaders:     []string{"Header1", "Header2"},
					ExposeHeaders:    []string{"ExposedHeader1", "ExposedHeader2"},
					MaxAge:           &types.Duration{Seconds: 1},
					AllowCredentials: &types.BoolValue{Value: false},
				},
			},
		}
		expectedCorsPolicy := &v1alpha3.CorsPolicy{
			AllowOrigins: []*v1alpha3.StringMatch{
				{MatchType: &v1alpha3.StringMatch_Exact{Exact: "exact"}},
				{MatchType: &v1alpha3.StringMatch_Prefix{Prefix: "prefix"}},
				{MatchType: &v1alpha3.StringMatch_Regex{Regex: "regex"}},
			},
			AllowMethods:     []string{"GET", "POST"},
			AllowHeaders:     []string{"Header1", "Header2"},
			ExposeHeaders:    []string{"ExposedHeader1", "ExposedHeader2"},
			MaxAge:           &types.Duration{Seconds: 1},
			AllowCredentials: &types.BoolValue{Value: false},
		}
		err := corsDecorator.ApplyTrafficPolicyToVirtualService(appliedPolicy, nil, nil, output, registerField)
		Expect(err).ToNot(HaveOccurred())
		Expect(output.CorsPolicy).To(Equal(expectedCorsPolicy))
	})

	It("should not set CorsPolicy if error during field registration", func() {
		testErr := eris.New("registration error")
		registerField := func(fieldPtr, val interface{}) error {
			return testErr
		}
		appliedPolicy := &discoveryv1alpha2.TrafficTargetStatus_AppliedTrafficPolicy{
			Spec: &v1alpha2.TrafficPolicySpec{
				CorsPolicy: &v1alpha2.TrafficPolicySpec_CorsPolicy{},
			},
		}
		err := corsDecorator.ApplyTrafficPolicyToVirtualService(appliedPolicy, nil, nil, output, registerField)
		Expect(err).To(testutils.HaveInErrorChain(testErr))
		Expect(output.CorsPolicy).To(BeNil())
	})
})
