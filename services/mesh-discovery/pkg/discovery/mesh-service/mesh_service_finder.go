package mesh_service

import (
	"context"
	"fmt"

	protobuf_types "github.com/gogo/protobuf/types"
	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1/controller"
	"github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	kubernetes_core "github.com/solo-io/mesh-projects/pkg/clients/kubernetes/core"
	discovery_core "github.com/solo-io/mesh-projects/pkg/clients/zephyr/discovery"
	corev1_controllers "github.com/solo-io/mesh-projects/services/common/cluster/core/v1/controller"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewMeshServiceFinder(
	ctx context.Context,
	clusterName string,
	writeNamespace string,
	serviceClient kubernetes_core.ServiceClient,
	meshServiceClient discovery_core.MeshServiceClient,
	meshWorkloadClient discovery_core.MeshWorkloadClient,
) MeshServiceFinder {

	return &meshServiceFinder{
		ctx:                ctx,
		clusterName:        clusterName,
		writeNamespace:     writeNamespace,
		serviceClient:      serviceClient,
		meshServiceClient:  meshServiceClient,
		meshWorkloadClient: meshWorkloadClient,
	}
}

func (m *meshServiceFinder) StartDiscovery(
	serviceController corev1_controllers.ServiceController,
	meshWorkloadController controller.MeshWorkloadController,
) error {

	err := serviceController.AddEventHandler(m.ctx, &ServiceEventHandler{
		Ctx:                 m.ctx,
		ClusterName:         m.clusterName,
		HandleServiceUpsert: m.handleServiceUpsert,
	})
	if err != nil {
		return err
	}

	return meshWorkloadController.AddEventHandler(m.ctx, &MeshWorkloadEventHandler{
		Ctx:                      m.ctx,
		ClusterName:              m.clusterName,
		HandleMeshWorkloadUpsert: m.handleMeshWorkloadUpsert,
	})
}

type meshServiceFinder struct {
	ctx                context.Context
	clusterName        string
	writeNamespace     string
	serviceClient      kubernetes_core.ServiceClient
	meshServiceClient  discovery_core.MeshServiceClient
	meshWorkloadClient discovery_core.MeshWorkloadClient
}

// handle non-delete events
func (m *meshServiceFinder) handleServiceUpsert(service *corev1.Service) error {
	// early optimization- bail out early if we know that this service can't select anything
	// otherwise we'll have to check all the mesh workloads
	if len(service.Spec.Selector) == 0 {
		return nil
	}

	meshWorkloads, err := m.meshWorkloadClient.List(m.ctx)
	if err != nil {
		return err
	}

	for _, meshWorkload := range meshWorkloads.Items {
		if m.isServiceBackedByWorkload(service, &meshWorkload) {
			return m.upsertMeshService(service, meshWorkload.Spec.Mesh)
		}
	}

	// TODO: handle deletions https://github.com/solo-io/mesh-projects/issues/169
	return nil
}

// handle non-delete events
func (m *meshServiceFinder) handleMeshWorkloadUpsert(meshWorkload *v1alpha1.MeshWorkload) error {
	podLabels := meshWorkload.Spec.GetKubePod().GetLabels()

	// the `AreLabelsInWhiteList` check later on has undesirable behavior when the "whitelist" is empty,
	// so just handle that manually now- if the pod has no labels, the service cannot select it
	if len(podLabels) == 0 {
		return nil
	}

	services, err := m.serviceClient.List(m.ctx)
	if err != nil {
		return err
	}

	for _, service := range services.Items {
		if m.isServiceBackedByWorkload(&service, meshWorkload) {
			return m.upsertMeshService(&service, meshWorkload.Spec.Mesh)
		}
	}

	return nil
}

func (m *meshServiceFinder) isServiceBackedByWorkload(service *corev1.Service, meshWorkload *v1alpha1.MeshWorkload) bool {
	// if either the service has no selector labels or the mesh workload's corresponding pod has no labels,
	// then this service cannot be backed by this mesh workload
	// the library call below returns true for either case, so we explicitly check for it here
	if len(service.Spec.Selector) == 0 || len(meshWorkload.Spec.GetKubePod().GetLabels()) == 0 {
		return false
	}

	return labels.AreLabelsInWhiteList(service.Spec.Selector, meshWorkload.Spec.GetKubePod().GetLabels())
}

func (m *meshServiceFinder) buildMeshService(service *corev1.Service, meshRef *core_types.ResourceRef) *v1alpha1.MeshService {
	return &v1alpha1.MeshService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.buildMeshServiceName(service),
			Namespace: m.writeNamespace,
		},
		Spec: types.MeshServiceSpec{
			KubeService: &types.KubeService{
				Ref: &core_types.ResourceRef{
					Name:      service.GetName(),
					Namespace: service.GetNamespace(),
					Cluster:   &protobuf_types.StringValue{Value: m.clusterName},
				},
				SelectorLabels: service.Spec.Selector,
			},
			Mesh: meshRef,
		},
	}
}

func (m *meshServiceFinder) upsertMeshService(service *corev1.Service, meshRef *core_types.ResourceRef) error {
	computedMeshService := m.buildMeshService(service, meshRef)

	existingMeshService, err := m.meshServiceClient.Get(m.ctx, client.ObjectKey{
		Name:      computedMeshService.GetName(),
		Namespace: computedMeshService.GetNamespace(),
	})
	if errors.IsNotFound(err) {
		err = m.meshServiceClient.Create(m.ctx, computedMeshService)
	} else if !existingMeshService.Spec.Equal(computedMeshService.Spec) {
		err = m.meshServiceClient.Update(m.ctx, computedMeshService)
	}

	return err
}

func (m *meshServiceFinder) buildMeshServiceName(service *corev1.Service) string {
	return fmt.Sprintf("%s-%s-%s", service.GetName(), service.GetNamespace(), m.clusterName)
}
