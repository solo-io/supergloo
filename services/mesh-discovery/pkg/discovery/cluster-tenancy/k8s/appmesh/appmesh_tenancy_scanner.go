package appmesh_tenancy

import (
	"context"

	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	"github.com/solo-io/service-mesh-hub/pkg/metadata"
	aws_utils "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/compute-target/aws/parser"
	k8s_tenancy "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/cluster-tenancy/k8s"
	"github.com/solo-io/skv2/pkg/utils"
	k8s_core_types "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type appmeshTenancyScanner struct {
	appmeshScanner aws_utils.AppMeshScanner
	meshClient     zephyr_discovery.MeshClient
	remoteClient   client.Client
}

func AppMeshTenancyScannerFactoryProvider(
	appmeshParser aws_utils.AppMeshScanner,
) k8s_tenancy.ClusterTenancyScannerFactory {
	return func(
		meshClient zephyr_discovery.MeshClient,
		remoteClient client.Client,
	) k8s_tenancy.ClusterTenancyRegistrar {
		return NewAppmeshTenancyScanner(
			appmeshParser,
			meshClient,
			remoteClient,
		)
	}
}

func NewAppmeshTenancyScanner(
	appmeshScanner aws_utils.AppMeshScanner,
	meshClient zephyr_discovery.MeshClient,
	remoteClient client.Client,
) k8s_tenancy.ClusterTenancyRegistrar {
	return &appmeshTenancyScanner{
		appmeshScanner: appmeshScanner,
		meshClient:     meshClient,
		remoteClient:   remoteClient,
	}
}

func (a *appmeshTenancyScanner) MeshFromSidecar(
	ctx context.Context,
	pod *k8s_core_types.Pod,
) (*zephyr_discovery.Mesh, error) {
	appMesh, err := a.appmeshScanner.ScanPodForAppMesh(ctx, pod, a.remoteClient)
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

func (a *appmeshTenancyScanner) ClusterHostsMesh(clusterName string, mesh *zephyr_discovery.Mesh) bool {
	return utils.ContainsString(mesh.Spec.GetAwsAppMesh().Clusters, clusterName)
}

func (a *appmeshTenancyScanner) RegisterMesh(ctx context.Context, clusterName string, mesh *zephyr_discovery.Mesh) error {
	if !isAppMesh(mesh) {
		return nil
	}
	// Avoid issuing an update if not needed, an optimization that assists in reaching a steady state
	if a.ClusterHostsMesh(clusterName, mesh) {
		return nil
	}
	mesh.Spec.GetAwsAppMesh().Clusters = append(mesh.Spec.GetAwsAppMesh().GetClusters(), clusterName)
	return a.meshClient.UpdateMesh(ctx, mesh)
}

func (a *appmeshTenancyScanner) DeregisterMesh(ctx context.Context, clusterName string, mesh *zephyr_discovery.Mesh) error {
	if !isAppMesh(mesh) {
		return nil
	}
	// Avoid issuing an update if not needed, an optimization that assists in reaching a steady state
	if !a.ClusterHostsMesh(clusterName, mesh) {
		return nil
	}
	mesh.Spec.GetAwsAppMesh().Clusters = utils.RemoveString(mesh.Spec.GetAwsAppMesh().GetClusters(), clusterName)
	return a.meshClient.UpdateMesh(ctx, mesh)
}

func isAppMesh(mesh *zephyr_discovery.Mesh) bool {
	return mesh.Spec.GetAwsAppMesh() != nil
}
