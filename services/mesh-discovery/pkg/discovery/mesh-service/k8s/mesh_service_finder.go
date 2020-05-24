package k8s

import (
	"context"
	"fmt"
	"strings"

	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_controller "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/controller"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	k8s_core "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1"
	k8s_core_controller "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1/controller"
	"github.com/solo-io/service-mesh-hub/pkg/clients"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/container-runtime"
	"github.com/solo-io/service-mesh-hub/pkg/metadata"
	"github.com/solo-io/service-mesh-hub/services/common/constants"
	"go.uber.org/zap"
	k8s_core_types "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	DiscoveryLabels = func(meshType zephyr_core_types.MeshType, cluster, kubeServiceName, kubeServiceNamespace string) map[string]string {
		return map[string]string{
			constants.DISCOVERED_BY:          constants.MESH_WORKLOAD_DISCOVERY,
			constants.MESH_TYPE:              strings.ToLower(meshType.String()),
			constants.KUBE_SERVICE_NAME:      kubeServiceName,
			constants.KUBE_SERVICE_NAMESPACE: kubeServiceNamespace,
			constants.COMPUTE_TARGET:         cluster,
		}
	}

	skippedLabels = sets.NewString(
		"pod-template-hash",
		"service.istio.io/canonical-revision",
	)
)

func NewMeshServiceFinder(
	ctx context.Context,
	clusterName, writeNamespace string,
	serviceClient k8s_core.ServiceClient,
	meshServiceClient zephyr_discovery.MeshServiceClient,
	meshWorkloadClient zephyr_discovery.MeshWorkloadClient,
	meshClient zephyr_discovery.MeshClient,
) MeshServiceFinder {
	return &meshServiceFinder{
		ctx:                ctx,
		writeNamespace:     writeNamespace,
		clusterName:        clusterName,
		serviceClient:      serviceClient,
		meshServiceClient:  meshServiceClient,
		meshWorkloadClient: meshWorkloadClient,
		meshClient:         meshClient,
	}
}

func (m *meshServiceFinder) StartDiscovery(
	serviceEventWatcher k8s_core_controller.ServiceEventWatcher,
	meshWorkloadEventWatcher zephyr_discovery_controller.MeshWorkloadEventWatcher,
) error {
	allMeshWorkloads, err := m.loadAllMeshWorkloadsInCluster()
	if err != nil {
		return err
	}

	err = m.reconcileMeshServices(allMeshWorkloads)
	if err != nil {
		return err
	}

	err = serviceEventWatcher.AddEventHandler(m.ctx, &k8s_core_controller.ServiceEventHandlerFuncs{
		OnCreate: func(obj *k8s_core_types.Service) error {
			logger := container_runtime.BuildEventLogger(m.ctx, container_runtime.CreateEvent, obj)
			err := m.handleServiceUpsert(logger, obj)
			if err != nil {
				logger.Errorf("%+v", err)
			}

			return nil
		},
		OnUpdate: func(_, new *k8s_core_types.Service) error {
			logger := container_runtime.BuildEventLogger(m.ctx, container_runtime.UpdateEvent, new)
			err := m.handleServiceUpsert(logger, new)
			if err != nil {
				logger.Errorf("%+v", err)
			}

			return nil
		},
		OnDelete: func(obj *k8s_core_types.Service) error {
			logger := container_runtime.BuildEventLogger(m.ctx, container_runtime.DeleteEvent, obj)
			err := m.handleServiceDelete(logger, obj)
			if err != nil {
				logger.Errorf("%+v", err)
			}

			return nil
		},
	})
	if err != nil {
		return err
	}

	return meshWorkloadEventWatcher.AddEventHandler(m.ctx, &zephyr_discovery_controller.MeshWorkloadEventHandlerFuncs{
		OnCreate: func(obj *zephyr_discovery.MeshWorkload) error {
			logger := container_runtime.BuildEventLogger(m.ctx, container_runtime.CreateEvent, obj)
			err := m.handleMeshWorkloadUpsert(logger, obj)
			if err != nil {
				logger.Errorf("%+v", err)
			}

			return nil
		},
		OnUpdate: func(_, new *zephyr_discovery.MeshWorkload) error {
			logger := container_runtime.BuildEventLogger(m.ctx, container_runtime.UpdateEvent, new)
			err := m.handleMeshWorkloadUpsert(logger, new)
			if err != nil {
				logger.Errorf("%+v", err)
			}

			return nil
		},
		OnDelete: func(obj *zephyr_discovery.MeshWorkload) error {
			logger := container_runtime.BuildEventLogger(m.ctx, container_runtime.DeleteEvent, obj)
			err := m.handleMeshWorkloadDelete(logger, obj)
			if err != nil {
				logger.Errorf("%+v", err)
			}

			return nil
		},
	})
}

