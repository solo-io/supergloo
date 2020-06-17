package metadata

import (
	"fmt"

	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	k8s_core_types "k8s.io/api/core/v1"
)

func BuildLocalFQDN(meshService *smh_discovery.MeshService) string {
	return meshService.Spec.GetKubeService().GetRef().GetName() + "." +
		meshService.Spec.GetKubeService().GetRef().GetNamespace() +
		".svc.cluster.local"
}

func BuildMeshServiceName(service *k8s_core_types.Service, clusterName string) string {
	return fmt.Sprintf("%s-%s-%s", service.GetName(), service.GetNamespace(), clusterName)
}
