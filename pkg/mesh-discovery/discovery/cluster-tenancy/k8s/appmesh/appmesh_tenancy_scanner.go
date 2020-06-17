package appmesh_tenancy

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	aws_utils "github.com/solo-io/service-mesh-hub/pkg/common/aws/parser"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/common/container-runtime"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/metadata"
	k8s_tenancy "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/discovery/cluster-tenancy/k8s"
	"github.com/solo-io/skv2/pkg/utils"
	k8s_core_types "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type appmeshTenancyScanner struct {
	appmeshScanner      aws_utils.AppMeshScanner
	awsAccountIdFetcher aws_utils.AwsAccountIdFetcher
	meshClient          smh_discovery.MeshClient
}

func AppMeshTenancyScannerFactoryProvider(
	appmeshParser aws_utils.AppMeshScanner,
) k8s_tenancy.ClusterTenancyScannerFactory {
	return func(
		meshClient smh_discovery.MeshClient,
	) k8s_tenancy.ClusterTenancyRegistrar {
		return NewAppmeshTenancyScanner(
			appmeshParser,
			meshClient,
		)
	}
}

func NewAppmeshTenancyScanner(
	appmeshScanner aws_utils.AppMeshScanner,
	meshClient smh_discovery.MeshClient,
	//awsAccountIdFetcher aws_utils.AwsAccountIdFetcher, TODO(harveyxia) fix this wiring when we get to Mesh discovery migration
) k8s_tenancy.ClusterTenancyRegistrar {
	return &appmeshTenancyScanner{
		appmeshScanner:      appmeshScanner,
		meshClient:          meshClient,
		awsAccountIdFetcher: nil,
	}
}

func (a *appmeshTenancyScanner) MeshFromSidecar(
	ctx context.Context,
	pod *k8s_core_types.Pod,
	clusterName string,
) (*smh_discovery.Mesh, error) {
	awsAccountId, err := a.awsAccountIdFetcher.GetEksAccountId(ctx, clusterName)
	if err != nil {
		contextutils.LoggerFrom(ctx).Warnf("Error fetching AWS Account ID from ConfigMap: %+v", err)
	}
	appMesh, err := a.appmeshScanner.ScanPodForAppMesh(pod, awsAccountId)
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
			Namespace: container_runtime.GetWriteNamespace(),
		},
	)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}
	return mesh, nil
}

func (a *appmeshTenancyScanner) ClusterHostsMesh(clusterName string, mesh *smh_discovery.Mesh) bool {
	return utils.ContainsString(mesh.Spec.GetAwsAppMesh().GetClusters(), clusterName)
}

func (a *appmeshTenancyScanner) RegisterMesh(ctx context.Context, clusterName string, mesh *smh_discovery.Mesh) error {
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

func (a *appmeshTenancyScanner) DeregisterMesh(ctx context.Context, clusterName string, mesh *smh_discovery.Mesh) error {
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

func isAppMesh(mesh *smh_discovery.Mesh) bool {
	return mesh.Spec.GetAwsAppMesh() != nil
}
