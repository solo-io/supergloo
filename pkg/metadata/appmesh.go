package metadata

import "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"

func BuildVirtualServiceName(meshService *v1alpha1.MeshService) string {
	return meshService.Spec.GetKubeService().GetRef().GetName()
}

func BuildVirtualRouterName(meshService *v1alpha1.MeshService) string {
	return BuildVirtualServiceName(meshService) + "-virtual-router"
}
