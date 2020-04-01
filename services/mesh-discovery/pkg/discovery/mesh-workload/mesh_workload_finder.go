package mesh_workload

import (
	"context"

	"github.com/hashicorp/go-multierror"
	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	discoveryv1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	discovery_types "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	zephyr_core "github.com/solo-io/mesh-projects/pkg/clients/zephyr/discovery"
	"github.com/solo-io/mesh-projects/pkg/logging"
	"github.com/solo-io/mesh-projects/services/common/cluster/core/v1/controller"
	"github.com/solo-io/mesh-projects/services/common/constants"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	MeshWorkloadProcessingError    = "Error processing deployment for mesh workload discovery"
	MeshWorkloadProcessingNonFatal = "Non-fatal error occurred while scanning for mesh workloads"
)

func NewMeshWorkloadFinder(
	ctx context.Context,
	clusterName string,
	localMeshWorkloadClient zephyr_core.MeshWorkloadClient,
	localMeshClient zephyr_core.MeshClient,
	meshWorkloadScanners []MeshWorkloadScanner,
) MeshWorkloadFinder {

	return &meshWorkloadFinder{
		clusterName:             clusterName,
		ctx:                     ctx,
		meshWorkloadScanners:    meshWorkloadScanners,
		localMeshWorkloadClient: localMeshWorkloadClient,
		localMeshClient:         localMeshClient,
	}
}

type meshWorkloadFinder struct {
	clusterName             string
	ctx                     context.Context
	meshWorkloadScanners    []MeshWorkloadScanner
	localMeshWorkloadClient zephyr_core.MeshWorkloadClient
	localMeshClient         zephyr_core.MeshClient
}

func (d *meshWorkloadFinder) StartDiscovery(podController controller.PodController, predicates []predicate.Predicate) error {
	return podController.AddEventHandler(
		d.ctx,
		d,
		predicates...,
	)
}

func (d *meshWorkloadFinder) Create(pod *corev1.Pod) error {
	pod.SetClusterName(d.clusterName)
	logger := logging.BuildEventLogger(d.ctx, logging.CreateEvent, pod)

	discoveredMeshWorkload, err := d.discoverMeshWorkload(pod)

	if err != nil && discoveredMeshWorkload == nil {
		logger.Errorw(MeshWorkloadProcessingError, zap.Error(err))
		return err
	} else if err != nil && discoveredMeshWorkload != nil {
		logger.Warnw(MeshWorkloadProcessingNonFatal, zap.Error(err))
	} else if discoveredMeshWorkload == nil {
		logger.Debugf("MeshWorkload not found for pod %s/%s", pod.Namespace, pod.Name)
		return nil
	}
	return d.createOrUpdateWorkload(discoveredMeshWorkload)
}

func (d *meshWorkloadFinder) Update(old, new *corev1.Pod) error {
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
			return d.localMeshWorkloadClient.Update(d.ctx, newMeshWorkload)
		}
	}
}

func (d *meshWorkloadFinder) Delete(pod *corev1.Pod) error {
	logger := logging.BuildEventLogger(d.ctx, logging.DeleteEvent, pod)
	logger.Error("Deletion of MeshWorkloads is currently not supported")
	return nil
}

func (d *meshWorkloadFinder) Generic(pod *corev1.Pod) error {
	logger := logging.BuildEventLogger(d.ctx, logging.GenericEvent, pod)
	logger.Error("MeshWorkload generic events are not currently supported")
	return nil
}

func (d *meshWorkloadFinder) attachGeneralDiscoveryLabels(controllerRef *core_types.ResourceRef, meshWorkload *discoveryv1alpha1.MeshWorkload) {
	if meshWorkload.Labels == nil {
		meshWorkload.Labels = map[string]string{}
	}
	meshWorkload.Labels[constants.DISCOVERED_BY] = constants.MESH_WORKLOAD_DISCOVERY
	meshWorkload.Labels[constants.CLUSTER] = d.clusterName
	meshWorkload.Labels[constants.KUBE_CONTROLLER_NAME] = controllerRef.GetName()
	meshWorkload.Labels[constants.KUBE_CONTROLLER_NAMESPACE] = controllerRef.GetNamespace()
}

func (d *meshWorkloadFinder) discoverMeshWorkload(pod *corev1.Pod) (*discoveryv1alpha1.MeshWorkload, error) {
	var (
		multiErr               *multierror.Error
		discoveredMeshWorkload *discoveryv1alpha1.MeshWorkload
		controllerRef          *core_types.ResourceRef
	)
	for _, meshWorkloadScanner := range d.meshWorkloadScanners {
		discoveredControllerRef, discoveredMeshWorkloadObjectMeta, err := meshWorkloadScanner.ScanPod(d.ctx, pod)
		if err != nil {
			multiErr = multierror.Append(multiErr, err)
		}
		if discoveredControllerRef != nil {
			controllerRef = discoveredControllerRef

			meshRef, err := d.createMeshResourceRef(d.ctx)
			if err != nil {
				multiErr = multierror.Append(multiErr, err)
				return nil, multiErr.ErrorOrNil()
			}
			discoveredMeshWorkload = &discoveryv1alpha1.MeshWorkload{
				ObjectMeta: discoveredMeshWorkloadObjectMeta,
				Spec: discovery_types.MeshWorkloadSpec{
					KubeController: &discovery_types.MeshWorkloadSpec_KubeController{
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
	return discoveredMeshWorkload, multiErr.ErrorOrNil()
}

func (d *meshWorkloadFinder) createOrUpdateWorkload(discoveredWorkload *discoveryv1alpha1.MeshWorkload) error {
	objectKey, err := client.ObjectKeyFromObject(discoveredWorkload)
	if err != nil {
		return err
	}
	mw, err := d.localMeshWorkloadClient.Get(d.ctx, objectKey)
	if err != nil {
		if errors.IsNotFound(err) {
			return d.localMeshWorkloadClient.Create(d.ctx, discoveredWorkload)
		}
		return err
	}
	// Need to do this, as we need metadata from previous object, (ResourceVersion), for update

	mw.Spec = discoveredWorkload.Spec
	mw.Labels = discoveredWorkload.Labels
	return d.localMeshWorkloadClient.Update(d.ctx, mw)
}

func (d *meshWorkloadFinder) createMeshResourceRef(ctx context.Context) (*core_types.ResourceRef, error) {
	meshList, err := d.localMeshClient.List(ctx, &client.ListOptions{})
	if err != nil {
		return nil, err
	}
	// assume at most one instance of Istio per cluster, thus it must be the Mesh for the MeshWorkload if it exists
	for _, mesh := range meshList.Items {
		if mesh.Spec.Cluster.Name == d.clusterName {
			return &core_types.ResourceRef{
				Name:      mesh.Name,
				Namespace: mesh.Namespace,
				Cluster:   d.clusterName,
			}, nil
		}
	}
	return nil, nil
}
