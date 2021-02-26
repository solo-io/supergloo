package destinationutils

import (
	"github.com/rotisserie/eris"
	v1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	"github.com/solo-io/skv2/pkg/ezkube"
)

func FindDestinationForKubeService(destinations v1.DestinationSlice, kubeService ezkube.ClusterResourceId) (*v1.Destination, error) {
	for _, dest := range destinations {
		if IsDestinationForKubeService(dest, kubeService) {
			return dest, nil
		}
	}
	return nil, eris.Errorf("Destination not found for kube service %s", sets.Key(kubeService))
}

func IsDestinationForKubeService(destination *v1.Destination, kubeService ezkube.ClusterResourceId) bool {
	ref := destination.Spec.GetKubeService().GetRef()
	if ref == nil {
		// not a kube service
		return false
	}
	return ezkube.ClusterRefsMatch(ref, kubeService)
}
