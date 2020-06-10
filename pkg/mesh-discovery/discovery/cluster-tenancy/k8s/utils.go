package k8s_tenancy

import (
	"github.com/solo-io/go-utils/stringutils"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
)

// Uses metadata recorded on the Mesh CRD to determine whether mesh is hosted on cluster.
func ClusterHostsMesh(clusterName string, mesh *smh_discovery.Mesh) bool {
	if mesh == nil {
		return false
	}
	if mesh.Spec.GetAwsAppMesh() != nil {
		return stringutils.ContainsString(clusterName, mesh.Spec.GetAwsAppMesh().GetClusters())
	} else {
		return mesh.Spec.GetCluster().GetName() == clusterName
	}
}
