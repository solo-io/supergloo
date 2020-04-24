package linkerd

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

var (
	DiscoveryLabels = func() map[string]string {
		return map[string]string{
			constants.MESH_TYPE: strings.ToLower(zephyr_core_types.MeshType_LINKERD.String()),
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

func (i *linkerdMeshWorkloadScanner) ScanPod(ctx context.Context, pod *k8s_core_types.Pod) (*zephyr_core_types.ResourceRef, k8s_meta_types.ObjectMeta, error) {
	if !i.isLinkerdPod(pod) {
		return nil, k8s_meta_types.ObjectMeta{}, nil
	}
	deployment, err := i.deploymentFetcher.GetDeployment(ctx, pod)
	if err != nil {
		return nil, k8s_meta_types.ObjectMeta{}, err
	}
	return &zephyr_core_types.ResourceRef{
			Name:      deployment.Name,
			Namespace: deployment.Namespace,
			Cluster:   pod.ClusterName,
		}, k8s_meta_types.ObjectMeta{
			Name:      i.buildMeshWorkloadName(deployment.Name, deployment.Namespace, pod.ClusterName),
			Namespace: env.GetWriteNamespace(),
			Labels:    DiscoveryLabels(),
		}, nil
}

// iterate through pod's containers and check for one with name containing "linkerd" and "proxy"
func (i *linkerdMeshWorkloadScanner) isLinkerdPod(pod *k8s_core_types.Pod) bool {
	for _, container := range pod.Spec.Containers {
		if container.Name == "linkerd-proxy" {
			return true
		}
	}
	return false
}

func (i *linkerdMeshWorkloadScanner) buildMeshWorkloadName(deploymentName string, namespace string, clusterName string) string {
	// TODO: https://github.com/solo-io/service-mesh-hub/issues/141
	return fmt.Sprintf("%s-%s-%s-%s", "linkerd", deploymentName, namespace, clusterName)
}
