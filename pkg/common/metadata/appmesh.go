package metadata

import (
	"strings"

	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
)

// The VirtualService name must be a resolvable hostname. For k8s this means <svc-name>.<svc-namespace>
// Reference: https://docs.aws.amazon.com/app-mesh/latest/userguide/virtual_services.html
func BuildVirtualServiceName(meshService *smh_discovery.MeshService) string {
	return meshService.Spec.GetKubeService().GetRef().GetName() + "." + meshService.Spec.GetKubeService().GetRef().GetNamespace()
}

func BuildVirtualRouterName(meshService *smh_discovery.MeshService) string {
	// VirtualRouter can only contain 255 letters, numbers, hyphens, and underscores.
	return strings.ReplaceAll(BuildVirtualServiceName(meshService), ".", "-") + "-virtual-router"
}

func BuildVirtualNodeName(meshWorkload *smh_discovery.MeshWorkload) string {
	return meshWorkload.Spec.GetAppmesh().GetVirtualNodeName()
}
