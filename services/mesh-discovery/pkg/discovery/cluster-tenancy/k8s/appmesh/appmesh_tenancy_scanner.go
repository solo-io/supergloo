package appmesh_tenancy

import (
	"context"

	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	"github.com/solo-io/service-mesh-hub/pkg/metadata"
	k8s_tenancy "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/cluster-tenancy/k8s"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/mesh-platform/aws"
	"github.com/solo-io/skv2/pkg/utils"
	k8s_core_types "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type appmeshTenancyScanner struct {
	appmeshParser aws.AppMeshParser
	meshClient    zephyr_discovery.MeshClient
}

func AppMeshTenancyScannerFactoryProvider(
	appmeshParser aws.AppMeshParser,
) k8s_tenancy.ClusterTenancyScannerFactory {
	return func(meshClient zephyr_discovery.MeshClient) k8s_tenancy.ClusterTenancyRegistrar {
		return NewAppmeshTenancyScanner(
			appmeshParser,
			meshClient,
		)
	}
}

func NewAppmeshTenancyScanner(
	appmeshParser aws.AppMeshParser,
	meshClient zephyr_discovery.MeshClient,
) k8s_tenancy.ClusterTenancyRegistrar {
	return &appmeshTenancyScanner{
		appmeshParser: appmeshParser,
		meshClient:    meshClient,
	}
}

func (a *appmeshTenancyScanner) MeshForWorkload(
	ctx context.Context,
	pod *k8s_core_types.Pod,
) (*zephyr_discovery.Mesh, error) {
	appMesh, err := a.appmeshParser.ScanPodForAppMesh(pod)
	if err != nil {
		return nil, err
	}
	if appMesh == nil {
		return nil, nil
	}
	mesh, err := a.meshClient.GetMesh(
		ctx,
		client.ObjectKey{
			Name:      metadata.BuildAppMeshName(appMesh.AppMeshName, appMesh.Region, appMesh.AwsAccountID),
			Namespace: env.GetWriteNamespace(),
		},
	)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}
	return mesh, nil
}

func (a *appmeshTenancyScanner) RegisterMesh(ctx context.Context, clusterName string, mesh *zephyr_discovery.Mesh) error {
	if !isAppMesh(mesh) {
		return nil
	}
	if utils.ContainsString(mesh.Spec.GetAwsAppMesh().Clusters, clusterName) {
		return nil
	}
	mesh.Spec.GetAwsAppMesh().Clusters = append(mesh.Spec.GetAwsAppMesh().GetClusters(), clusterName)
	return a.meshClient.UpdateMesh(ctx, mesh)
}

func (a *appmeshTenancyScanner) DeregisterMesh(ctx context.Context, clusterName string, mesh *zephyr_discovery.Mesh) error {
	if !isAppMesh(mesh) {
		return nil
	}
	mesh.Spec.GetAwsAppMesh().Clusters = utils.RemoveString(mesh.Spec.GetAwsAppMesh().GetClusters(), clusterName)
	return a.meshClient.UpdateMesh(ctx, mesh)
}

func isAppMesh(mesh *zephyr_discovery.Mesh) bool {
	return mesh.Spec.GetAwsAppMesh() != nil
}
