package meshserviceutils

import (
	"github.com/rotisserie/eris"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/resourceidutils"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	"github.com/solo-io/skv2/pkg/ezkube"
)

func FindMeshServiceForKubeService(meshServices v1alpha2.MeshServiceSlice, kubeService ezkube.ClusterResourceId) (*v1alpha2.MeshService, error) {
	for _, service := range meshServices {
		if IsMeshServiceForKubeService(service, kubeService) {
			return service, nil
		}
	}
	return nil, eris.Errorf("MeshService not found for kube service %s", sets.Key(kubeService))
}

func IsMeshServiceForKubeService(meshService *v1alpha2.MeshService, kubeService ezkube.ClusterResourceId) bool {
	ref := meshService.Spec.GetKubeService().GetRef()
	if ref == nil {
		// not a kube service
		return false
	}
	return resourceidutils.ClusterRefsEqual(ref, kubeService)
}
