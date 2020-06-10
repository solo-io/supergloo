package istio

import (
	"context"
	"fmt"
	"strings"

	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/container-runtime"
	"github.com/solo-io/service-mesh-hub/pkg/kube"
	"github.com/solo-io/service-mesh-hub/pkg/kube/metadata"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh-workload/k8s"
	k8s_core_types "k8s.io/api/core/v1"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	DiscoveryLabels = func(meshType smh_core_types.MeshType) map[string]string {
		return map[string]string{
			kube.MESH_TYPE: strings.ToLower(meshType.String()),
		}
	}
)

// visible for testing
func NewIstioMeshWorkloadScanner(
	ownerFetcher k8s.OwnerFetcher,
	meshClient smh_discovery.MeshClient,
	_ client.Client,
) k8s.MeshWorkloadScanner {
	return &istioMeshWorkloadScanner{
		ownerFetcher: ownerFetcher,
		meshClient:   meshClient,
	}
}

type istioMeshWorkloadScanner struct {
	ownerFetcher k8s.OwnerFetcher
	meshClient   smh_discovery.MeshClient
}

func (i *istioMeshWorkloadScanner) ScanPod(ctx context.Context, pod *k8s_core_types.Pod, clusterName string) (*smh_discovery.MeshWorkload, error) {
	if !i.isIstioPod(pod) {
		return nil, nil
	}
	deployment, err := i.ownerFetcher.GetDeployment(ctx, pod)
	if err != nil {
		return nil, err
	}
	meshType, meshRef, err := i.getMeshResourceRef(ctx, clusterName)
	if err != nil {
		return nil, err
	}
	return &smh_discovery.MeshWorkload{
		ObjectMeta: k8s_meta_types.ObjectMeta{
			Name:      i.buildMeshWorkloadName(deployment.GetName(), deployment.GetNamespace(), clusterName),
			Namespace: container_runtime.GetWriteNamespace(),
			Labels:    DiscoveryLabels(meshType),
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

// iterate through pod's containers and check for one with name containing "istio" and "proxy"
func (i *istioMeshWorkloadScanner) isIstioPod(pod *k8s_core_types.Pod) bool {
	for _, container := range pod.Spec.Containers {
		if strings.Contains(container.Image, "istio") && strings.Contains(container.Image, "proxy") {
			return true
		}
	}
	return false
}

func (i *istioMeshWorkloadScanner) getMeshResourceRef(ctx context.Context, clusterName string) (smh_core_types.MeshType, *smh_core_types.ResourceRef, error) {
	meshList, err := i.meshClient.ListMesh(ctx)
	if err != nil {
		return 0, nil, err
	}
	for _, mesh := range meshList.Items {
		meshType, err := metadata.MeshToMeshType(&mesh)
		if err != nil {
			return 0, nil, err
		}

		isIstio5Or6 := meshType == smh_core_types.MeshType_ISTIO1_5 || meshType == smh_core_types.MeshType_ISTIO1_6

		// Assume single tenancy for clusters with Istio mesh
		if isIstio5Or6 && mesh.Spec.GetCluster().GetName() == clusterName {
			return meshType, &smh_core_types.ResourceRef{
				Name:      mesh.GetName(),
				Namespace: mesh.GetNamespace(),
				Cluster:   clusterName,
			}, nil
		}
	}
	return 0, nil, nil
}

func (i *istioMeshWorkloadScanner) buildMeshWorkloadName(deploymentName string, namespace string, clusterName string) string {
	// TODO: https://github.com/solo-io/service-mesh-hub/issues/141
	return fmt.Sprintf("%s-%s-%s-%s", "istio", deploymentName, namespace, clusterName)
}
