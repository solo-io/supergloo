package k8s_tenancy

import (
	"github.com/solo-io/go-utils/stringutils"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
)

func ClusterHostsMesh(clusterName string, mesh *zephyr_discovery.Mesh) bool {
	if mesh == nil {
		return false
	}
	if mesh.Spec.GetAwsAppMesh() != nil {
		return stringutils.ContainsString(clusterName, mesh.Spec.GetAwsAppMesh().GetClusters())
	} else {
		return mesh.Spec.GetCluster().GetName() == clusterName
	}
}
