package cors_test

import (
	"github.com/gogo/protobuf/types"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/wrappers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	commonv1 "github.com/solo-io/gloo-mesh/pkg/api/common.mesh.gloo.solo.io/v1"
	v1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators/cors"
	"github.com/solo-io/go-utils/testutils"
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
		appliedPolicy := &v1.AppliedTrafficPolicy{
			Spec: &v1.TrafficPolicySpec{
				Policy: &v1.TrafficPolicySpec_Policy{
					CorsPolicy: &v1.TrafficPolicySpec_Policy_CorsPolicy{
						AllowOrigins: []*commonv1.StringMatch{
							{MatchType: &commonv1.StringMatch_Exact{Exact: "exact"}},
							{MatchType: &commonv1.StringMatch_Prefix{Prefix: "prefix"}},
							{MatchType: &commonv1.StringMatch_Regex{Regex: "regex"}},
						},
						AllowMethods:     []string{"GET", "POST"},
						AllowHeaders:     []string{"Header1", "Header2"},
						ExposeHeaders:    []string{"ExposedHeader1", "ExposedHeader2"},
						MaxAge:           &duration.Duration{Seconds: 1},
						AllowCredentials: &wrappers.BoolValue{Value: false},
					},
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
		appliedPolicy := &v1.AppliedTrafficPolicy{
			Spec: &v1.TrafficPolicySpec{
				Policy: &v1.TrafficPolicySpec_Policy{
					CorsPolicy: &v1.TrafficPolicySpec_Policy_CorsPolicy{},
				},
			},
		}
		err := corsDecorator.ApplyTrafficPolicyToVirtualService(appliedPolicy, nil, nil, output, registerField)
		Expect(err).To(testutils.HaveInErrorChain(testErr))
		Expect(output.CorsPolicy).To(BeNil())
	})
})
