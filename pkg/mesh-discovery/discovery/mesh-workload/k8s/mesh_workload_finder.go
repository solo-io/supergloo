package k8s

import (
	"context"

	k8s_core "github.com/solo-io/external-apis/pkg/api/k8s/core/v1"
	k8s_core_controller "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/controller"
	"github.com/solo-io/go-utils/contextutils"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery_controller "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/controller"
	smh_discovery_sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/sets"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/common/container-runtime"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/metadata"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/selection"
	k8s_tenancy "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/discovery/cluster-tenancy/k8s"
	k8s_core_types "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewMeshWorkloadFinder(
	ctx context.Context,
	clusterName string,
	localMeshClient smh_discovery.MeshClient,
	localMeshWorkloadClient smh_discovery.MeshWorkloadClient,
	meshWorkloadScannerImplementations MeshWorkloadScannerImplementations,
	podClient k8s_core.PodClient,
) MeshWorkloadFinder {

	return &meshWorkloadFinder{
		ctx:                                ctx,
		clusterName:                        clusterName,
		meshWorkloadScannerImplementations: meshWorkloadScannerImplementations,
		localMeshClient:                    localMeshClient,
		localMeshWorkloadClient:            localMeshWorkloadClient,
		podClient:                          podClient,
	}
}

type meshWorkloadFinder struct {
	clusterName                        string
	ctx                                context.Context
	meshWorkloadScannerImplementations MeshWorkloadScannerImplementations
	localMeshClient                    smh_discovery.MeshClient
	localMeshWorkloadClient            smh_discovery.MeshWorkloadClient
	podClient                          k8s_core.PodClient
}

func (m *meshWorkloadFinder) StartDiscovery(
	podEventWatcher k8s_core_controller.PodEventWatcher,
	meshEventWatcher smh_discovery_controller.MeshEventWatcher,
) error {
	// ensure the existing state in the cluster is accurate before starting to handle events
	err := m.reconcile()
	if err != nil {
		return err
	}

	err = podEventWatcher.AddEventHandler(
		m.ctx,
		&k8s_core_controller.PodEventHandlerFuncs{
			OnCreate: func(obj *k8s_core_types.Pod) error {
				logger := container_runtime.BuildEventLogger(m.ctx, container_runtime.CreateEvent, obj)
				logger.Debugf("Handling create for pod %s.%s", obj.GetName(), obj.GetNamespace())
				err = m.reconcile()
				if err != nil {
					logger.Errorf("%+v", err)
				}
				return err
			},
			OnUpdate: func(old, new *k8s_core_types.Pod) error {
				logger := container_runtime.BuildEventLogger(m.ctx, container_runtime.UpdateEvent, new)
				logger.Debugf("Handling update for pod %s.%s", new.GetName(), new.GetNamespace())
				err = m.reconcile()
				if err != nil {
					logger.Errorf("%+v", err)
				}
				return err
			},
			OnDelete: func(obj *k8s_core_types.Pod) error {
				logger := container_runtime.BuildEventLogger(m.ctx, container_runtime.DeleteEvent, obj)
				logger.Debugf("Handling delete for pod %s.%s", obj.GetName(), obj.GetNamespace())
				err = m.reconcile()
				if err != nil {
					logger.Errorf("%+v", err)
				}
				return err
			},
		},
	)
	if err != nil {
		return err
	}
	err = meshEventWatcher.AddEventHandler(
		m.ctx,
		&smh_discovery_controller.MeshEventHandlerFuncs{
			OnCreate: func(obj *smh_discovery.Mesh) error {
				logger := container_runtime.BuildEventLogger(m.ctx, container_runtime.CreateEvent, obj)
				logger.Debugf("mesh create event for %s.%s", obj.GetName(), obj.GetNamespace())
				err = m.reconcile()
				if err != nil {
					logger.Errorf("%+v", err)
				}
				return err
			},
			OnUpdate: func(old, new *smh_discovery.Mesh) error {
				logger := container_runtime.BuildEventLogger(m.ctx, container_runtime.UpdateEvent, new)
				logger.Debugf("mesh update event for %s.%s", new.GetName(), new.GetNamespace())
				err = m.reconcile()
				if err != nil {
					logger.Errorf("%+v", err)
				}
				return err
			},
			OnDelete: func(obj *smh_discovery.Mesh) error {
				logger := container_runtime.BuildEventLogger(m.ctx, container_runtime.DeleteEvent, obj)
				logger.Debugf("mesh delete event for %s.%s", obj.GetName(), obj.GetNamespace())
				err = m.reconcile()
				if err != nil {
					logger.Errorf("%+v", err)
				}
				return err
			},
		},
	)
	return err
}

