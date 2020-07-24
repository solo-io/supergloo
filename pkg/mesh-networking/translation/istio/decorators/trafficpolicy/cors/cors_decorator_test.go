package cors_test

import (
	"github.com/gogo/protobuf/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/testutils"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators/trafficpolicy"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators/trafficpolicy/cors"
	"istio.io/api/networking/v1alpha3"
)

var _ = Describe("CorsDecorator", func() {
	var (
		corsDecorator trafficpolicy.VirtualServiceDecorator
	)

	BeforeEach(func() {
		corsDecorator = cors.NewCorsDecorator()
	})

	It("should set cors policy", func() {
		registerField := func(fieldPtr, val interface{}) error {
			return nil
		}
		appliedPolicy := &discoveryv1alpha2.MeshServiceStatus_AppliedTrafficPolicy{
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
		output := &v1alpha3.HTTPRoute{}
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
		err := corsDecorator.ApplyToVirtualService(
			appliedPolicy,
			nil,
			output,
			registerField,
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(output.CorsPolicy).To(Equal(expectedCorsPolicy))
	})

	It("should not set CorsPolicy if error during field registration", func() {
		testErr := eris.New("registration error")
		registerField := func(fieldPtr, val interface{}) error {
			return testErr
		}
		appliedPolicy := &discoveryv1alpha2.MeshServiceStatus_AppliedTrafficPolicy{
			Spec: &v1alpha2.TrafficPolicySpec{
				CorsPolicy: &v1alpha2.TrafficPolicySpec_CorsPolicy{},
			},
		}
		output := &v1alpha3.HTTPRoute{}
		err := corsDecorator.ApplyToVirtualService(
			appliedPolicy,
			nil,
			output,
			registerField,
		)
		Expect(err).To(testutils.HaveInErrorChain(testErr))
		Expect(output.CorsPolicy).To(BeNil())
	})
})