type meshServiceFinder struct {
	ctx                context.Context
	writeNamespace     string
	clusterName        string
	serviceClient      k8s_core.ServiceClient
	meshServiceClient  zephyr_discovery.MeshServiceClient
	meshWorkloadClient zephyr_discovery.MeshWorkloadClient
	meshClient         zephyr_discovery.MeshClient
}

func (m *meshServiceFinder) loadAllMeshWorkloadsInCluster() ([]*zephyr_discovery.MeshWorkload, error) {
	meshWorkloadList, err := m.meshWorkloadClient.ListMeshWorkload(m.ctx, client.MatchingLabels{
		constants.COMPUTE_TARGET: m.clusterName,
	})
	if err != nil {
		return nil, err
	}

	// at the start, everything we've recorded is "live" so far
	var allMeshWorkloads []*zephyr_discovery.MeshWorkload
	for _, meshWorkloadIter := range meshWorkloadList.Items {
		meshWorkload := meshWorkloadIter

		allMeshWorkloads = append(allMeshWorkloads, &meshWorkload)
	}

	return allMeshWorkloads, nil
}

// handle non-delete events
func (m *meshServiceFinder) handleServiceUpsert(logger *zap.SugaredLogger, service *k8s_core_types.Service) error {
	mesh, backingWorkloads, err := m.findMeshAndWorkloadsForService(service)
	if err != nil {
		return err
	}

	if mesh != nil {
		logger.Debugf("Upserting mesh service for service %s.%s", service.GetName(), service.GetNamespace())
		err = m.upsertMeshService(
			service,
			mesh,
			m.findSubsets(backingWorkloads),
			m.clusterName,
		)
		if err != nil {
			return err
		}
	}

	logger.Debugf("No mesh discovered for service %s.%s", service.GetName(), service.GetNamespace())

	return nil
}

func (m *meshServiceFinder) handleServiceDelete(logger *zap.SugaredLogger, service *k8s_core_types.Service) error {
	mesh, _, err := m.findMeshAndWorkloadsForService(service)
	if err != nil {
		return err
	}

	// not part of any mesh, so no delete action needs to be taken
	if mesh == nil {
		return nil
	}

	err = m.meshServiceClient.DeleteMeshService(m.ctx, client.ObjectKey{
		Name:      m.buildMeshServiceName(service, m.clusterName),
		Namespace: m.writeNamespace,
	})
	if err != nil {
		return err
	}

	return nil
}

// Find the mesh that the service is a part of. Also find the workloads that back it.
// Returning nil, nil, nil from this method means that the service was not a part of any mesh.
func (m *meshServiceFinder) findMeshAndWorkloadsForService(service *k8s_core_types.Service) (*zephyr_discovery.Mesh, []*zephyr_discovery.MeshWorkload, error) {
	// early optimization- bail out early if we know that this service can't select anything
	// otherwise we'll have to check all the mesh workloads
	if len(service.Spec.Selector) == 0 {
		return nil, nil, nil
	}

	meshWorkloads, err := m.meshWorkloadClient.ListMeshWorkload(m.ctx, client.MatchingLabels{
		constants.COMPUTE_TARGET: m.clusterName,
	})
	if err != nil {
		return nil, nil, err
	}

	var backingWorkloads []*zephyr_discovery.MeshWorkload
	var mesh *zephyr_discovery.Mesh
	for _, meshWorkloadIter := range meshWorkloads.Items {
		meshWorkload := meshWorkloadIter

		meshForWorkload, err := m.meshClient.GetMesh(m.ctx, clients.ResourceRefToObjectKey(meshWorkload.Spec.GetMesh()))
		if err != nil {
			return nil, nil, err
		}

		if m.isServiceBackedByWorkload(service, &meshWorkload) {
			mesh = meshForWorkload
			backingWorkloads = append(backingWorkloads, &meshWorkload)
		}
	}

	return mesh, backingWorkloads, nil
}

