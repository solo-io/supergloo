package k8s

import (
	"context"

	k8s_core_clients "github.com/solo-io/external-apis/pkg/api/k8s/core/v1"
	"github.com/solo-io/go-utils/contextutils"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery_sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/sets"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/metadata"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/selection"
	k8s_tenancy "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/discovery/cluster-tenancy/k8s"
	k8s_core_types "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewMeshWorkloadDiscovery(
	meshClient smh_discovery.MeshClient,
	meshWorkloadClient smh_discovery.MeshWorkloadClient,
	meshWorkloadScannerImplementations MeshWorkloadScannerImplementations,
	multiclusterClientset k8s_core_clients.MulticlusterClientset,
) MeshWorkloadDiscovery {
	return &meshWorkloadDiscovery{
		meshWorkloadScannerImplementations: meshWorkloadScannerImplementations,
		meshClient:                         meshClient,
		meshWorkloadClient:                 meshWorkloadClient,
		multiclusterClientset:              multiclusterClientset,
	}
}

type meshWorkloadDiscovery struct {
	meshWorkloadScannerImplementations MeshWorkloadScannerImplementations
	meshClient                         smh_discovery.MeshClient
	meshWorkloadClient                 smh_discovery.MeshWorkloadClient
	multiclusterClientset              k8s_core_clients.MulticlusterClientset
}

func (m *meshWorkloadDiscovery) DiscoverMeshWorkloads(
	ctx context.Context,
	clusterName string,
) error {
	discoveredMeshTypes, err := m.getDiscoveredMeshTypes(ctx, clusterName)
	if err != nil {
		return err
	}
	existingWorkloads, err := m.getExistingWorkloads(ctx, clusterName)
	if err != nil {
		return err
	}
	discoveredWorkloads, err := m.discoverAllWorkloads(ctx, clusterName, discoveredMeshTypes)
	if err != nil {
		return err
	}
	existingWorkloadMap := existingWorkloads.Map()
	// For each workload that is discovered, create if it doesn't exist or update it if the spec has changed.
	for discoveredWorkloadKey, discoveredWorkload := range discoveredWorkloads.Map() {
		existingWorkload, ok := existingWorkloadMap[discoveredWorkloadKey]
		if !ok || !existingWorkload.Spec.Equal(discoveredWorkload.Spec) {
			err = m.meshWorkloadClient.UpsertMeshWorkload(ctx, discoveredWorkload)
			if err != nil {
				return err
			}
		}
	}
	// Delete MeshWorkloads that no longer exist.
	for _, existingWorkload := range existingWorkloads.Difference(discoveredWorkloads).List() {
		err = m.meshWorkloadClient.DeleteMeshWorkload(ctx, selection.ObjectMetaToObjectKey(existingWorkload.ObjectMeta))
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *meshWorkloadDiscovery) getExistingWorkloads(
	ctx context.Context,
	clusterName string,
) (smh_discovery_sets.MeshWorkloadSet, error) {
	inThisCluster := client.MatchingLabels{
		kube.COMPUTE_TARGET: clusterName,
	}
	meshWorkloadList, err := m.meshWorkloadClient.ListMeshWorkload(ctx, inThisCluster)
	if err != nil {
		return nil, err
	}
	workloads := smh_discovery_sets.NewMeshWorkloadSet()
	for _, meshWorkload := range meshWorkloadList.Items {
		meshWorkload := meshWorkload
		workloads.Insert(&meshWorkload)
	}
	return workloads, nil
}

func (m *meshWorkloadDiscovery) discoverAllWorkloads(
	ctx context.Context,
	clusterName string,
	discoveredMeshTypes sets.Int32,
) (smh_discovery_sets.MeshWorkloadSet, error) {
	clientset, err := m.multiclusterClientset.Cluster(clusterName)
	podClient := clientset.Pods()
	podList, err := podClient.ListPod(ctx)
	if err != nil {
		return nil, err
	}
	workloads := smh_discovery_sets.NewMeshWorkloadSet()
	for _, pod := range podList.Items {
		pod := pod
		discoveredWorkload, err := m.discoverMeshWorkload(ctx, clusterName, &pod, discoveredMeshTypes)
		if err != nil {
			contextutils.LoggerFrom(ctx).Warnf("Error scanning pod %s.%s: %+v", pod.GetName(), pod.GetNamespace(), err)
			continue
		} else if discoveredWorkload == nil {
			continue
		}
		workloads.Insert(discoveredWorkload)
	}
	return workloads, nil
}

func (m *meshWorkloadDiscovery) getDiscoveredMeshTypes(
	ctx context.Context,
	clusterName string,
) (sets.Int32, error) {
	meshList, err := m.meshClient.ListMesh(ctx)
	if err != nil {
		return nil, err
	}
	discoveredMeshTypes := sets.Int32{}
	for _, mesh := range meshList.Items {
		mesh := mesh
		// ensure we are only watching for meshes discovered on this same cluster
		// otherwise we can hit a race where:
		//  - Istio is discovered on cluster A
		//  - that mesh is recorded here
		//  - we start discovering workloads on cluster B using the Istio mesh workload discovery
		//  - but we haven't yet discovered Istio on this cluster
		if !k8s_tenancy.ClusterHostsMesh(clusterName, &mesh) {
			continue
		}
		meshType, err := metadata.MeshToMeshType(&mesh)
		if err != nil {
			return nil, err
		}
		discoveredMeshTypes.Insert(int32(meshType))
	}
	return discoveredMeshTypes, nil
}

func (m *meshWorkloadDiscovery) discoverMeshWorkload(
	ctx context.Context,
	clusterName string,
	pod *k8s_core_types.Pod,
	discoveredMeshTypes sets.Int32,
) (*smh_discovery.MeshWorkload, error) {
	logger := contextutils.LoggerFrom(ctx)
	var discoveredMeshWorkload *smh_discovery.MeshWorkload
	var err error

	for _, discoveredMeshType := range discoveredMeshTypes.List() {
		meshWorkloadScanner, ok := m.meshWorkloadScannerImplementations[smh_core_types.MeshType(discoveredMeshType)]
		if !ok {
			logger.Warnf("No MeshWorkloadScanner found for mesh type: %s", smh_core_types.MeshType(discoveredMeshType).String())
			continue
		}
		discoveredMeshWorkload, err = meshWorkloadScanner.ScanPod(ctx, pod, clusterName)
		if err != nil {
			return nil, err
		}
		if discoveredMeshWorkload != nil {
			break
		}
	}
	// the mesh workload needs to have our standard discovery labels attached to it, like cluster name, etc
	if discoveredMeshWorkload != nil {
		m.attachGeneralDiscoveryLabels(clusterName, discoveredMeshWorkload)
	}
	return discoveredMeshWorkload, nil
}

func (m *meshWorkloadDiscovery) attachGeneralDiscoveryLabels(
	clusterName string,
	meshWorkload *smh_discovery.MeshWorkload,
) {
	if meshWorkload.Labels == nil {
		meshWorkload.Labels = map[string]string{}
	}
	meshWorkload.Labels[kube.DISCOVERED_BY] = kube.MESH_WORKLOAD_DISCOVERY
	meshWorkload.Labels[kube.COMPUTE_TARGET] = clusterName
	meshWorkload.Labels[kube.KUBE_CONTROLLER_NAME] = meshWorkload.Spec.GetKubeController().GetKubeControllerRef().GetName()
	meshWorkload.Labels[kube.KUBE_CONTROLLER_NAMESPACE] = meshWorkload.Spec.GetKubeController().GetKubeControllerRef().GetNamespace()
}
