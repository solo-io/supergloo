package outlierdetection_test

import (
	"github.com/gogo/protobuf/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/testutils"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators/trafficpolicy"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators/trafficpolicy/outlierdetection"
	"istio.io/api/networking/v1alpha3"
)

var _ = Describe("OutlierDetectionDecorator", func() {
	var (
		outlierDecorator trafficpolicy.DestinationRuleDecorator
	)

	BeforeEach(func() {
		outlierDecorator = outlierdetection.NewOutlierDetectionDecorator()
	})

	It("should set outlier detection", func() {
		registerField := func(fieldPtr, val interface{}) error {
			return nil
		}
		appliedPolicy := &discoveryv1alpha2.MeshServiceStatus_AppliedTrafficPolicy{
			Spec: &v1alpha2.TrafficPolicySpec{
				OutlierDetection: &v1alpha2.TrafficPolicySpec_OutlierDetection{
					ConsecutiveErrors: 2,
				},
			},
		}
		output := &v1alpha3.DestinationRule{
			TrafficPolicy: &v1alpha3.TrafficPolicy{},
		}
		expectedOutlierDetection := &v1alpha3.OutlierDetection{
			Consecutive_5XxErrors: &types.UInt32Value{Value: 2},
			Interval:              &types.Duration{Seconds: 10},
			BaseEjectionTime:      &types.Duration{Seconds: 30},
		}
		err := outlierDecorator.ApplyToDestinationRule(
			appliedPolicy,
			nil,
			output,
			registerField,
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(output.TrafficPolicy.OutlierDetection).To(Equal(expectedOutlierDetection))
	})

	It("should not set outlier detection if error during field registration", func() {
		testErr := eris.New("registration error")
		registerField := func(fieldPtr, val interface{}) error {
			return testErr
		}
		appliedPolicy := &discoveryv1alpha2.MeshServiceStatus_AppliedTrafficPolicy{
			Spec: &v1alpha2.TrafficPolicySpec{
				OutlierDetection: &v1alpha2.TrafficPolicySpec_OutlierDetection{
					ConsecutiveErrors: 2,
				},
			},
		}
		output := &v1alpha3.DestinationRule{
			TrafficPolicy: &v1alpha3.TrafficPolicy{},
		}
		err := outlierDecorator.ApplyToDestinationRule(
			appliedPolicy,
			nil,
			output,
			registerField,
		)
		Expect(err).To(testutils.HaveInErrorChain(testErr))
	})
})
