package k8s

import (
	"context"
	"sync"

	"github.com/solo-io/go-utils/contextutils"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_controller "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/controller"
	k8s_core "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1"
	k8s_core_controller "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1/controller"
	"github.com/solo-io/service-mesh-hub/pkg/clients"
	"github.com/solo-io/service-mesh-hub/pkg/enum_conversion"
	"github.com/solo-io/service-mesh-hub/pkg/logging"
	"github.com/solo-io/service-mesh-hub/services/common/constants"
	k8s_tenancy "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/cluster-tenancy/k8s"
	k8s_core_types "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	MeshWorkloadProcessingError = "Error processing deployment for mesh workload discovery"
)

func NewMeshWorkloadFinder(
	ctx context.Context,
	clusterName string,
	localMeshClient zephyr_discovery.MeshClient,
	localMeshWorkloadClient zephyr_discovery.MeshWorkloadClient,
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
		discoveredMeshTypes:                sets.NewInt32(),
	}
}

type meshWorkloadFinder struct {
	clusterName                        string
	ctx                                context.Context
	meshWorkloadScannerImplementations MeshWorkloadScannerImplementations
	localMeshClient                    zephyr_discovery.MeshClient
	localMeshWorkloadClient            zephyr_discovery.MeshWorkloadClient
	podClient                          k8s_core.PodClient

	// stateful record of the meshes we have discovered on this cluster.
	// as meshes get discovered, they will get added here and kick off discovery on pods again
	discoveredMeshTypes      sets.Int32
	discoveredMeshTypesMutex sync.RWMutex
}

