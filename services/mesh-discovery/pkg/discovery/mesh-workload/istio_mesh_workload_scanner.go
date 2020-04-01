package mesh_workload

import (
	"context"
	"fmt"
	"strings"

	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/mesh-projects/pkg/env"
	"github.com/solo-io/mesh-projects/services/common/constants"
	core_v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	DiscoveryLabels = func() map[string]string {
		return map[string]string{
			constants.MESH_TYPE: core_types.MeshType_ISTIO.String(),
		}
	}
)

type IstioMeshWorkloadScanner MeshWorkloadScanner

// visible for testing
func NewIstioMeshWorkloadScanner(ownerFetcher OwnerFetcher) IstioMeshWorkloadScanner {
	return &istioMeshWorkloadScanner{
		deploymentFetcher: ownerFetcher,
	}
}

type istioMeshWorkloadScanner struct {
	deploymentFetcher OwnerFetcher
}

func (i *istioMeshWorkloadScanner) ScanPod(ctx context.Context, pod *core_v1.Pod) (*core_types.ResourceRef, metav1.ObjectMeta, error) {
	if !i.isIstioPod(pod) {
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
			Namespace: env.DefaultWriteNamespace,
			Labels:    DiscoveryLabels(),
		}, nil
}

// iterate through pod's containers and check for one with name containing "istio" and "proxy"
func (i *istioMeshWorkloadScanner) isIstioPod(pod *core_v1.Pod) bool {
	for _, container := range pod.Spec.Containers {
		if strings.Contains(container.Image, "istio") && strings.Contains(container.Image, "proxy") {
			return true
		}
	}
	return false
}

func (i *istioMeshWorkloadScanner) buildMeshWorkloadName(deploymentName string, namespace string, clusterName string) string {
	// TODO: https://github.com/solo-io/mesh-projects/issues/141
	return fmt.Sprintf("%s-%s-%s-%s", "istio", deploymentName, namespace, clusterName)
}
