package metadata

import smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"

func BuildVirtualServiceName(meshService *smh_discovery.MeshService) string {
	return meshService.Spec.GetKubeService().GetRef().GetName() + "-" + meshService.Spec.GetKubeService().GetRef().GetNamespace()
}

func BuildVirtualRouterName(meshService *smh_discovery.MeshService) string {
	return BuildVirtualServiceName(meshService) + "-virtual-router"
}

func BuildVirtualNodeName(meshWorkload *smh_discovery.MeshWorkload) string {
	return meshWorkload.Spec.GetAppmesh().GetVirtualNodeName()
}
