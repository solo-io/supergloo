package mesh_workload

import (
	"context"
	"sync"

	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/contextutils"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_controller "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/controller"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	k8s_core "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1"
	k8s_core_controller "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1/controller"
	"github.com/solo-io/service-mesh-hub/pkg/clients"
	"github.com/solo-io/service-mesh-hub/pkg/enum_conversion"
	"github.com/solo-io/service-mesh-hub/pkg/logging"
	"github.com/solo-io/service-mesh-hub/services/common/constants"
	"go.uber.org/zap"
	k8s_core_types "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	MeshWorkloadProcessingError    = "Error processing deployment for mesh workload discovery"
	MeshWorkloadProcessingNonFatal = "Non-fatal error occurred while scanning for mesh workloads"
)

var (
	FailedToReprocessPods = func(clusterName string, err error) error {
		return eris.Wrapf(err, "Failed to re-process pods in cluster %s for mesh workload discovery", clusterName)
	}
	FailedToComputeMeshRef = func(workloadName, workloadNamespace, clusterName string) error {
		return eris.Errorf("Failed to compute the owner mesh for mesh workload %s.%s in cluster %s", workloadName, workloadNamespace, clusterName)
	}
)

func NewMeshWorkloadFinder(
	ctx context.Context,
	clusterName string,
	localMeshWorkloadClient zephyr_discovery.MeshWorkloadClient,
	localMeshClient zephyr_discovery.MeshClient,
	meshWorkloadScannerImplementations MeshWorkloadScannerImplementations,
	podClient k8s_core.PodClient,
) MeshWorkloadFinder {

	return &meshWorkloadFinder{
		ctx:                                ctx,
		clusterName:                        clusterName,
		meshWorkloadScannerImplementations: meshWorkloadScannerImplementations,
		localMeshWorkloadClient:            localMeshWorkloadClient,
		localMeshClient:                    localMeshClient,
		podClient:                          podClient,
		discoveredMeshTypes:                sets.NewInt32(),
	}
}

type meshWorkloadFinder struct {
	clusterName                        string
	ctx                                context.Context
	meshWorkloadScannerImplementations MeshWorkloadScannerImplementations
	localMeshWorkloadClient            zephyr_discovery.MeshWorkloadClient
	localMeshClient                    zephyr_discovery.MeshClient
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
			OnCreate: m.handlePodCreate,
			OnUpdate: m.handlePodUpdate,
			OnDelete: m.handlePodDelete,
		},
	)
	if err != nil {
		return err
	}

	err = meshEventWatcher.AddEventHandler(
		m.ctx,
		&zephyr_discovery_controller.MeshEventHandlerFuncs{
			OnCreate: m.handleMeshCreate,
			OnDelete: m.handleMeshDelete,
		},
	)
	return err
}

