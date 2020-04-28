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
	"github.com/solo-io/service-mesh-hub/pkg/enum_conversion"
	"github.com/solo-io/service-mesh-hub/services/common/constants"
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
			constants.MESH_PLATFORM:          cluster,
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

	err := serviceEventWatcher.AddEventHandler(m.ctx, &ServiceEventHandler{
		Ctx:                 m.ctx,
		HandleServiceUpsert: m.handleServiceUpsert,
	})
	if err != nil {
		return err
	}

	return meshWorkloadEventWatcher.AddEventHandler(m.ctx, &MeshWorkloadEventHandler{
		Ctx:                      m.ctx,
		HandleMeshWorkloadUpsert: m.handleMeshWorkloadUpsert,
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

// handle non-delete events
func (m *meshServiceFinder) handleServiceUpsert(service *k8s_core_types.Service) error {
	// early optimization- bail out early if we know that this service can't select anything
	// otherwise we'll have to check all the mesh workloads
	if len(service.Spec.Selector) == 0 {
		return nil
	}

	meshWorkloads, err := m.meshWorkloadClient.ListMeshWorkload(m.ctx)
	if err != nil {
		return err
	}

	for _, meshWorkload := range meshWorkloads.Items {

		mesh, err := m.meshClient.GetMesh(m.ctx, clients.ResourceRefToObjectKey(meshWorkload.Spec.GetMesh()))
		if err != nil {
			return err
		}

		if m.isServiceBackedByWorkload(service, &meshWorkload, mesh) {
			return m.upsertMeshService(
				service,
				mesh,
				m.findSubsets(service, meshWorkloads, mesh),
				m.clusterName,
			)
		}
	}

	// TODO: handle deletions https://github.com/solo-io/service-mesh-hub/issues/169
	return nil
}

// handle non-delete events
func (m *meshServiceFinder) handleMeshWorkloadUpsert(meshWorkload *zephyr_discovery.MeshWorkload) error {
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
		if m.isServiceBackedByWorkload(&service, meshWorkload, workloadMesh) {
			meshWorkloads, err := m.meshWorkloadClient.ListMeshWorkload(m.ctx)
			if err != nil {
				return err
			}
			return m.upsertMeshService(
				&service,
				workloadMesh,
				m.findSubsets(&service, meshWorkloads, workloadMesh),
				m.clusterName,
			)
		}
	}

	return nil
}

func (m *meshServiceFinder) findSubsets(
	service *k8s_core_types.Service,
	meshWorkloads *zephyr_discovery.MeshWorkloadList,
	mesh *zephyr_discovery.Mesh,
) map[string]*zephyr_discovery_types.MeshServiceSpec_Subset {

	uniqueLabels := make(map[string]sets.String)
	for _, workload := range meshWorkloads.Items {
		if !m.isServiceBackedByWorkload(service, &workload, mesh) {
			continue
		}
		for key, val := range workload.Spec.GetKubeController().GetLabels() {
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
	mesh *zephyr_discovery.Mesh,
) bool {

	// If the meshworkload is not on the same cluster as the service, it can be skipped safely
	// The event handler accepts events from MeshWorkloads which may "match" the incoming service
	// but be on a different cluster, so it is important to check that here.
	if mesh.Spec.GetCluster().GetName() != m.clusterName {
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
	meshType zephyr_core_types.MeshType,
	meshObjectMeta k8s_meta_types.ObjectMeta,
	subsets map[string]*zephyr_discovery_types.MeshServiceSpec_Subset,
	clusterName string,
) *zephyr_discovery.MeshService {
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
			Mesh:    clients.ObjectMetaToResourceRef(meshObjectMeta),
			Subsets: subsets,
		},
	}
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
	meshType, err := enum_conversion.MeshToMeshType(mesh)
	if err != nil {
		return err
	}

	computedMeshService := m.buildMeshService(service, meshType, mesh.ObjectMeta, subsets, clusterName)

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
