package appmesh

import (
	"context"
	"fmt"
	"strings"

	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	"github.com/solo-io/service-mesh-hub/services/common/constants"
	mesh_workload "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/k8s/mesh-workload"
	k8s_core_types "k8s.io/api/core/v1"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// Used to infer parent AppMesh Mesh name
	AppMeshVirtualNodeEnvVarName = "APPMESH_VIRTUAL_NODE_NAME"
)

var (
	DiscoveryLabels = func() map[string]string {
		return map[string]string{
			constants.MESH_TYPE: strings.ToLower(zephyr_core_types.MeshType_APPMESH.String()),
		}
	}
)

// visible for testing
func NewAppMeshWorkloadScanner(ownerFetcher mesh_workload.OwnerFetcher) mesh_workload.MeshWorkloadScanner {
	return &appMeshWorkloadScanner{
		deploymentFetcher: ownerFetcher,
	}
}

type appMeshWorkloadScanner struct {
	deploymentFetcher mesh_workload.OwnerFetcher
}

func (i *appMeshWorkloadScanner) ScanPod(ctx context.Context, pod *k8s_core_types.Pod) (*zephyr_core_types.ResourceRef, k8s_meta_types.ObjectMeta, string, error) {
	isAppMeshPod, appMeshName := i.isAppMeshPod(pod)
	if !isAppMeshPod {
		return nil, k8s_meta_types.ObjectMeta{}, "", nil
	}
	deployment, err := i.deploymentFetcher.GetDeployment(ctx, pod)
	if err != nil {
		return nil, k8s_meta_types.ObjectMeta{}, "", err
	}
	return &zephyr_core_types.ResourceRef{
			Name:      deployment.Name,
			Namespace: deployment.Namespace,
			Cluster:   pod.ClusterName,
		}, k8s_meta_types.ObjectMeta{
			Name:      i.buildMeshWorkloadName(deployment.Name, deployment.Namespace, pod.ClusterName),
			Namespace: env.GetWriteNamespace(),
			Labels:    DiscoveryLabels(),
		}, appMeshName, nil
}

// iterate through pod's containers and check for one with name containing "appmesh" and "proxy"
// if true, return inferred AppMesh name
func (i *appMeshWorkloadScanner) isAppMeshPod(pod *k8s_core_types.Pod) (bool, string) {
	for _, container := range pod.Spec.Containers {
		if strings.Contains(container.Image, "appmesh") && strings.Contains(container.Image, "envoy") {
			var appMeshName string
			for _, env := range container.Env {
				if env.Name == AppMeshVirtualNodeEnvVarName {
					// Value takes format mesh/<appmesh-mesh-name>/virtualNode/<virtual-node-name>"
					// TODO perhaps record the virtual node name on the CRD because of AWS naming constraints between the Deployment and the correspodning VirtualNode
					// https://docs.aws.amazon.com/eks/latest/userguide/mesh-k8s-integration.html
					appMeshName = strings.Split(env.Value, "/")[1]
				}
			}
			return true, appMeshName
		}
	}
	return false, ""
}

func (i *appMeshWorkloadScanner) buildMeshWorkloadName(deploymentName string, namespace string, clusterName string) string {
	// TODO: https://github.com/solo-io/service-mesh-hub/issues/141
	return fmt.Sprintf("%s-%s-%s-%s", "appmesh", deploymentName, namespace, clusterName)
}