func (m *meshWorkloadFinder) StartDiscovery(
	podEventWatcher k8s_core_controller.PodEventWatcher,
	meshEventWatcher zephyr_discovery_controller.MeshEventWatcher,
) error {
	// ensure the existing state in the cluster is accurate before starting to handle events
	err := m.reconcileExistingState()
	if err != nil {
		return err
	}

	err = podEventWatcher.AddEventHandler(
		m.ctx,
		&k8s_core_controller.PodEventHandlerFuncs{
			OnCreate: func(obj *k8s_core_types.Pod) error {
				logger := logging.BuildEventLogger(m.ctx, logging.CreateEvent, obj)
				logger.Debugf("Handling create for pod %s.%s", obj.GetName(), obj.GetNamespace())
				err = m.reconcileExistingState()
				if err != nil {
					logger.Errorf("%+v", err)
				}
				return err
			},
			OnUpdate: func(old, new *k8s_core_types.Pod) error {
				logger := logging.BuildEventLogger(m.ctx, logging.UpdateEvent, new)
				logger.Debugf("Handling create for pod %s.%s", new.GetName(), new.GetNamespace())
				err = m.reconcileExistingState()
				if err != nil {
					logger.Errorf("%+v", err)
				}
				return err
			},
			OnDelete: func(obj *k8s_core_types.Pod) error {
				logger := logging.BuildEventLogger(m.ctx, logging.DeleteEvent, obj)
				logger.Debugf("Handling create for pod %s.%s", obj.GetName(), obj.GetNamespace())
				err = m.reconcileExistingState()
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
		&zephyr_discovery_controller.MeshEventHandlerFuncs{
			OnCreate: func(obj *zephyr_discovery.Mesh) error {
				logger := logging.BuildEventLogger(m.ctx, logging.CreateEvent, obj)
				logger.Debugf("mesh create event for %s", obj.Name)
				err = m.reconcileExistingState()
				if err != nil {
					logger.Errorf("%+v", err)
				}
				return err
			},
			OnUpdate: func(old, new *zephyr_discovery.Mesh) error {
				logger := logging.BuildEventLogger(m.ctx, logging.UpdateEvent, new)
				logger.Debugf("mesh create event for %s", new.Name)
				err = m.reconcileExistingState()
				if err != nil {
					logger.Errorf("%+v", err)
				}
				return err
			},
			OnDelete: func(obj *zephyr_discovery.Mesh) error {
				logger := logging.BuildEventLogger(m.ctx, logging.DeleteEvent, obj)
				logger.Debugf("mesh create event for %s", obj.Name)
				err = m.reconcileExistingState()
				if err != nil {
					logger.Errorf("%+v", err)
				}
				return err
			},
		},
	)
	return err
}

func (m *meshWorkloadFinder) reconcileExistingState() error {
	inThisCluster := client.MatchingLabels{
		constants.COMPUTE_TARGET: m.clusterName,
	}

	existingMeshWorkloads, err := m.localMeshWorkloadClient.ListMeshWorkload(m.ctx, inThisCluster)
	if err != nil {
		return err
	}

	// nothing discovered yet, bail out
	if len(existingMeshWorkloads.Items) == 0 {
		return nil
	}

	podsInCluster, err := m.podClient.ListPod(m.ctx)
	if err != nil {
		return err
	}

	discoveredMeshTypes, err := m.getDiscoveredMeshTypes(m.ctx)
	if err != nil {
		return err
	}

	rediscoveredWorkloadsByName := map[string]*zephyr_discovery.MeshWorkload{}
	for _, podIter := range podsInCluster.Items {
		pod := podIter

		discoveredWorkload, err := m.discoverMeshWorkload(&pod, discoveredMeshTypes)
		if err != nil {
			return err
		} else if discoveredWorkload == nil {
			continue
		}

		rediscoveredWorkloadsByName[discoveredWorkload.GetName()] = discoveredWorkload
	}

	// for each mesh that we have rediscovered, ensure that we can match it to a mesh that currently exists in the cluster
	for _, meshWorkloadIter := range existingMeshWorkloads.Items {
		existingMeshWorkload := meshWorkloadIter

		discoveredWorkload, notDeleted := rediscoveredWorkloadsByName[existingMeshWorkload.GetName()]

		// if we both 1. rediscovered the workload from a pod, and 2. its spec is still the same,
		// then do nothing since we have not missed any event
		if notDeleted && existingMeshWorkload.Spec.Equal(discoveredWorkload.Spec) {
			continue
		}

		if discoveredWorkload != nil && !existingMeshWorkload.Spec.Equal(discoveredWorkload.Spec) {

			// we missed an update event, make sure the state of the cluster matches what we discovered from the pod
			err := m.localMeshWorkloadClient.UpsertMeshWorkloadSpec(m.ctx, discoveredWorkload)
			if err != nil {
				return err
			}
		} else {
			// else the exiting mesh workload was not re-discovered, so we missed a delete event- delete it
			err := m.localMeshWorkloadClient.DeleteMeshWorkload(m.ctx, clients.ObjectMetaToObjectKey(existingMeshWorkload.ObjectMeta))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *meshWorkloadFinder) handleMeshDelete(deletedMesh *zephyr_discovery.Mesh) error {
	logger := logging.BuildEventLogger(m.ctx, logging.DeleteEvent, deletedMesh)

	logger.Debugf("Handling delete for mesh %s.%s", deletedMesh.GetName(), deletedMesh.GetNamespace())

	// ignore meshes that are not running on this cluster
	if !k8s_tenancy.ClusterHostsMesh(m.clusterName, deletedMesh) {
		return nil
	}

	meshType, err := enum_conversion.MeshToMeshType(deletedMesh)
	if err != nil {
		logger.Errorf("%+v", err)
		return nil
	}

	// un-register the cluster from our internal representation so that later pod events in this cluster don't get re-discovered
	m.discoveredMeshTypesMutex.Lock()
	m.discoveredMeshTypes.Delete(int32(meshType))
	m.discoveredMeshTypesMutex.Unlock()

	allMeshWorkloads, err := m.localMeshWorkloadClient.ListMeshWorkload(m.ctx)
	if err != nil {
		logger.Errorf("%+v", err)
		return nil
	}

	for _, meshWorkloadIter := range allMeshWorkloads.Items {
		meshWorkload := meshWorkloadIter

		meshForWorkload := clients.ResourceRefToObjectMeta(meshWorkload.Spec.GetMesh())
		if clients.SameObject(meshForWorkload, deletedMesh.ObjectMeta) {
			err = m.localMeshWorkloadClient.DeleteMeshWorkload(m.ctx, clients.ObjectMetaToObjectKey(meshWorkload.ObjectMeta))
			if err != nil {
				logger.Errorf("%+v", err)
				return nil
			}
		}
	}

	return nil
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
		meshType, err := enum_conversion.MeshToMeshType(&mesh)
		if err != nil {
			return nil, err
		}
		discoveredMeshTypes.Insert(int32(meshType))
	}
	return discoveredMeshTypes, nil
}

func (m *meshWorkloadFinder) attachGeneralDiscoveryLabels(meshWorkload *zephyr_discovery.MeshWorkload) {
	if meshWorkload.Labels == nil {
		meshWorkload.Labels = map[string]string{}
	}
	meshWorkload.Labels[constants.DISCOVERED_BY] = constants.MESH_WORKLOAD_DISCOVERY
	meshWorkload.Labels[constants.COMPUTE_TARGET] = m.clusterName
	meshWorkload.Labels[constants.KUBE_CONTROLLER_NAME] = meshWorkload.Spec.GetKubeController().GetKubeControllerRef().GetName()
	meshWorkload.Labels[constants.KUBE_CONTROLLER_NAMESPACE] = meshWorkload.Spec.GetKubeController().GetKubeControllerRef().GetNamespace()
}

func (m *meshWorkloadFinder) discoverMeshWorkload(pod *k8s_core_types.Pod, discoveredMeshTypes sets.Int32) (*zephyr_discovery.MeshWorkload, error) {
	logger := contextutils.LoggerFrom(m.ctx)
	var discoveredMeshWorkload *zephyr_discovery.MeshWorkload
	var err error

	for _, discoveredMeshType := range discoveredMeshTypes.List() {
		meshWorkloadScanner, ok := m.meshWorkloadScannerImplementations[zephyr_core_types.MeshType(discoveredMeshType)]
		if !ok {
			logger.Warnf("No MeshWorkloadScanner found for mesh type: %s", zephyr_core_types.MeshType(discoveredMeshType).String())
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
