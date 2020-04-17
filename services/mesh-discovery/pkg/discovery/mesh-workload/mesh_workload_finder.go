package mesh_workload

import (
	"context"
	"fmt"

	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/contextutils"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_controller "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/controller"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	k8s_core "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1"
	k8s_core_controller "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1/controller"
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
	discoveredMeshTypes sets.Int32
}

func (d *meshWorkloadFinder) StartDiscovery(
	podEventWatcher k8s_core_controller.PodEventWatcher,
	meshEventWatcher zephyr_discovery_controller.MeshEventWatcher,
) error {
	err := podEventWatcher.AddEventHandler(
		d.ctx,
		&k8s_core_controller.PodEventHandlerFuncs{
			OnCreate: d.handlePodCreate,
			OnUpdate: d.handlePodUpdate,
		},
	)
	if err != nil {
		return err
	}

	err = meshEventWatcher.AddEventHandler(
		d.ctx,
		&zephyr_discovery_controller.MeshEventHandlerFuncs{
			OnCreate: d.handleMeshCreate,
		},
	)
	return err
}

func (d *meshWorkloadFinder) handleMeshCreate(mesh *zephyr_discovery.Mesh) error {
	logger := contextutils.LoggerFrom(d.ctx)
	logger.Debugf("mesh create event for %s", mesh.Name)

	// first, register in our stateful representation that we have seen this mesh
	// this allows us to use the mesh workload scanner implementation later on
	switch mesh.Spec.GetMeshType().(type) {
	case *zephyr_discovery_types.MeshSpec_Istio:
		d.discoveredMeshTypes.Insert(int32(zephyr_core_types.MeshType_ISTIO))
	case *zephyr_discovery_types.MeshSpec_Linkerd:
		d.discoveredMeshTypes.Insert(int32(zephyr_core_types.MeshType_LINKERD))
	case *zephyr_discovery_types.MeshSpec_ConsulConnect:
		d.discoveredMeshTypes.Insert(int32(zephyr_core_types.MeshType_CONSUL))
	case *zephyr_discovery_types.MeshSpec_AwsAppMesh_:
		d.discoveredMeshTypes.Insert(int32(zephyr_core_types.MeshType_APPMESH))
	default:
		message := fmt.Sprintf("Unexpected error: unhandled mesh type in mesh workload discovery: %+v", mesh.Spec)
		logger.Error(message)
		return nil // don't want to requeue this event
	}

	allPods, err := d.podClient.ListPod(d.ctx)
	if err != nil {
		contextutils.LoggerFrom(d.ctx).Errorf("Error loading all pods from mesh create event %s", mesh.Name)
		return FailedToReprocessPods(d.clusterName, err)
	}

	// now that we have discovered a new mesh on this cluster, kick off a re-process of all existing pods
	for _, podIter := range allPods.Items {
		pod := podIter
		contextutils.LoggerFrom(d.ctx).Debugf("mesh create event for %s - processing pod %s", mesh.Name, pod.Name)
		err := d.handlePodCreate(&pod)
		if err != nil {
			contextutils.LoggerFrom(d.ctx).Errorf("Error reprocessing all pods from mesh create event %s", mesh.Name)
			return FailedToReprocessPods(d.clusterName, err)
		}
	}

	return nil
}

func (d *meshWorkloadFinder) handlePodCreate(pod *k8s_core_types.Pod) error {
	pod.SetClusterName(d.clusterName)
	logger := logging.BuildEventLogger(d.ctx, logging.CreateEvent, pod)

	logger.Debugf("Handling create for pod %s.%s", pod.GetName(), pod.GetNamespace())
	discoveredMeshWorkload, err := d.discoverMeshWorkload(pod)

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
	err = d.createOrUpdateWorkload(discoveredMeshWorkload)
	logger.Debugf("pod %s.%s resulted in error after upsert: %t", pod.GetName(), pod.GetNamespace(), err != nil)
	if err != nil {
		logger.Errorf("Error: %+v", err)
	}
	return err
}

func (d *meshWorkloadFinder) handlePodUpdate(old, new *k8s_core_types.Pod) error {
	old.SetClusterName(d.clusterName)
	new.SetClusterName(d.clusterName)
	logger := logging.BuildEventLogger(d.ctx, logging.UpdateEvent, new)
	oldMeshWorkload, err := d.discoverMeshWorkload(old)
	if err != nil && oldMeshWorkload == nil {
		logger.Errorw(MeshWorkloadProcessingError, zap.Error(err))
		return err
	} else if err != nil && oldMeshWorkload != nil {
		logger.Warnw(MeshWorkloadProcessingNonFatal, zap.Error(err))
	}
	newMeshWorkload, err := d.discoverMeshWorkload(new)
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
		return d.createOrUpdateWorkload(newMeshWorkload)
	} else if oldMeshWorkload != nil && newMeshWorkload == nil {
		// existing pod is no longer mesh-injected
		// TODO: delete
		return nil
	} else {
		// old and new MeshWorkloads equivalent
		if oldMeshWorkload.Spec.Equal(newMeshWorkload.Spec) {
			return nil
		} else {
			return d.localMeshWorkloadClient.UpdateMeshWorkload(d.ctx, newMeshWorkload)
		}
	}
}

