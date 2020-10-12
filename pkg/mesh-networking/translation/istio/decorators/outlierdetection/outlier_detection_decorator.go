package outlierdetection

import (
	"github.com/gogo/protobuf/types"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators"
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
	appliedPolicy *discoveryv1alpha2.TrafficTargetStatus_AppliedTrafficPolicy,
	_ *discoveryv1alpha2.TrafficTarget,
	output *networkingv1alpha3spec.DestinationRule,
	registerField decorators.RegisterField,
) error {
	if outlierDetection := d.translateOutlierDetection(appliedPolicy.Spec); outlierDetection != nil {
		if err := registerField(&output.TrafficPolicy.OutlierDetection, outlierDetection); err != nil {
			return err
		}
		output.TrafficPolicy.OutlierDetection = outlierDetection
	}
	return nil
}

func (d *outlierDetectionDecorator) translateOutlierDetection(
	trafficPolicy *v1alpha2.TrafficPolicySpec,
) *networkingv1alpha3spec.OutlierDetection {
	outlierDetection := trafficPolicy.GetOutlierDetection()
	if outlierDetection == nil {
		return nil
	}

	consecutiveErrs := &types.UInt32Value{Value: defaultConsecutiveErrs}
	if userConsecutiveErrs := outlierDetection.GetConsecutiveErrors(); userConsecutiveErrs != 0 {
		consecutiveErrs.Value = userConsecutiveErrs
	}
	interval := &types.Duration{Seconds: defaultInterval}
	if userInterval := outlierDetection.GetInterval(); userInterval != nil {
		interval = userInterval
	}
	ejectionTime := &types.Duration{Seconds: defaultEjectionTime}
	if userEjectionTime := outlierDetection.GetBaseEjectionTime(); userEjectionTime != nil {
		ejectionTime = userEjectionTime
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