func (m *meshWorkloadFinder) reconcileExistingState() error {
	inThisCluster := client.MatchingLabels{
		constants.CLUSTER: m.clusterName,
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

	meshes, err := m.localMeshClient.ListMesh(m.ctx, inThisCluster)
	if err != nil {
		return err
	}

	// manually run mesh create events for all meshes we already have recorded
	// this ensures that this component's stateful map of mesh types matches what it should
	for _, meshIter := range meshes.Items {
		mesh := meshIter

		err := m.handleMeshCreate(&mesh)
		if err != nil {
			return err
		}
	}

	rediscoveredWorkloadsByName := map[string]*zephyr_discovery.MeshWorkload{}
	for _, podIter := range podsInCluster.Items {
		pod := podIter

		discoveredWorkload, err := m.discoverMeshWorkload(&pod)
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
			existingMeshWorkload.Spec = discoveredWorkload.Spec
			err := m.localMeshWorkloadClient.UpdateMeshWorkload(m.ctx, &existingMeshWorkload)
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

func (m *meshWorkloadFinder) handleMeshDelete(meshBeingDeleted *zephyr_discovery.Mesh) error {
	logger := logging.BuildEventLogger(m.ctx, logging.DeleteEvent, meshBeingDeleted)

	logger.Debugf("Handling delete for mesh %s.%s", meshBeingDeleted.GetName(), meshBeingDeleted.GetNamespace())

	// ignore meshes that are not running on this cluster
	if meshBeingDeleted.Spec.GetCluster().GetName() != m.clusterName {
		return nil
	}

	meshType, err := enum_conversion.MeshToMeshType(meshBeingDeleted)
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
		if clients.SameObject(meshForWorkload, meshBeingDeleted.ObjectMeta) {
			err = m.localMeshWorkloadClient.DeleteMeshWorkload(m.ctx, clients.ObjectMetaToObjectKey(meshWorkload.ObjectMeta))
			if err != nil {
				logger.Errorf("%+v", err)
				return nil
			}
		}
	}

	return nil
}

func (m *meshWorkloadFinder) handleMeshCreate(mesh *zephyr_discovery.Mesh) error {
	logger := contextutils.LoggerFrom(m.ctx)
	logger.Debugf("mesh create event for %s", mesh.Name)

	// ensure we are only watching for meshes discovered on this same cluster
	// otherwise we can hit a race where:
	//  - Istio is discovered on cluster A
	//  - that mesh is recorded here
	//  - we start discovering workloads on cluster B using the Istio mesh workload discovery
	//  - but we haven't yet discovered Istio on this cluster
	if mesh.Spec.GetCluster().GetName() != m.clusterName {
		return nil
	}

	meshType, err := enum_conversion.MeshToMeshType(mesh)
	if err != nil {
		logger.Errorf("%+v", err)
		return nil
	}

	// first, register in our stateful representation that we have seen this mesh
	// this allows us to use the mesh workload scanner implementation later on
	m.discoveredMeshTypesMutex.Lock()
	m.discoveredMeshTypes.Insert(int32(meshType))
	m.discoveredMeshTypesMutex.Unlock()

	allPods, err := m.podClient.ListPod(m.ctx)
	if err != nil {
		contextutils.LoggerFrom(m.ctx).Errorf("Error loading all pods from mesh create event %s", mesh.Name)
		return FailedToReprocessPods(m.clusterName, err)
	}

	// now that we have discovered a new mesh on this cluster, kick off a re-process of all existing pods
	for _, podIter := range allPods.Items {
		pod := podIter
		contextutils.LoggerFrom(m.ctx).Debugf("mesh create event for %s - processing pod %s", mesh.Name, pod.Name)
		err := m.handlePodCreate(&pod)
		if err != nil {
			contextutils.LoggerFrom(m.ctx).Errorf("Error reprocessing all pods from mesh create event %s", mesh.Name)
			return FailedToReprocessPods(m.clusterName, err)
		}
	}

	return nil
}

func (m *meshWorkloadFinder) handlePodDelete(pod *k8s_core_types.Pod) error {
	logger := logging.BuildEventLogger(m.ctx, logging.DeleteEvent, pod)

	logger.Debugf("Handling delete for pod %s.%s", pod.GetName(), pod.GetNamespace())

	discoveredMeshWorkload, err := m.discoverMeshWorkload(pod)
	if err != nil {
		logger.Errorf("Error while handling pod delete: %+v", err)
		return nil
	}

	if discoveredMeshWorkload == nil {
		return nil
	}

	err = m.localMeshWorkloadClient.DeleteMeshWorkload(m.ctx, clients.ObjectMetaToObjectKey(discoveredMeshWorkload.ObjectMeta))
	if err != nil {
		logger.Errorf("Could not delete mesh workload: %+v", err)
		return nil
	}

	return nil
}

func (m *meshWorkloadFinder) handlePodCreate(pod *k8s_core_types.Pod) error {
	pod.SetClusterName(m.clusterName)
	logger := logging.BuildEventLogger(m.ctx, logging.CreateEvent, pod)

	logger.Debugf("Handling create for pod %s.%s", pod.GetName(), pod.GetNamespace())
	discoveredMeshWorkload, err := m.discoverMeshWorkload(pod)

	logger.Debugf("pod %s.%s discovered: %t", pod.GetName(), pod.GetNamespace(), discoveredMeshWorkload != nil)

	if err != nil && discoveredMeshWorkload == nil {
		logger.Errorw(MeshWorkloadProcessingError, zap.Error(err))
		return err
	} else if err != nil && discoveredMeshWorkload != nil {
		logger.Warnw(MeshWorkloadProcessingNonFatal, zap.Error(err))
	} else if discoveredMeshWorkload == nil {
		logger.Debugf("MeshWorkload not found for pod %s/%s", pod.Namespace, pod.Name)
		return nil
	}
	err = m.createOrUpdateWorkload(discoveredMeshWorkload)
	logger.Debugf("pod %s.%s resulted in error after upsert: %t", pod.GetName(), pod.GetNamespace(), err != nil)
	if err != nil {
		logger.Errorf("Error: %+v", err)
	}
	return err
}

func (m *meshWorkloadFinder) handlePodUpdate(old, new *k8s_core_types.Pod) error {
	old.SetClusterName(m.clusterName)
	new.SetClusterName(m.clusterName)

	logger := logging.BuildEventLogger(m.ctx, logging.UpdateEvent, new)

	oldMeshWorkload, err := m.discoverMeshWorkload(old)
	if err != nil && oldMeshWorkload == nil {
		logger.Errorw(MeshWorkloadProcessingError, zap.Error(err))
		return err
	} else if err != nil && oldMeshWorkload != nil {
		logger.Warnw(MeshWorkloadProcessingNonFatal, zap.Error(err))
	}
	newMeshWorkload, err := m.discoverMeshWorkload(new)
	if err != nil && newMeshWorkload == nil {
		logger.Errorw(MeshWorkloadProcessingError, zap.Error(err))
		return err
	} else if err != nil && newMeshWorkload != nil {
		logger.Warnw(MeshWorkloadProcessingNonFatal, zap.Error(err))
	}

	if oldMeshWorkload == nil && newMeshWorkload == nil {
		// irrelevant pod, ignore
		return nil
	} else if oldMeshWorkload == nil && newMeshWorkload != nil {
		// existing pod is now mesh-injected
		return m.createOrUpdateWorkload(newMeshWorkload)
	} else if oldMeshWorkload != nil && newMeshWorkload == nil {
		// existing pod is no longer mesh-injected
		// TODO: delete
		return nil
	} else {
		// old and new MeshWorkloads equivalent
		if oldMeshWorkload.Spec.Equal(newMeshWorkload.Spec) {
			return nil
		} else {
			return m.localMeshWorkloadClient.UpdateMeshWorkload(m.ctx, newMeshWorkload)
		}
	}
}

func (m *meshWorkloadFinder) DeletePod(pod *k8s_core_types.Pod) error {
	logger := logging.BuildEventLogger(m.ctx, logging.DeleteEvent, pod)
	logger.Error("Deletion of MeshWorkloads is currently not supported")
	return nil
}

func (m *meshWorkloadFinder) GenericPod(pod *k8s_core_types.Pod) error {
	logger := logging.BuildEventLogger(m.ctx, logging.GenericEvent, pod)
	logger.Error("MeshWorkload generic events are not currently supported")
	return nil
}

func (m *meshWorkloadFinder) attachGeneralDiscoveryLabels(controllerRef *zephyr_core_types.ResourceRef, meshWorkload *zephyr_discovery.MeshWorkload) {
	if meshWorkload.Labels == nil {
		meshWorkload.Labels = map[string]string{}
	}
	meshWorkload.Labels[constants.DISCOVERED_BY] = constants.MESH_WORKLOAD_DISCOVERY
	meshWorkload.Labels[constants.CLUSTER] = m.clusterName
	meshWorkload.Labels[constants.KUBE_CONTROLLER_NAME] = controllerRef.GetName()
	meshWorkload.Labels[constants.KUBE_CONTROLLER_NAMESPACE] = controllerRef.GetNamespace()
}

func (m *meshWorkloadFinder) discoverMeshWorkload(pod *k8s_core_types.Pod) (*zephyr_discovery.MeshWorkload, error) {
	var (
		discoveredMeshWorkload *zephyr_discovery.MeshWorkload
		controllerRef          *zephyr_core_types.ResourceRef
	)

	// Only run mesh workload scanner implementations for meshes that are known to have been discovered.
	// This prevents a race condition: If a mesh and its injected pods exist already when mesh-discovery comes online,
	// then the pods can be discovered first and will end up having a nil mesh reference on them.
	m.discoveredMeshTypesMutex.RLock()
	discoveredMeshTypes := m.discoveredMeshTypes.List()
	m.discoveredMeshTypesMutex.RUnlock()

	for _, discoveredMeshType := range discoveredMeshTypes {
		meshWorkloadScanner, ok := m.meshWorkloadScannerImplementations[zephyr_core_types.MeshType(discoveredMeshType)]
		if !ok {
			return nil, eris.Errorf("Missing mesh workload scanner implementation for mesh type %d", discoveredMeshType)
		}

		// some of the workload scanners depend on this being set in memory
		// :(
		pod.ClusterName = m.clusterName
		discoveredControllerRef, discoveredMeshWorkloadObjectMeta, err := meshWorkloadScanner.ScanPod(m.ctx, pod)
		if err != nil {
			return nil, err
		}
		if discoveredControllerRef != nil {
			controllerRef = discoveredControllerRef

			meshRef, err := m.createMeshResourceRef(m.ctx)
			if err != nil {
				return nil, err
			}
			if meshRef == nil {
				return nil, FailedToComputeMeshRef(discoveredMeshWorkloadObjectMeta.Name, discoveredMeshWorkloadObjectMeta.Namespace, m.clusterName)
			}
			discoveredMeshWorkload = &zephyr_discovery.MeshWorkload{
				ObjectMeta: discoveredMeshWorkloadObjectMeta,
				Spec: zephyr_discovery_types.MeshWorkloadSpec{
					KubeController: &zephyr_discovery_types.MeshWorkloadSpec_KubeController{
						KubeControllerRef:  controllerRef,
						Labels:             pod.Labels,
						ServiceAccountName: pod.Spec.ServiceAccountName,
					},
					Mesh: meshRef,
				},
			}
			break
		}
	}

	// the mesh workload needs to have our standard discovery labels attached to it, like cluster name, etc
	if discoveredMeshWorkload != nil {
		m.attachGeneralDiscoveryLabels(controllerRef, discoveredMeshWorkload)
	}
	return discoveredMeshWorkload, nil
}

func (m *meshWorkloadFinder) createOrUpdateWorkload(discoveredWorkload *zephyr_discovery.MeshWorkload) error {
	objectKey, err := client.ObjectKeyFromObject(discoveredWorkload)
	if err != nil {
		return err
	}
	mw, err := m.localMeshWorkloadClient.GetMeshWorkload(m.ctx, objectKey)
	if errors.IsNotFound(err) {
		return m.localMeshWorkloadClient.CreateMeshWorkload(m.ctx, discoveredWorkload)
	} else if err != nil {
		return err
	} else if mw.Spec.Equal(discoveredWorkload.Spec) {
		return nil
	}
	// Need to do this, as we need metadata from previous object, (ResourceVersion), for update

	mw.Spec = discoveredWorkload.Spec
	mw.Labels = discoveredWorkload.Labels
	return m.localMeshWorkloadClient.UpdateMeshWorkload(m.ctx, mw)
}

func (m *meshWorkloadFinder) createMeshResourceRef(ctx context.Context) (*zephyr_core_types.ResourceRef, error) {
	meshList, err := m.localMeshClient.ListMesh(ctx, &client.ListOptions{})
	if err != nil {
		return nil, err
	}
	// assume at most one instance of Istio per cluster, thus it must be the Mesh for the MeshWorkload if it exists
	for _, mesh := range meshList.Items {
		if mesh.Spec.Cluster.Name == m.clusterName {
			return &zephyr_core_types.ResourceRef{
				Name:      mesh.Name,
				Namespace: mesh.Namespace,
				Cluster:   m.clusterName,
			}, nil
		}
	}
	return nil, nil
}
