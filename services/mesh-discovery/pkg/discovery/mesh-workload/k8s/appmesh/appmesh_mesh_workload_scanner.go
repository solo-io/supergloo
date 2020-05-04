package appmesh

import (
	"context"
	"fmt"
	"strings"

	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	"github.com/solo-io/service-mesh-hub/services/common/constants"
	aws_utils "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/compute-target/aws/parser"
	meshworkload_discovery "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh-workload/k8s"
	k8s_core_types "k8s.io/api/core/v1"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	DiscoveryLabels = func() map[string]string {
		return map[string]string{
			constants.MESH_TYPE: strings.ToLower(zephyr_core_types.MeshType_APPMESH.String()),
		}
	}
)

func AppMeshWorkloadScannerFactoryProvider(
	appMeshParser aws_utils.AppMeshParser,
) meshworkload_discovery.MeshWorkloadScannerFactory {
	return func(
		ownerFetcher meshworkload_discovery.OwnerFetcher,
		meshClient zephyr_discovery.MeshClient,
	) meshworkload_discovery.MeshWorkloadScanner {
		return NewAppMeshWorkloadScanner(ownerFetcher, appMeshParser, meshClient)
	}
}

// visible for testing
func NewAppMeshWorkloadScanner(
	ownerFetcher meshworkload_discovery.OwnerFetcher,
	appMeshParser aws_utils.AppMeshParser,
	meshClient zephyr_discovery.MeshClient,
) meshworkload_discovery.MeshWorkloadScanner {
	return &appMeshWorkloadScanner{
		ownerFetcher:  ownerFetcher,
		meshClient:    meshClient,
		appMeshParser: appMeshParser,
	}
}

type appMeshWorkloadScanner struct {
	ownerFetcher  meshworkload_discovery.OwnerFetcher
	appMeshParser aws_utils.AppMeshParser
	meshClient    zephyr_discovery.MeshClient
}

func (a *appMeshWorkloadScanner) ScanPod(ctx context.Context, pod *k8s_core_types.Pod, clusterName string) (*zephyr_discovery.MeshWorkload, error) {
	appMeshPod, err := a.appMeshParser.ScanPodForAppMesh(pod)
	if err != nil {
		return nil, err
	}
	if appMeshPod == nil {
		return nil, nil
	}
	deployment, err := a.ownerFetcher.GetDeployment(ctx, pod)
	if err != nil {
		return nil, err
	}
	meshRef, err := a.getMeshResourceRef(ctx, appMeshPod.AppMeshName, appMeshPod.Region, clusterName)
	if err != nil {
		return nil, err
	}
	return &zephyr_discovery.MeshWorkload{
		ObjectMeta: k8s_meta_types.ObjectMeta{
			Name:      a.buildMeshWorkloadName(deployment.GetName(), deployment.GetNamespace(), clusterName),
			Namespace: env.GetWriteNamespace(),
			Labels:    DiscoveryLabels(),
		},
		Spec: zephyr_discovery_types.MeshWorkloadSpec{
			KubeController: &zephyr_discovery_types.MeshWorkloadSpec_KubeController{
				KubeControllerRef: &zephyr_core_types.ResourceRef{
					Name:      deployment.GetName(),
					Namespace: deployment.GetNamespace(),
					Cluster:   clusterName,
				},
				Labels:             pod.GetLabels(),
				ServiceAccountName: pod.Spec.ServiceAccountName,
			},
			Mesh: meshRef,
		},
	}, nil
}

func (a *appMeshWorkloadScanner) getMeshResourceRef(ctx context.Context, meshName, region, clusterName string) (*zephyr_core_types.ResourceRef, error) {
	meshList, err := a.meshClient.ListMesh(ctx)
	if err != nil {
		return nil, err
	}
	for _, mesh := range meshList.Items {
		// To support multitenant AppMesh on a single cluster, disambiguate parent mesh with name and region
		// We assume that the kubernetes cluster is managed by only a single AWS account
		if mesh.Spec.GetCluster().GetName() == clusterName &&
			mesh.Spec.GetAwsAppMesh().GetName() == meshName &&
			mesh.Spec.GetAwsAppMesh().GetRegion() == region {
			return &zephyr_core_types.ResourceRef{
				Name:      mesh.GetName(),
				Namespace: mesh.GetNamespace(),
				Cluster:   clusterName,
			}, nil
		}
	}
	return nil, nil
}

func (a *appMeshWorkloadScanner) buildMeshWorkloadName(deploymentName string, namespace string, clusterName string) string {
	// TODO: https://github.com/solo-io/service-mesh-hub/issues/141
	return fmt.Sprintf("%s-%s-%s-%s", "appmesh", deploymentName, namespace, clusterName)
}
