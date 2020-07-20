package outlierdetection

import (
	"github.com/gogo/protobuf/types"
	discoveryv1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/decorators"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/istio/decorators/trafficpolicy"
	istiov1alpha3spec "istio.io/api/networking/v1alpha3"
)

const (
	decoratorName          = "outlier-detection"
	defaultConsecutiveErrs = 5
	defaultInterval        = 10
	defaultEjectionTime    = 30
)

func init() {
	decorators.Register(decoratorConstructor)
}

func decoratorConstructor(_ decorators.Parameters) decorators.Decorator {
	return NewOutlierDetectionDecorator()
}

// Handles setting OutlierDetection on a DestinationRule.
type outlierDetectionDecorator struct{}

var _ trafficpolicy.DestinationRuleDecorator = &outlierDetectionDecorator{}

func NewOutlierDetectionDecorator() *outlierDetectionDecorator {
	return &outlierDetectionDecorator{}
}

func (d *outlierDetectionDecorator) DecoratorName() string {
	return decoratorName
}

func (d *outlierDetectionDecorator) ApplyToDestinationRule(
	appliedPolicy *discoveryv1alpha1.MeshServiceStatus_AppliedTrafficPolicy,
	_ *discoveryv1alpha1.MeshService,
	output *istiov1alpha3spec.DestinationRule,
	registerField decorators.RegisterField,
) error {
	outlierDetection := d.translateOutlierDetection(appliedPolicy.Spec)
	if outlierDetection != nil {
		if err := registerField(&output.TrafficPolicy.OutlierDetection, outlierDetection); err != nil {
			return err
		}
		output.TrafficPolicy.OutlierDetection = outlierDetection
	}
	return nil
}

func (d *outlierDetectionDecorator) translateOutlierDetection(
	trafficPolicy *v1alpha1.TrafficPolicySpec,
) *istiov1alpha3spec.OutlierDetection {
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

	return &istiov1alpha3spec.OutlierDetection{
		Consecutive_5XxErrors: consecutiveErrs,
		Interval:              interval,
		BaseEjectionTime:      ejectionTime,
	}
}
