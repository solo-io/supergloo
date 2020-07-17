package outlierdetection

import (
	"github.com/gogo/protobuf/types"
	discoveryv1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	"github.com/solo-io/smh/pkg/mesh-networking/plugins"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/istio/plugins/trafficpolicy"
	istiov1alpha3spec "istio.io/api/networking/v1alpha3"
)

const (
	pluginName = "outlier-detection"
)

func init() {
	plugins.Register(pluginConstructor)
}

func pluginConstructor(_ plugins.Parameters) plugins.Plugin {
	return NewOutlierDetectionPlugin()
}

// Handles setting OutlierDetection on a DestinationRule.
type outlierDetectionPlugin struct{}

var _ trafficpolicy.DestinationRuleDecorator = &outlierDetectionPlugin{}

func NewOutlierDetectionPlugin() *outlierDetectionPlugin {
	return &outlierDetectionPlugin{}
}

func (o *outlierDetectionPlugin) PluginName() string {
	return pluginName
}

func (o *outlierDetectionPlugin) DecorateDestinationRule(
	appliedPolicy *discoveryv1alpha1.MeshServiceStatus_AppliedTrafficPolicy,
	_ *discoveryv1alpha1.MeshService,
	output *istiov1alpha3spec.DestinationRule,
	registerField plugins.RegisterField,
) error {
	outlierDetection := o.translateOutlierDetection(appliedPolicy.GetSpec())
	if outlierDetection != nil {
		if err := registerField(&output.TrafficPolicy.OutlierDetection, outlierDetection); err != nil {
			return err
		}
		output.TrafficPolicy.OutlierDetection = outlierDetection
	}
	return nil
}

func (o *outlierDetectionPlugin) translateOutlierDetection(
	trafficPolicy *v1alpha1.TrafficPolicySpec,
) *istiov1alpha3spec.OutlierDetection {
	outlierDetection := trafficPolicy.GetOutlierDetection()
	if outlierDetection == nil {
		return nil
	}
	istioOutlierDetection := &istiov1alpha3spec.OutlierDetection{}
	// Set defaults if needed
	if consecutiveErrs := outlierDetection.GetConsecutiveErrors(); consecutiveErrs != 0 {
		istioOutlierDetection.Consecutive_5XxErrors = &types.UInt32Value{Value: consecutiveErrs}
	} else {
		istioOutlierDetection.Consecutive_5XxErrors = &types.UInt32Value{Value: 5}
	}
	if interval := outlierDetection.GetInterval(); interval != nil {
		istioOutlierDetection.Interval = interval
	} else {
		istioOutlierDetection.Interval = &types.Duration{Seconds: 10}
	}
	if ejectionTime := outlierDetection.GetBaseEjectionTime(); ejectionTime != nil {
		istioOutlierDetection.BaseEjectionTime = ejectionTime
	} else {
		istioOutlierDetection.BaseEjectionTime = &types.Duration{Seconds: 30}
	}
	return istioOutlierDetection
}
