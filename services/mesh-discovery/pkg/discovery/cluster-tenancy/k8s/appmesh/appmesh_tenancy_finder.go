package appmesh_tenancy

import (
	"context"

	"github.com/solo-io/go-utils/stringutils"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	k8s_tenancy "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/cluster-tenancy/k8s"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/mesh-platform/aws"
	k8s_core_types "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type appmeshTenancyScanner struct {
	appmeshParser aws.AppMeshParser
}

func NewAppmeshTenancyScanner(
	appmeshParser aws.AppMeshParser,
) k8s_tenancy.ClusterTenancyScanner {
	return &appmeshTenancyScanner{
		appmeshParser: appmeshParser,
	}
}

func (a *appmeshTenancyScanner) UpdateMeshTenancy(
	ctx context.Context,
	clusterName string,
	pod *k8s_core_types.Pod,
	meshClient zephyr_discovery.MeshClient,
) error {
	appMesh, err := a.appmeshParser.ScanPodForAppMesh(pod)
	if err != nil {
		return err
	}
	if appMesh == nil {
		return nil
	}
	mesh, err := meshClient.GetMesh(ctx, client.ObjectKey{Name: appMesh.AppMeshName, Namespace: env.GetWriteNamespace()})
	if errors.IsNotFound(err) {
		// Mesh has not yet been discovered, do nothing (wait for Mesh discovery to process the Mesh)
		return nil
	} else if !stringutils.ContainsString(clusterName, mesh.Spec.GetAwsAppMesh().GetClusters()) {
		// Record this Mesh as a tenant of this cluster
		mesh.Spec.GetAwsAppMesh().Clusters = append(mesh.Spec.GetAwsAppMesh().GetClusters(), clusterName)
		return meshClient.UpdateMesh(ctx, mesh)
	}
	return nil
}
