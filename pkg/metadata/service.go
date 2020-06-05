package metadata

import zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"

func BuildLocalFQDN(meshService *zephyr_discovery.MeshService) string {
	return meshService.Spec.GetKubeService().GetRef().GetName() + "." +
		meshService.Spec.GetKubeService().GetRef().GetNamespace() +
		".svc.cluster.local"
}
