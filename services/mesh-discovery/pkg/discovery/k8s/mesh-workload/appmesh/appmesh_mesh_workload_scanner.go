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
	mesh_workload "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/k8s/mesh-workload"
	k8s_core_types "k8s.io/api/core/v1"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// Used to infer parent AppMesh Mesh name
	AppMeshVirtualNodeEnvVarName = "APPMESH_VIRTUAL_NODE_NAME"
	AppMeshRegionEnvVarName      = "AWS_REGION"
)

var (
	DiscoveryLabels = func() map[string]string {
		return map[string]string{
			constants.MESH_TYPE: strings.ToLower(zephyr_core_types.MeshType_APPMESH.String()),
		}
	}
)

// visible for testing
func NewAppMeshWorkloadScanner(
	ownerFetcher mesh_workload.OwnerFetcher,
	meshClient zephyr_discovery.MeshClient,
) mesh_workload.MeshWorkloadScanner {
	return &appMeshWorkloadScanner{
		ownerFetcher: ownerFetcher,
		meshClient:   meshClient,
	}
}

type appMeshWorkloadScanner struct {
	meshClient   zephyr_discovery.MeshClient
	ownerFetcher mesh_workload.OwnerFetcher
}

func (a *appMeshWorkloadScanner) ScanPod(ctx context.Context, pod *k8s_core_types.Pod, clusterName string) (*zephyr_discovery.MeshWorkload, error) {
	isAppMeshPod, appMeshName, region := a.isAppMeshPod(pod)
	if !isAppMeshPod {
		return nil, nil
	}
	deployment, err := a.ownerFetcher.GetDeployment(ctx, pod)
	if err != nil {
		return nil, err
	}
	meshRef, err := a.getMeshResourceRef(ctx, appMeshName, region, clusterName)
	if err != nil {
		return nil, err
	}
	return &zephyr_discovery.MeshWorkload{
		ObjectMeta: k8s_meta_types.ObjectMeta{
			Name:      a.buildMeshWorkloadName(deployment.GetName(), deployment.GetNamespace(), pod.GetClusterName()),
			Namespace: env.GetWriteNamespace(),
			Labels:    DiscoveryLabels(),
		},
		Spec: zephyr_discovery_types.MeshWorkloadSpec{
			KubeController: &zephyr_discovery_types.MeshWorkloadSpec_KubeController{
				KubeControllerRef: &zephyr_core_types.ResourceRef{
					Name:      deployment.GetName(),
					Namespace: deployment.GetNamespace(),
					Cluster:   pod.GetClusterName(),
				},
				Labels:             pod.GetLabels(),
				ServiceAccountName: pod.Spec.ServiceAccountName,
			},
			Mesh: meshRef,
		},
	}, nil
}

// iterate through pod's containers and check for one with name containing "appmesh" and "proxy"
// if true, return inferred AppMesh name
func (a *appMeshWorkloadScanner) isAppMeshPod(pod *k8s_core_types.Pod) (bool, string, string) {
	for _, container := range pod.Spec.Containers {
		if strings.Contains(container.Image, "appmesh") && strings.Contains(container.Image, "proxy") {
			var appMeshName string
			var region string
			for _, env := range container.Env {
				if env.Name == AppMeshVirtualNodeEnvVarName {
					// Value takes format "mesh/<appmesh-mesh-name>/virtualNode/<virtual-node-name>"
					// TODO perhaps record the virtual node name on the CRD because of AWS naming constraints between the Deployment and the correspodning VirtualNode
					// https://docs.aws.amazon.com/eks/latest/userguide/mesh-k8s-integration.html
					appMeshName = strings.Split(env.Value, "/")[1]
				} else if env.Name == AppMeshRegionEnvVarName {
					region = env.Value
				}
			}
			return true, appMeshName, region
		}
	}
	return false, "", ""
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
