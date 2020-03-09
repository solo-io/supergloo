package mesh_service

import (
	"context"
	"fmt"

	protobuf_types "github.com/gogo/protobuf/types"
	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1/controller"
	discovery_types "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	kubernetes_core "github.com/solo-io/mesh-projects/pkg/clients/kubernetes/core"
	discovery_core "github.com/solo-io/mesh-projects/pkg/clients/zephyr/discovery"
	corev1_controllers "github.com/solo-io/mesh-projects/services/common/cluster/core/v1/controller"
	"github.com/solo-io/mesh-projects/services/common/constants"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	DiscoveryLabels = func(cluster, kubeServiceName, kubeServiceNamespace string) map[string]string {
		return map[string]string{
			constants.DISCOVERED_BY:          constants.MESH_WORKLOAD_DISCOVERY,
			constants.MESH_TYPE:              core_types.MeshType_ISTIO.String(),
			constants.KUBE_SERVICE_NAME:      kubeServiceName,
			constants.KUBE_SERVICE_NAMESPACE: kubeServiceNamespace,
			constants.CLUSTER:                cluster,
		}
	}

	skippedLabels = sets.NewString("pod-template-hash")
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
			return m.upsertMeshService(service, meshWorkload.Spec.Mesh, m.findSubsets(service, meshWorkloads))
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
			meshWorkloads, err := m.meshWorkloadClient.List(m.ctx)
			if err != nil {
				return err
			}
			return m.upsertMeshService(&service, meshWorkload.Spec.Mesh, m.findSubsets(&service, meshWorkloads))
		}
	}

	return nil
}

func (m *meshServiceFinder) findSubsets(
	service *corev1.Service,
	meshWorkloads *v1alpha1.MeshWorkloadList,
) map[string]*discovery_types.MeshServiceSpec_Subset {

	uniqueLabels := make(map[string]sets.String)
	for _, workload := range meshWorkloads.Items {
		if !m.isServiceBackedByWorkload(service, &workload) {
			continue
		}
		for key, val := range workload.Spec.GetKubePod().GetLabels() {
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
	subsets := make(map[string]*discovery_types.MeshServiceSpec_Subset)
	for k, v := range uniqueLabels {
		if v.Len() > 1 {
			subsets[k] = &discovery_types.MeshServiceSpec_Subset{Values: v.List()}
		}
	}
	return subsets
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

func (m *meshServiceFinder) buildMeshService(
	service *corev1.Service,
	meshRef *core_types.ResourceRef,
	subsets map[string]*discovery_types.MeshServiceSpec_Subset,
) *v1alpha1.MeshService {
	return &v1alpha1.MeshService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.buildMeshServiceName(service),
			Namespace: m.writeNamespace,
			Labels:    DiscoveryLabels(m.clusterName, service.GetName(), service.GetNamespace()),
		},
		Spec: discovery_types.MeshServiceSpec{
			KubeService: &discovery_types.KubeService{
				Ref: &core_types.ResourceRef{
					Name:      service.GetName(),
					Namespace: service.GetNamespace(),
					Cluster:   &protobuf_types.StringValue{Value: m.clusterName},
				},
				WorkloadSelectorLabels: service.Spec.Selector,
				Labels:                 service.GetLabels(),
			},
			Mesh:    meshRef,
			Subsets: subsets,
		},
	}
}

func (m *meshServiceFinder) upsertMeshService(
	service *corev1.Service,
	meshRef *core_types.ResourceRef,
	subsets map[string]*discovery_types.MeshServiceSpec_Subset,
) error {
	computedMeshService := m.buildMeshService(service, meshRef, subsets)

	existingMeshService, err := m.meshServiceClient.Get(m.ctx, client.ObjectKey{
		Name:      computedMeshService.GetName(),
		Namespace: computedMeshService.GetNamespace(),
	})
	if errors.IsNotFound(err) {
		err = m.meshServiceClient.Create(m.ctx, computedMeshService)
	} else if !existingMeshService.Spec.Equal(computedMeshService.Spec) {
		existingMeshService.Spec = computedMeshService.Spec
		existingMeshService.Labels = computedMeshService.Labels
		err = m.meshServiceClient.Update(m.ctx, existingMeshService)
	}

	return err
}

func (m *meshServiceFinder) buildMeshServiceName(service *corev1.Service) string {
	return fmt.Sprintf("%s-%s-%s", service.GetName(), service.GetNamespace(), m.clusterName)
}
