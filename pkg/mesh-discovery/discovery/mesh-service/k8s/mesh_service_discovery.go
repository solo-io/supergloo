package k8s

import (
	"context"
	"strings"

	k8s_core_providers "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/providers"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery_sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/sets"
	smh_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/common/container-runtime"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube"
	kube_metadata "github.com/solo-io/service-mesh-hub/pkg/common/kube/metadata"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/selection"
	"github.com/solo-io/service-mesh-hub/pkg/common/metadata"
	skv2_sets "github.com/solo-io/skv2/contrib/pkg/sets"
	"github.com/solo-io/skv2/pkg/multicluster"
	k8s_core_types "k8s.io/api/core/v1"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:generate mockgen -source ./mesh_service_discovery.go -destination ./mocks/mock_interfaces.go -package service_discovery_mocks

type MeshServiceDiscovery interface {
	// Ensure that the existing MeshService CR's match the set of discovered MeshServices,
	// creating, updating, or deleting MeshServices as necessary.
	// TODO: replace client writes with an output snapshot
	DiscoverMeshServices(ctx context.Context, clusterName string) error
}

var (
	DiscoveryLabels = func(meshType smh_core_types.MeshType, cluster, kubeServiceName, kubeServiceNamespace string) map[string]string {
		return map[string]string{
			kube.DISCOVERED_BY:          kube.MESH_WORKLOAD_DISCOVERY,
			kube.MESH_TYPE:              strings.ToLower(meshType.String()),
			kube.KUBE_SERVICE_NAME:      kubeServiceName,
			kube.KUBE_SERVICE_NAMESPACE: kubeServiceNamespace,
			kube.COMPUTE_TARGET:         cluster,
		}
	}

	skippedLabels = sets.NewString(
		"pod-template-hash",
		"service.istio.io/canonical-revision",
	)
)

func NewMeshServiceDiscovery(
	meshServiceClient smh_discovery.MeshServiceClient,
	meshWorkloadClient smh_discovery.MeshWorkloadClient,
	meshClient smh_discovery.MeshClient,
	serviceClientFactory k8s_core_providers.ServiceClientFactory,
	mcClient multicluster.Client,
) MeshServiceDiscovery {
	return &meshServiceDiscovery{
		meshServiceClient:    meshServiceClient,
		meshWorkloadClient:   meshWorkloadClient,
		meshClient:           meshClient,
		serviceClientFactory: serviceClientFactory,
		mcClient:             mcClient,
	}
}

type meshServiceDiscovery struct {
	meshServiceClient    smh_discovery.MeshServiceClient
	meshWorkloadClient   smh_discovery.MeshWorkloadClient
	meshClient           smh_discovery.MeshClient
	serviceClientFactory k8s_core_providers.ServiceClientFactory
	mcClient             multicluster.Client
}

func (m *meshServiceDiscovery) DiscoverMeshServices(ctx context.Context, clusterName string) error {
	existingMeshServices, err := m.getExistingMeshServices(ctx, clusterName)
	if err != nil {
		return err
	}
	discoveredMeshServices, err := m.discoverMeshServices(ctx, clusterName)
	if err != nil {
		return err
	}
	existingMeshServiceMap := existingMeshServices.Map()
	// For each service that is discovered, create if it doesn't exist or update it if the spec has changed.
	for _, discoveredMeshService := range discoveredMeshServices.List() {
		existingMeshService, ok := existingMeshServiceMap[skv2_sets.Key(discoveredMeshService)]
		if !ok || !existingMeshService.Spec.Equal(discoveredMeshService.Spec) {
			err = m.meshServiceClient.UpsertMeshService(ctx, discoveredMeshService)
			if err != nil {
				return err
			}
		}
	}
	// Delete MeshServices that no longer exist.
	for _, existingMeshService := range existingMeshServices.Difference(discoveredMeshServices).List() {
		err = m.meshServiceClient.DeleteMeshService(ctx, selection.ObjectMetaToObjectKey(existingMeshService.ObjectMeta))
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *meshServiceDiscovery) getExistingMeshServices(
	ctx context.Context,
	clusterName string,
) (smh_discovery_sets.MeshServiceSet, error) {
	existingMeshServices := smh_discovery_sets.NewMeshServiceSet()
	meshServiceList, err := m.meshServiceClient.ListMeshService(ctx, client.MatchingLabels{
		kube.COMPUTE_TARGET: clusterName,
	})
	if err != nil {
		return nil, err
	}
	for _, meshService := range meshServiceList.Items {
		meshService := meshService
		existingMeshServices.Insert(&meshService)
	}
	return existingMeshServices, nil
}

func (m *meshServiceDiscovery) discoverMeshServices(
	ctx context.Context,
	clusterName string,
) (smh_discovery_sets.MeshServiceSet, error) {
	discoveredMeshServices := smh_discovery_sets.NewMeshServiceSet()
	clusterClient, err := m.mcClient.Cluster(clusterName)
	if err != nil {
		return nil, err
	}
	serviceList, err := m.serviceClientFactory(clusterClient).ListService(ctx)
	if err != nil {
		return nil, err
	}
	for _, service := range serviceList.Items {
		service := service
		mesh, backingWorkloads, err := m.findMeshAndWorkloadsForService(ctx, clusterName, &service)
		if err != nil {
			return nil, err
		}
		// Only discover services that are backed by workloads
		if len(backingWorkloads) == 0 {
			continue
		}
		discoveredMeshService, err := m.buildMeshService(&service, mesh, m.findSubsets(backingWorkloads), clusterName)
		if err != nil {
			return nil, err
		}
		discoveredMeshServices.Insert(discoveredMeshService)
	}
	return discoveredMeshServices, nil
}

func (m *meshServiceDiscovery) findMeshAndWorkloadsForService(
	ctx context.Context,
	clusterName string,
	service *k8s_core_types.Service,
) (*smh_discovery.Mesh, []*smh_discovery.MeshWorkload, error) {
	// early optimization- bail out early if we know that this service can't select anything
	// otherwise we'll have to check all the mesh workloads
	if len(service.Spec.Selector) == 0 {
		return nil, nil, nil
	}
	meshWorkloads, err := m.meshWorkloadClient.ListMeshWorkload(ctx, client.MatchingLabels{
		kube.COMPUTE_TARGET: clusterName,
	})
	if err != nil {
		return nil, nil, err
	}
	var backingWorkloads []*smh_discovery.MeshWorkload
	var mesh *smh_discovery.Mesh
	for _, meshWorkloadIter := range meshWorkloads.Items {
		meshWorkload := meshWorkloadIter
		meshForWorkload, err := m.meshClient.GetMesh(ctx, selection.ResourceRefToObjectKey(meshWorkload.Spec.GetMesh()))
		if err != nil {
			return nil, nil, err
		}
		if m.isServiceBackedByWorkload(clusterName, service, &meshWorkload) {
			mesh = meshForWorkload
			backingWorkloads = append(backingWorkloads, &meshWorkload)
		}
	}
	return mesh, backingWorkloads, nil
}

// expects a list of just the workloads that back the service you're finding subsets for
func (m *meshServiceDiscovery) findSubsets(backingWorkloads []*smh_discovery.MeshWorkload) map[string]*smh_discovery_types.MeshServiceSpec_Subset {

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
	subsets := make(map[string]*smh_discovery_types.MeshServiceSpec_Subset)
	for k, v := range uniqueLabels {
		if v.Len() > 1 {
			subsets[k] = &smh_discovery_types.MeshServiceSpec_Subset{Values: v.List()}
		}
	}
	return subsets
}

func (m *meshServiceDiscovery) isServiceBackedByWorkload(
	clusterName string,
	service *k8s_core_types.Service,
	meshWorkload *smh_discovery.MeshWorkload,
) bool {
	workloadCluster := meshWorkload.Labels[kube.COMPUTE_TARGET]

	// If the meshworkload is not on the same cluster as the service, it can be skipped safely
	// The event handler accepts events from MeshWorkloads which may "match" the incoming service
	// but be on a different cluster, so it is important to check that here.
	if workloadCluster != clusterName {
		return false
	}

	// if either the service has no selector labels or the mesh workload's corresponding pod has no labels,
	// then this service cannot be backed by this mesh workload
	// the library call below returns true for either case, so we explicitly check for it here
	if len(service.Spec.Selector) == 0 || len(meshWorkload.Spec.GetKubeController().GetLabels()) == 0 {
		return false
	}

	// If service not in same namespace as workload, continue
	if service.GetNamespace() != meshWorkload.Spec.GetKubeController().GetKubeControllerRef().GetNamespace() {
		return false
	}

	return labels.AreLabelsInWhiteList(service.Spec.Selector, meshWorkload.Spec.GetKubeController().GetLabels())
}

func (m *meshServiceDiscovery) buildMeshService(
	service *k8s_core_types.Service,
	mesh *smh_discovery.Mesh,
	subsets map[string]*smh_discovery_types.MeshServiceSpec_Subset,
	clusterName string,
) (*smh_discovery.MeshService, error) {
	meshType, err := kube_metadata.MeshToMeshType(mesh)
	if err != nil {
		return nil, err
	}

	return &smh_discovery.MeshService{
		ObjectMeta: k8s_meta_types.ObjectMeta{
			Name:      metadata.BuildMeshServiceName(service, clusterName),
			Namespace: container_runtime.GetWriteNamespace(),
			Labels:    DiscoveryLabels(meshType, clusterName, service.GetName(), service.GetNamespace()),
		},
		Spec: smh_discovery_types.MeshServiceSpec{
			KubeService: &smh_discovery_types.MeshServiceSpec_KubeService{
				Ref: &smh_core_types.ResourceRef{
					Name:      service.GetName(),
					Namespace: service.GetNamespace(),
					Cluster:   clusterName,
				},
				WorkloadSelectorLabels: service.Spec.Selector,
				Labels:                 service.GetLabels(),
				Ports:                  m.convertPorts(service),
			},
			Mesh:    selection.ObjectMetaToResourceRef(mesh.ObjectMeta),
			Subsets: subsets,
		},
	}, nil
}

func (m *meshServiceDiscovery) convertPorts(service *k8s_core_types.Service) (ports []*smh_discovery_types.MeshServiceSpec_KubeService_KubeServicePort) {
	for _, kubePort := range service.Spec.Ports {
		ports = append(ports, &smh_discovery_types.MeshServiceSpec_KubeService_KubeServicePort{
			Port:     uint32(kubePort.Port),
			Name:     kubePort.Name,
			Protocol: string(kubePort.Protocol),
		})
	}
	return ports
}
