package metadata

import zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"

func BuildVirtualServiceName(meshService *zephyr_discovery.MeshService) string {
	return meshService.Spec.GetKubeService().GetRef().GetName() + meshService.Spec.GetKubeService().GetRef().GetNamespace()
}

func BuildVirtualRouterName(meshService *zephyr_discovery.MeshService) string {
	return BuildVirtualServiceName(meshService) + "-virtual-router"
}

func BuildVirtualNodeName(meshWorkload *zephyr_discovery.MeshWorkload) string {
	return meshWorkload.Spec.GetAppmesh().GetVirtualNodeName()
}
