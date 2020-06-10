package metadata

import smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"

func BuildLocalFQDN(meshService *smh_discovery.MeshService) string {
	return meshService.Spec.GetKubeService().GetRef().GetName() + "." +
		meshService.Spec.GetKubeService().GetRef().GetNamespace() +
		".svc.cluster.local"
}
