package utils

import (
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	networkingv1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/destinationutils"
)

// return true if TrafficPolicy references this Destination as a TrafficShift and specifies subsets
// because the subsets must be defined on the traffic shift destination's DestinationRule
// TODO(harveyxia): move this logic into the Applier
func ReferencedByTrafficShiftSubset(destination *discoveryv1.Destination, trafficPolicy *networkingv1.TrafficPolicy) bool {
	shouldTranslate := false

	trafficShiftDestinations := trafficPolicy.Spec.GetPolicy().GetTrafficShift().GetDestinations()
	for _, trafficShiftDestination := range trafficShiftDestinations {
		kubeService := trafficShiftDestination.GetKubeService()
		if len(kubeService.GetSubset()) > 0 && destinationutils.IsDestinationForKubeService(destination, kubeService) {
			shouldTranslate = true
		}
	}

	return shouldTranslate
}
