package traffictargetutils

import (
	"github.com/rotisserie/eris"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	"github.com/solo-io/skv2/pkg/ezkube"
)

func FindTrafficTargetForKubeService(trafficTargets v1alpha2.TrafficTargetSlice, kubeService ezkube.ClusterResourceId) (*v1alpha2.TrafficTarget, error) {
	for _, service := range trafficTargets {
		if IsTrafficTargetForKubeService(service, kubeService) {
			return service, nil
		}
	}
	return nil, eris.Errorf("TrafficTarget not found for kube service %s", sets.Key(kubeService))
}

func IsTrafficTargetForKubeService(trafficTarget *v1alpha2.TrafficTarget, kubeService ezkube.ClusterResourceId) bool {
	ref := trafficTarget.Spec.GetKubeService().GetRef()
	if ref == nil {
		// not a kube service
		return false
	}
	return ezkube.RefsMatch(ref, kubeService)
}
