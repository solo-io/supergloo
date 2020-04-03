package linkerd

import (
	"context"
	"fmt"

	mesh_workload "github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/discovery/mesh-workload"

	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/mesh-projects/pkg/env"
	"github.com/solo-io/mesh-projects/services/common/constants"
	core_v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	DiscoveryLabels = func() map[string]string {
		return map[string]string{
			constants.MESH_TYPE: core_types.MeshType_LINKERD.String(),
		}
	}
)

// visible for testing
func NewLinkerdMeshWorkloadScanner(ownerFetcher mesh_workload.OwnerFetcher) mesh_workload.MeshWorkloadScanner {
	return &linkerdMeshWorkloadScanner{
		deploymentFetcher: ownerFetcher,
	}
}

type linkerdMeshWorkloadScanner struct {
	deploymentFetcher mesh_workload.OwnerFetcher
}

func (i *linkerdMeshWorkloadScanner) ScanPod(ctx context.Context, pod *core_v1.Pod) (*core_types.ResourceRef, metav1.ObjectMeta, error) {
	if !i.isLinkerdPod(pod) {
		return nil, metav1.ObjectMeta{}, nil
	}
	deployment, err := i.deploymentFetcher.GetDeployment(ctx, pod)
	if err != nil {
		return nil, metav1.ObjectMeta{}, err
	}
	return &core_types.ResourceRef{
			Name:      deployment.Name,
			Namespace: deployment.Namespace,
			Cluster:   pod.ClusterName,
		}, metav1.ObjectMeta{
			Name:      i.buildMeshWorkloadName(deployment.Name, deployment.Namespace, pod.ClusterName),
			Namespace: env.GetWriteNamespace(),
			Labels:    DiscoveryLabels(),
		}, nil
}

// iterate through pod's containers and check for one with name containing "linkerd" and "proxy"
func (i *linkerdMeshWorkloadScanner) isLinkerdPod(pod *core_v1.Pod) bool {
	for _, container := range pod.Spec.Containers {
		if container.Name == "linkerd-proxy" {
			return true
		}
	}
	return false
}

func (i *linkerdMeshWorkloadScanner) buildMeshWorkloadName(deploymentName string, namespace string, clusterName string) string {
	// TODO: https://github.com/solo-io/mesh-projects/issues/141
	return fmt.Sprintf("%s-%s-%s-%s", "linkerd", deploymentName, namespace, clusterName)
}