// handle non-delete events
func (m *meshServiceFinder) handleMeshWorkloadUpsert(logger *zap.SugaredLogger, meshWorkload *zephyr_discovery.MeshWorkload) error {
	podLabels := meshWorkload.Spec.GetKubeController().GetLabels()

	// the `AreLabelsInWhiteList` check later on has undesirable behavior when the "whitelist" is empty,
	// so just handle that manually now- if the pod has no labels, the service cannot select it
	if len(podLabels) == 0 {
		return nil
	}

	workloadMesh, err := m.meshClient.GetMesh(m.ctx, clients.ResourceRefToObjectKey(meshWorkload.Spec.GetMesh()))
	if err != nil {
		return err
	}

	services, err := m.serviceClient.ListService(m.ctx)
	if err != nil {
		return err
	}

	for _, service := range services.Items {
		if m.isServiceBackedByWorkload(&service, meshWorkload) {
			_, backingWorkloads, err := m.findMeshAndWorkloadsForService(&service)

			err = m.upsertMeshService(
				&service,
				workloadMesh,
				m.findSubsets(backingWorkloads),
				m.clusterName,
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *meshServiceFinder) handleMeshWorkloadDelete(logger *zap.SugaredLogger, deletedMeshWorkload *zephyr_discovery.MeshWorkload) error {
	// careful to only delete mesh services that this cluster-scoped mesh service finder instance owns
	if deletedMeshWorkload.Labels[constants.COMPUTE_TARGET] != m.clusterName {
		logger.Debugf("Ignoring mesh workload delete event because cluster does not match for %+v", deletedMeshWorkload.Spec)
		return nil
	}

	// Start with all mesh workloads. We'll filter them in the next step
	allMeshWorkloads, err := m.meshWorkloadClient.ListMeshWorkload(m.ctx, client.MatchingLabels{
		constants.COMPUTE_TARGET: m.clusterName,
	})
	if err != nil {
		return err
	}

	// take care that we don't include the workload that just got deleted
	// (unsure of the controller-runtime cache behavior here- better safe than sorry)
	var remainingMeshWorkloads []*zephyr_discovery.MeshWorkload
	for _, candidateMeshWorkloadIter := range allMeshWorkloads.Items {
		candidateMeshWorkload := &candidateMeshWorkloadIter

		// save all the non-deleted mesh workloads
		if !clients.SameObject(candidateMeshWorkload.ObjectMeta, deletedMeshWorkload.ObjectMeta) {
			remainingMeshWorkloads = append(remainingMeshWorkloads, candidateMeshWorkload)
		}
	}

	err = m.reconcileMeshServices(remainingMeshWorkloads)
	if err != nil {
		return err
	}

	return nil
}

// Accepts a list of "live" mesh workloads on the cluster that this mesh service finder is processing
// This is to support the case where this is run in the context of a delete events rolling in and we want to
// be careful to avoid considering the one that just got deleted.
func (m *meshServiceFinder) reconcileMeshServices(liveMeshWorkloads []*zephyr_discovery.MeshWorkload) error {
	// find the mesh services on this cluster (ie, the only ones that could be backed by the workloads we're handed)
	meshServicesOnCluster, err := m.meshServiceClient.ListMeshService(m.ctx, client.MatchingLabels{
		constants.COMPUTE_TARGET: m.clusterName,
	})
	if err != nil {
		return err
	}

	// for each mesh service on this cluster, check whether it still has workloads backing it
	for _, meshService := range meshServicesOnCluster.Items {
		kubeService, err := m.serviceClient.GetService(m.ctx, clients.ResourceRefToObjectKey(meshService.Spec.GetKubeService().GetRef()))
		if errors.IsNotFound(err) {
			// the service backing this has already disappeared, just clean it up immediately
			err = m.meshServiceClient.DeleteMeshService(m.ctx, clients.ObjectMetaToObjectKey(meshService.ObjectMeta))
			if err != nil {
				return err
			}

			continue
		} else if err != nil {
			return err
		}

		hasBackingWorkloads := false
		for _, meshWorkload := range liveMeshWorkloads {
			hasBackingWorkloads = m.isServiceBackedByWorkload(kubeService, meshWorkload)

			if hasBackingWorkloads {
				break
			}
		}

		// if we couldn't find backing workloads, then delete the mesh service we're processing in this loop iteration
		if !hasBackingWorkloads {
			err = m.meshServiceClient.DeleteMeshService(m.ctx, clients.ObjectMetaToObjectKey(meshService.ObjectMeta))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// expects a list of just the workloads that back the service you're finding subsets for
func (m *meshServiceFinder) findSubsets(backingWorkloads []*zephyr_discovery.MeshWorkload) map[string]*zephyr_discovery_types.MeshServiceSpec_Subset {

	uniqueLabels := make(map[string]sets.String)
	for _, backingWorkload := range backingWorkloads {
		for key, val := range backingWorkload.Spec.GetKubeController().GetLabels() {
			// skip known kubernetes values
			if skippedLabels.Has(key) {
				continue
			}
			existing, ok := uniqueLabels[key]
			if !ok {
				uniqueLabels[key] = sets.NewString(val)
			} else {
				existing.Insert(val)
			}
		}
	}
	/*
		Only select the keys with > 1 value
		The subsets worth noting will be sets of labels which share the same key, but have different values, such as:

			version:
				- v1
				- v2
	*/
	subsets := make(map[string]*zephyr_discovery_types.MeshServiceSpec_Subset)
	for k, v := range uniqueLabels {
		if v.Len() > 1 {
			subsets[k] = &zephyr_discovery_types.MeshServiceSpec_Subset{Values: v.List()}
		}
	}
	return subsets
}

func (m *meshServiceFinder) isServiceBackedByWorkload(
	service *k8s_core_types.Service,
	meshWorkload *zephyr_discovery.MeshWorkload,
) bool {
	workloadCluster := meshWorkload.Labels[constants.COMPUTE_TARGET]

	// If the meshworkload is not on the same cluster as the service, it can be skipped safely
	// The event handler accepts events from MeshWorkloads which may "match" the incoming service
	// but be on a different cluster, so it is important to check that here.
	if workloadCluster != m.clusterName {
		return false
	}

	// if either the service has no selector labels or the mesh workload's corresponding pod has no labels,
	// then this service cannot be backed by this mesh workload
	// the library call below returns true for either case, so we explicitly check for it here
	if len(service.Spec.Selector) == 0 || len(meshWorkload.Spec.GetKubeController().GetLabels()) == 0 {
		return false
	}

	return labels.AreLabelsInWhiteList(service.Spec.Selector, meshWorkload.Spec.GetKubeController().GetLabels())
}

func (m *meshServiceFinder) buildMeshService(
	service *k8s_core_types.Service,
	mesh *zephyr_discovery.Mesh,
	subsets map[string]*zephyr_discovery_types.MeshServiceSpec_Subset,
	clusterName string,
) (*zephyr_discovery.MeshService, error) {
	meshType, err := metadata.MeshToMeshType(mesh)
	if err != nil {
		return nil, err
	}

	return &zephyr_discovery.MeshService{
		ObjectMeta: k8s_meta_types.ObjectMeta{
			Name:      m.buildMeshServiceName(service, clusterName),
			Namespace: m.writeNamespace,
			Labels:    DiscoveryLabels(meshType, clusterName, service.GetName(), service.GetNamespace()),
		},
		Spec: zephyr_discovery_types.MeshServiceSpec{
			KubeService: &zephyr_discovery_types.MeshServiceSpec_KubeService{
				Ref: &zephyr_core_types.ResourceRef{
					Name:      service.GetName(),
					Namespace: service.GetNamespace(),
					Cluster:   clusterName,
				},
				WorkloadSelectorLabels: service.Spec.Selector,
				Labels:                 service.GetLabels(),
				Ports:                  m.convertPorts(service),
			},
			Mesh:    clients.ObjectMetaToResourceRef(mesh.ObjectMeta),
			Subsets: subsets,
		},
	}, nil
}

func (m *meshServiceFinder) convertPorts(service *k8s_core_types.Service) (ports []*zephyr_discovery_types.MeshServiceSpec_KubeService_KubeServicePort) {
	for _, kubePort := range service.Spec.Ports {
		ports = append(ports, &zephyr_discovery_types.MeshServiceSpec_KubeService_KubeServicePort{
			Port:     uint32(kubePort.Port),
			Name:     kubePort.Name,
			Protocol: string(kubePort.Protocol),
		})
	}
	return ports
}

func (m *meshServiceFinder) upsertMeshService(
	service *k8s_core_types.Service,
	mesh *zephyr_discovery.Mesh,
	subsets map[string]*zephyr_discovery_types.MeshServiceSpec_Subset,
	clusterName string,
) error {
	computedMeshService, err := m.buildMeshService(service, mesh, subsets, clusterName)
	if err != nil {
		return err
	}

	existingMeshService, err := m.meshServiceClient.GetMeshService(m.ctx, client.ObjectKey{
		Name:      computedMeshService.GetName(),
		Namespace: computedMeshService.GetNamespace(),
	})
	if errors.IsNotFound(err) {
		err = m.meshServiceClient.CreateMeshService(m.ctx, computedMeshService)
	} else if err != nil {
		return err
	} else if !existingMeshService.Spec.Equal(computedMeshService.Spec) {
		existingMeshService.Spec = computedMeshService.Spec
		existingMeshService.Labels = computedMeshService.Labels
		return m.meshServiceClient.UpdateMeshService(m.ctx, existingMeshService)
	}

	return nil
}

func (m *meshServiceFinder) buildMeshServiceName(service *k8s_core_types.Service, clusterName string) string {
	return fmt.Sprintf("%s-%s-%s", service.GetName(), service.GetNamespace(), clusterName)
}
