package meshserviceutils

import (
	"github.com/rotisserie/eris"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/skv2/pkg/ezkube"
)

func FindMeshServiceForKubeService(meshServices v1alpha2.MeshServiceSlice, kubeService ezkube.ClusterResourceId) (*v1alpha2.MeshService, error) {
	for _, service := range meshServices {
		if IsMeshServiceForKubeService(service, kubeService) {
			return service, nil
		}
	}
	return nil, eris.Errorf("MeshService not found for kube ")
}

func IsMeshServiceForKubeService(meshService *v1alpha2.MeshService, kubeService ezkube.ClusterResourceId) bool {
	ref := meshService.Spec.GetKubeService().GetRef()
	if ref == nil {
		// not a kube service
		return false
	}
	return ref.GetName() == kubeService.GetName() &&
		ref.GetNamespace() == kubeService.GetNamespace() &&
		ref.GetClusterName() == kubeService.GetClusterName()
}
