package mesh_workload

import (
	"context"

	pb_types "github.com/gogo/protobuf/types"
	"github.com/hashicorp/go-multierror"
	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	discoveryv1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_core "github.com/solo-io/mesh-projects/pkg/clients/zephyr/discovery"
	"github.com/solo-io/mesh-projects/pkg/logging"
	"github.com/solo-io/mesh-projects/services/common/cluster/core/v1/controller"
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
	logger := logging.BuildEventLogger(d.ctx, logging.CreateEvent, d.clusterName)
	pod.SetClusterName(d.clusterName)
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
	return d.createMeshWorkloadIfNotExists(discoveredMeshWorkload)
}

func (d *meshWorkloadFinder) Update(old, new *corev1.Pod) error {
	logger := logging.BuildEventLogger(d.ctx, logging.UpdateEvent, d.clusterName)
	old.SetClusterName(d.clusterName)
	new.SetClusterName(d.clusterName)
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
		return d.createMeshWorkloadIfNotExists(newMeshWorkload)
	} else if oldMeshWorkload != nil && newMeshWorkload == nil {
		// existing pod is no longer mesh-injected
		// TODO: delete
		return nil
	} else {
		// old and new MeshWorkloads equivalent
		if oldMeshWorkload.Spec.Equal(newMeshWorkload.Spec) {
			return nil
		} else {
			// TODO: delete
			return d.localMeshWorkloadClient.Update(d.ctx, newMeshWorkload)
		}
	}
}

func (d *meshWorkloadFinder) Delete(pod *corev1.Pod) error {
	logger := logging.BuildEventLogger(d.ctx, logging.DeleteEvent, d.clusterName)
	logger.Error("Deletion of MeshWorkloads is currently not supported")
	return nil
}

func (d *meshWorkloadFinder) Generic(pod *corev1.Pod) error {
	logger := logging.BuildEventLogger(d.ctx, logging.GenericEvent, d.clusterName)
	logger.Error("MeshWorkload generic events are not currently supported")
	return nil
}

func (d *meshWorkloadFinder) discoverMeshWorkload(pod *corev1.Pod) (*discoveryv1alpha1.MeshWorkload, error) {
	var multiErr *multierror.Error
	var discoveredMeshWorkload *discoveryv1alpha1.MeshWorkload
	var err error
	for _, meshWorkloadScanner := range d.meshWorkloadScanners {
		discoveredMeshWorkload, err = meshWorkloadScanner.ScanPod(d.ctx, pod)
		if err != nil {
			multiErr = multierror.Append(multiErr, err)
		}
		if discoveredMeshWorkload != nil {
			err = d.populateMeshResourceRef(d.ctx, discoveredMeshWorkload)
			if err != nil {
				multiErr = multierror.Append(multiErr, err)
				return nil, multiErr.ErrorOrNil()
			}
			break
		}
	}
	return discoveredMeshWorkload, multiErr.ErrorOrNil()
}

func (d *meshWorkloadFinder) createMeshWorkloadIfNotExists(discoveredWorkload *discoveryv1alpha1.MeshWorkload) error {
	objectKey, err := client.ObjectKeyFromObject(discoveredWorkload)
	if err != nil {
		return err
	}
	_, err = d.localMeshWorkloadClient.Get(d.ctx, objectKey)
	if errors.IsNotFound(err) {
		return d.localMeshWorkloadClient.Create(d.ctx, discoveredWorkload)
	} else if err != nil {
		return err
	}
	return nil
}

func (d *meshWorkloadFinder) populateMeshResourceRef(ctx context.Context, discoveredWorkload *discoveryv1alpha1.MeshWorkload) error {
	meshList, err := d.localMeshClient.List(ctx, &client.ListOptions{})
	if err != nil {
		return err
	}
	// assume at most one instance of Istio per cluster, thus it must be the Mesh for the MeshWorkload if it exists
	for _, mesh := range meshList.Items {
		if mesh.Spec.Cluster.Name == d.clusterName {
			discoveredWorkload.Spec.Mesh = &core_types.ResourceRef{
				Kind:      &pb_types.StringValue{Value: mesh.TypeMeta.Kind},
				Name:      mesh.Name,
				Namespace: mesh.Namespace,
				Cluster:   &pb_types.StringValue{Value: d.clusterName},
			}
			return nil
		}
	}
	return nil
}
