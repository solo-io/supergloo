package linkerd

import (
	"context"
	"fmt"
	"strings"

	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/common/container-runtime"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/discovery/mesh-workload/k8s"
	k8s_core_types "k8s.io/api/core/v1"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	DiscoveryLabels = func() map[string]string {
		return map[string]string{
			kube.MESH_TYPE: strings.ToLower(smh_core_types.MeshType_LINKERD.String()),
		}
	}
)

// visible for testing
func NewLinkerdMeshWorkloadScanner(
	ownerFetcherFactory k8s.OwnerFetcherFactory,
	meshClient smh_discovery.MeshClient,
) k8s.MeshWorkloadScanner {
	return &linkerdMeshWorkloadScanner{
		ownerFetcherFactory: ownerFetcherFactory,
		meshClient:          meshClient,
	}
}

type linkerdMeshWorkloadScanner struct {
	ownerFetcherFactory k8s.OwnerFetcherFactory
	meshClient          smh_discovery.MeshClient
}

func (l *linkerdMeshWorkloadScanner) ScanPod(ctx context.Context, pod *k8s_core_types.Pod, clusterName string) (*smh_discovery.MeshWorkload, error) {
	ownerFetcher, err := l.ownerFetcherFactory(clusterName)
	if err != nil {
		return nil, err
	}
	if !l.isLinkerdPod(pod) {
		return nil, nil
	}
	deployment, err := ownerFetcher.GetDeployment(ctx, pod)
	if err != nil {
		return nil, err
	}
	meshRef, err := l.getMeshResourceRef(ctx, clusterName)
	if err != nil {
		return nil, err
	}
	return &smh_discovery.MeshWorkload{
		ObjectMeta: k8s_meta_types.ObjectMeta{
			Name:      l.buildMeshWorkloadName(deployment.GetName(), deployment.GetNamespace(), clusterName),
			Namespace: container_runtime.GetWriteNamespace(),
			Labels:    DiscoveryLabels(),
		},
		Spec: smh_discovery_types.MeshWorkloadSpec{
			KubeController: &smh_discovery_types.MeshWorkloadSpec_KubeController{
				KubeControllerRef: &smh_core_types.ResourceRef{
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

// iterate through pod's containers and check for one with name containing "linkerd" and "proxy"
func (l *linkerdMeshWorkloadScanner) isLinkerdPod(pod *k8s_core_types.Pod) bool {
	for _, container := range pod.Spec.Containers {
		if container.Name == "linkerd-proxy" {
			return true
		}
	}
	return false
}

func (l *linkerdMeshWorkloadScanner) getMeshResourceRef(ctx context.Context, clusterName string) (*smh_core_types.ResourceRef, error) {
	meshList, err := l.meshClient.ListMesh(ctx)
	if err != nil {
		return nil, err
	}
	for _, mesh := range meshList.Items {
		// Assume single tenancy for clusters with Linkerd mesh
		if mesh.Spec.GetLinkerd() != nil && mesh.Spec.GetCluster().GetName() == clusterName {
			return &smh_core_types.ResourceRef{
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