func (m *meshWorkloadFinder) reconcile() error {
	discoveredMeshTypes, err := m.getDiscoveredMeshTypes(m.ctx)
	if err != nil {
		return err
	}
	existingWorkloads, err := m.getExistingWorkloads()
	if err != nil {
		return err
	}
	discoveredWorkloads, err := m.discoverAllWorkloads(discoveredMeshTypes)
	if err != nil {
		return err
	}
	existingWorkloadMap := existingWorkloads.Map()
	// For each workload that is discovered, create if it doesn't exist or update it if the spec has changed.
	for discoveredWorkloadKey, discoveredWorkload := range discoveredWorkloads.Map() {
		existingWorkload, ok := existingWorkloadMap[discoveredWorkloadKey]
		if !ok || !existingWorkload.Spec.Equal(discoveredWorkload.Spec) {
			err = m.localMeshWorkloadClient.UpsertMeshWorkload(m.ctx, discoveredWorkload)
			if err != nil {
				return err
			}
		}
	}
	// Delete MeshWorkloads that no longer exist.
	for _, existingWorkload := range existingWorkloads.Difference(discoveredWorkloads).List() {
		err = m.localMeshWorkloadClient.DeleteMeshWorkload(m.ctx, selection.ObjectMetaToObjectKey(existingWorkload.ObjectMeta))
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *meshWorkloadFinder) getExistingWorkloads() (smh_discovery_sets.MeshWorkloadSet, error) {
	inThisCluster := client.MatchingLabels{
		kube.COMPUTE_TARGET: m.clusterName,
	}
	meshWorkloadList, err := m.localMeshWorkloadClient.ListMeshWorkload(m.ctx, inThisCluster)
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

func (m *meshWorkloadFinder) discoverAllWorkloads(discoveredMeshTypes sets.Int32) (smh_discovery_sets.MeshWorkloadSet, error) {
	podList, err := m.podClient.ListPod(m.ctx)
	if err != nil {
		return nil, err
	}
	workloads := smh_discovery_sets.NewMeshWorkloadSet()
	for _, pod := range podList.Items {
		pod := pod
		discoveredWorkload, err := m.discoverMeshWorkload(&pod, discoveredMeshTypes)
		if err != nil {
			contextutils.LoggerFrom(m.ctx).Warnf("Error scanning pod %s.%s: %+v", pod.GetName(), pod.GetNamespace(), err)
			continue
		} else if discoveredWorkload == nil {
			continue
		}
		workloads.Insert(discoveredWorkload)
	}
	return workloads, nil
}

func (m *meshWorkloadFinder) getDiscoveredMeshTypes(ctx context.Context) (sets.Int32, error) {
	meshList, err := m.localMeshClient.ListMesh(ctx)
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
		if !k8s_tenancy.ClusterHostsMesh(m.clusterName, &mesh) {
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

func (m *meshWorkloadFinder) discoverMeshWorkload(pod *k8s_core_types.Pod, discoveredMeshTypes sets.Int32) (*smh_discovery.MeshWorkload, error) {
	logger := contextutils.LoggerFrom(m.ctx)
	var discoveredMeshWorkload *smh_discovery.MeshWorkload
	var err error

	for _, discoveredMeshType := range discoveredMeshTypes.List() {
		meshWorkloadScanner, ok := m.meshWorkloadScannerImplementations[smh_core_types.MeshType(discoveredMeshType)]
		if !ok {
			logger.Warnf("No MeshWorkloadScanner found for mesh type: %s", smh_core_types.MeshType(discoveredMeshType).String())
			continue
		}
		discoveredMeshWorkload, err = meshWorkloadScanner.ScanPod(m.ctx, pod, m.clusterName)
		if err != nil {
			return nil, err
		}
		if discoveredMeshWorkload != nil {
			break
		}
	}
	// the mesh workload needs to have our standard discovery labels attached to it, like cluster name, etc
	if discoveredMeshWorkload != nil {
		m.attachGeneralDiscoveryLabels(discoveredMeshWorkload)
	}
	return discoveredMeshWorkload, nil
}

func (m *meshWorkloadFinder) attachGeneralDiscoveryLabels(meshWorkload *smh_discovery.MeshWorkload) {
	if meshWorkload.Labels == nil {
		meshWorkload.Labels = map[string]string{}
	}
	meshWorkload.Labels[kube.DISCOVERED_BY] = kube.MESH_WORKLOAD_DISCOVERY
	meshWorkload.Labels[kube.COMPUTE_TARGET] = m.clusterName
	meshWorkload.Labels[kube.KUBE_CONTROLLER_NAME] = meshWorkload.Spec.GetKubeController().GetKubeControllerRef().GetName()
	meshWorkload.Labels[kube.KUBE_CONTROLLER_NAMESPACE] = meshWorkload.Spec.GetKubeController().GetKubeControllerRef().GetNamespace()
}
