package istio_translator

import (
	"github.com/gogo/protobuf/types"
	smh_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	istio_networking_types "istio.io/api/networking/v1alpha3"
)

// Translate TrafficPolicy.OutlierDetection to Istio's OutlierDetection object.
// This functionality is needed in both TP translation and Federation (on the DestinationRule for the remote ServiceEntry)
func TranslateOutlierDetection(
	trafficPolicies []*smh_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy,
) *istio_networking_types.OutlierDetection {
	var istioOutlierDetection *istio_networking_types.OutlierDetection
	// Previous validation ensures that all OutlierDetection settings are equivalent across TrafficPolicies for this MeshService
	for _, tp := range trafficPolicies {
		outlierDetection := tp.GetTrafficPolicySpec().GetOutlierDetection()
		if outlierDetection == nil {
			continue
		}
		istioOutlierDetection = &istio_networking_types.OutlierDetection{}
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
	}
	return istioOutlierDetection
}