func (d *meshWorkloadFinder) DeletePod(pod *k8s_core_types.Pod) error {
	logger := logging.BuildEventLogger(d.ctx, logging.DeleteEvent, pod)
	logger.Error("Deletion of MeshWorkloads is currently not supported")
	return nil
}

func (d *meshWorkloadFinder) GenericPod(pod *k8s_core_types.Pod) error {
	logger := logging.BuildEventLogger(d.ctx, logging.GenericEvent, pod)
	logger.Error("MeshWorkload generic events are not currently supported")
	return nil
}

func (d *meshWorkloadFinder) attachGeneralDiscoveryLabels(controllerRef *zephyr_core_types.ResourceRef, meshWorkload *zephyr_discovery.MeshWorkload) {
	if meshWorkload.Labels == nil {
		meshWorkload.Labels = map[string]string{}
	}
	meshWorkload.Labels[constants.DISCOVERED_BY] = constants.MESH_WORKLOAD_DISCOVERY
	meshWorkload.Labels[constants.CLUSTER] = d.clusterName
	meshWorkload.Labels[constants.KUBE_CONTROLLER_NAME] = controllerRef.GetName()
	meshWorkload.Labels[constants.KUBE_CONTROLLER_NAMESPACE] = controllerRef.GetNamespace()
}

func (d *meshWorkloadFinder) discoverMeshWorkload(pod *k8s_core_types.Pod) (*zephyr_discovery.MeshWorkload, error) {
	var (
		discoveredMeshWorkload *zephyr_discovery.MeshWorkload
		controllerRef          *zephyr_core_types.ResourceRef
	)

	// Only run mesh workload scanner implementations for meshes that are known to have been discovered.
	// This prevents a race condition: If a mesh and its injected pods exist already when mesh-discovery comes online,
	// then the pods can be discovered first and will end up having a nil mesh reference on them.
	for _, discoveredMeshType := range d.discoveredMeshTypes.List() {
		meshWorkloadScanner, ok := d.meshWorkloadScannerImplementations[zephyr_core_types.MeshType(discoveredMeshType)]
		if !ok {
			return nil, eris.Errorf("Missing mesh workload scanner implementation for mesh type %d", discoveredMeshType)
		}

		discoveredControllerRef, discoveredMeshWorkloadObjectMeta, err := meshWorkloadScanner.ScanPod(d.ctx, pod)
		if err != nil {
			return nil, err
		}
		if discoveredControllerRef != nil {
			controllerRef = discoveredControllerRef

			meshRef, err := d.createMeshResourceRef(d.ctx)
			if err != nil {
				return nil, err
			}
			if meshRef == nil {
				return nil, FailedToComputeMeshRef(discoveredMeshWorkloadObjectMeta.Name, discoveredMeshWorkloadObjectMeta.Namespace, d.clusterName)
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
		d.attachGeneralDiscoveryLabels(controllerRef, discoveredMeshWorkload)
	}
	return discoveredMeshWorkload, nil
}

func (d *meshWorkloadFinder) createOrUpdateWorkload(discoveredWorkload *zephyr_discovery.MeshWorkload) error {
	objectKey, err := client.ObjectKeyFromObject(discoveredWorkload)
	if err != nil {
		return err
	}
	mw, err := d.localMeshWorkloadClient.GetMeshWorkload(d.ctx, objectKey)
	if err != nil {
		if errors.IsNotFound(err) {
			return d.localMeshWorkloadClient.CreateMeshWorkload(d.ctx, discoveredWorkload)
		}
		return err
	}
	// Need to do this, as we need metadata from previous object, (ResourceVersion), for update

	mw.Spec = discoveredWorkload.Spec
	mw.Labels = discoveredWorkload.Labels
	return d.localMeshWorkloadClient.UpdateMeshWorkload(d.ctx, mw)
}

func (d *meshWorkloadFinder) createMeshResourceRef(ctx context.Context) (*zephyr_core_types.ResourceRef, error) {
	meshList, err := d.localMeshClient.ListMesh(ctx, &client.ListOptions{})
	if err != nil {
		return nil, err
	}
	// assume at most one instance of Istio per cluster, thus it must be the Mesh for the MeshWorkload if it exists
	for _, mesh := range meshList.Items {
		if mesh.Spec.Cluster.Name == d.clusterName {
			return &zephyr_core_types.ResourceRef{
				Name:      mesh.Name,
				Namespace: mesh.Namespace,
				Cluster:   d.clusterName,
			}, nil
		}
	}
	return nil, nil
}
