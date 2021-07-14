package outlierdetection

import (
	"github.com/gogo/protobuf/types"
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	v1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/gogoutils"
	networkingv1alpha3spec "istio.io/api/networking/v1alpha3"
)

const (
	decoratorName                   = "outlier-detection"
	defaultConsecutiveErrs          = 5
	defaultInterval                 = 10
	defaultEjectionTime             = 30
	defaultMaxEjectionPercent int32 = 100
)

func init() {
	decorators.Register(decoratorConstructor)
}

func decoratorConstructor(_ decorators.Parameters) decorators.Decorator {
	return NewOutlierDetectionDecorator()
}

// Handles setting OutlierDetection on a DestinationRule.
type outlierDetectionDecorator struct{}

var _ decorators.TrafficPolicyDestinationRuleDecorator = &outlierDetectionDecorator{}

func NewOutlierDetectionDecorator() *outlierDetectionDecorator {
	return &outlierDetectionDecorator{}
}

func (d *outlierDetectionDecorator) DecoratorName() string {
	return decoratorName
}

func (d *outlierDetectionDecorator) ApplyTrafficPolicyToDestinationRule(
	appliedPolicy *v1.AppliedTrafficPolicy,
	_ *discoveryv1.Destination,
	output *networkingv1alpha3spec.DestinationRule,
	registerField decorators.RegisterField,
) error {
	if outlierDetection := TranslateOutlierDetection(appliedPolicy.Spec.GetPolicy().GetOutlierDetection()); outlierDetection != nil {
		if err := registerField(&output.TrafficPolicy.OutlierDetection, outlierDetection); err != nil {
			return err
		}
		output.TrafficPolicy.OutlierDetection = outlierDetection
	}
	return nil
}

// TranslateOutlierDetection public to be used in enterprise
func TranslateOutlierDetection(
	outlierDetection *v1.TrafficPolicySpec_Policy_OutlierDetection,
) *networkingv1alpha3spec.OutlierDetection {
	if outlierDetection == nil {
		return nil
	}

	consecutiveErrs := &types.UInt32Value{Value: defaultConsecutiveErrs}
	if userConsecutiveErrs := outlierDetection.GetConsecutiveErrors(); userConsecutiveErrs != 0 {
		consecutiveErrs.Value = userConsecutiveErrs
	}
	interval := &types.Duration{Seconds: defaultInterval}
	if userInterval := outlierDetection.GetInterval(); userInterval != nil {
		interval = gogoutils.DurationProtoToGogo(userInterval)
	}
	ejectionTime := &types.Duration{Seconds: defaultEjectionTime}
	if userEjectionTime := outlierDetection.GetBaseEjectionTime(); userEjectionTime != nil {
		ejectionTime = gogoutils.DurationProtoToGogo(userEjectionTime)
	}
	maxEjectionPercent := defaultMaxEjectionPercent
	if userMaxEjectionPercent := outlierDetection.GetMaxEjectionPercent(); userMaxEjectionPercent != 0 {
		maxEjectionPercent = int32(userMaxEjectionPercent)
	}

	return &networkingv1alpha3spec.OutlierDetection{
		Consecutive_5XxErrors: consecutiveErrs,
		Interval:              interval,
		BaseEjectionTime:      ejectionTime,
		MaxEjectionPercent:    maxEjectionPercent,
	}
}
