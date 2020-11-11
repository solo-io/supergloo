package failoverserviceutils

import (
	discoveryv1alpha2 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2"
	networkingv1alpha2 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1alpha2"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/traffictargetutils"
)

// Return true if the FailoverService's list of backing services references the specified TrafficTarget
func ContainsTrafficTarget(failoverService *networkingv1alpha2.FailoverService, trafficTarget *discoveryv1alpha2.TrafficTarget) bool {
	for _, backingService := range failoverService.Spec.BackingServices {
		switch backingServiceType := backingService.BackingServiceType.(type) {
		case *networkingv1alpha2.FailoverServiceSpec_BackingService_KubeService:
			if traffictargetutils.IsTrafficTargetForKubeService(trafficTarget, backingServiceType.KubeService) {
				return true
			}
		}
	}
	return false
}
