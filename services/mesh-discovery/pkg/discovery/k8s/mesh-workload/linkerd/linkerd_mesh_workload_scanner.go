package linkerd

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

var (
	DiscoveryLabels = func() map[string]string {
		return map[string]string{
			constants.MESH_TYPE: strings.ToLower(zephyr_core_types.MeshType_LINKERD.String()),
		}
	}
)

// visible for testing
func NewLinkerdMeshWorkloadScanner(
	ownerFetcher mesh_workload.OwnerFetcher,
	meshClient zephyr_discovery.MeshClient,
) mesh_workload.MeshWorkloadScanner {
	return &linkerdMeshWorkloadScanner{
		deploymentFetcher: ownerFetcher,
		meshClient:        meshClient,
	}
}

type linkerdMeshWorkloadScanner struct {
	deploymentFetcher mesh_workload.OwnerFetcher
	meshClient        zephyr_discovery.MeshClient
}

func (l *linkerdMeshWorkloadScanner) ScanPod(ctx context.Context, pod *k8s_core_types.Pod, clusterName string) (*zephyr_discovery.MeshWorkload, error) {
	if !l.isLinkerdPod(pod) {
		return nil, nil
	}
	deployment, err := l.deploymentFetcher.GetDeployment(ctx, pod)
	if err != nil {
		return nil, err
	}
	meshRef, err := l.getMeshResourceRef(ctx, clusterName)
	if err != nil {
		return nil, err
	}
	return &zephyr_discovery.MeshWorkload{
		ObjectMeta: k8s_meta_types.ObjectMeta{
			Name:      l.buildMeshWorkloadName(deployment.GetName(), deployment.GetNamespace(), pod.GetClusterName()),
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

// iterate through pod's containers and check for one with name containing "linkerd" and "proxy"
func (l *linkerdMeshWorkloadScanner) isLinkerdPod(pod *k8s_core_types.Pod) bool {
	for _, container := range pod.Spec.Containers {
		if container.Name == "linkerd-proxy" {
			return true
		}
	}
	return false
}

func (l *linkerdMeshWorkloadScanner) getMeshResourceRef(ctx context.Context, clusterName string) (*zephyr_core_types.ResourceRef, error) {
	meshList, err := l.meshClient.ListMesh(ctx)
	if err != nil {
		return nil, err
	}
	for _, mesh := range meshList.Items {
		// Assume single tenancy for clusters with Linkerd mesh
		if mesh.Spec.GetCluster().GetName() == clusterName {
			return &zephyr_core_types.ResourceRef{
				Name:      mesh.GetName(),
				Namespace: mesh.GetNamespace(),
				Cluster:   clusterName,
			}, nil
		}
	}
	return nil, nil
}

func (l *linkerdMeshWorkloadScanner) buildMeshWorkloadName(deploymentName string, namespace string, clusterName string) string {
	// TODO: https://github.com/solo-io/service-mesh-hub/issues/141
	return fmt.Sprintf("%s-%s-%s-%s", "linkerd", deploymentName, namespace, clusterName)
}
