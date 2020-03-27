package selector

import (
	"context"

	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/stringutils"
	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	discovery_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/mesh-projects/pkg/clients/zephyr/discovery"
	"github.com/solo-io/mesh-projects/services/common/constants"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	KubeServiceNotFound = func(name, namespace, cluster string) error {
		return eris.Errorf("Kubernetes Service with name: %s, namespace: %s, cluster: %s not found", name, namespace, cluster)
	}
	MultipleMeshServicesFound = func(name, namespace, clusterName string) error {
		return eris.Errorf("Multiple MeshServices found with labels %s=%s, %s=%s, %s=%s",
			constants.KUBE_SERVICE_NAME, name,
			constants.KUBE_SERVICE_NAMESPACE, namespace,
			constants.CLUSTER, clusterName)
	}
	MeshServiceNotFound = func(name, namespace, clusterName string) error {
		return eris.Errorf("No MeshService found with labels %s=%s, %s=%s, %s=%s",
			constants.KUBE_SERVICE_NAME, name,
			constants.KUBE_SERVICE_NAMESPACE, namespace,
			constants.CLUSTER, clusterName)
	}
	MustProvideClusterName = func(ref *core_types.ResourceRef) error {
		return eris.Errorf("Must provide cluster name in ref %+v", ref)
	}
)

type meshServiceSelector struct {
	meshServiceClient zephyr_discovery.MeshServiceClient
}

func NewMeshServiceSelector(meshServiceClient zephyr_discovery.MeshServiceClient) MeshServiceSelector {
	return &meshServiceSelector{meshServiceClient: meshServiceClient}
}

func (m *meshServiceSelector) GetBackingMeshService(
	ctx context.Context,
	kubeServiceName string,
	kubeServiceNamespace string,
	kubeServiceCluster string,
) (*discovery_v1alpha1.MeshService, error) {
	if kubeServiceCluster == "" {
		return nil, MustProvideClusterName(&core_types.ResourceRef{Name: kubeServiceName, Namespace: kubeServiceNamespace})
	}
	destinationKey := client.MatchingLabels(map[string]string{
		constants.KUBE_SERVICE_NAME:      kubeServiceName,
		constants.KUBE_SERVICE_NAMESPACE: kubeServiceNamespace,
		constants.CLUSTER:                kubeServiceCluster,
	})
	meshServiceList, err := m.meshServiceClient.List(ctx, destinationKey)
	if err != nil {
		return nil, err
	}
	// there should only be a single MeshService with the kube Service name/namespace/cluster key
	if len(meshServiceList.Items) > 1 {
		return nil, MultipleMeshServicesFound(kubeServiceName, kubeServiceNamespace, kubeServiceCluster)
	} else if len(meshServiceList.Items) < 1 {
		return nil, MeshServiceNotFound(kubeServiceName, kubeServiceNamespace, kubeServiceCluster)
	}
	return &meshServiceList.Items[0], nil
}

// List all MeshServices and filter for the ones associated with the k8s Services specified in the selector
func (m *meshServiceSelector) GetMatchingMeshServices(
	ctx context.Context,
	selector *core_types.ServiceSelector,
) ([]*discovery_v1alpha1.MeshService, error) {
	var selectedMeshServices []*discovery_v1alpha1.MeshService
	meshServiceList, err := m.meshServiceClient.List(ctx)
	if err != nil {
		return nil, err
	}
	allMeshServices := convertToPointerSlice(meshServiceList.Items)
	// select all MeshServices
	if selector.GetServiceSelectorType() == nil {
		return allMeshServices, nil
	}
	switch selectorType := selector.GetServiceSelectorType().(type) {
	case *core_types.ServiceSelector_Matcher_:
		selectedMeshServices = getMeshServicesBySelectorNamespace(
			selector.GetMatcher().GetLabels(),
			selector.GetMatcher().GetNamespaces(),
			selector.GetMatcher().GetClusters(),
			allMeshServices,
		)
	case *core_types.ServiceSelector_ServiceRefs_:
		for _, ref := range selector.GetServiceRefs().GetServices() {
			if ref.GetCluster() == "" {
				return nil, MustProvideClusterName(ref)
			}
			selectedMeshService := getMeshServiceByServiceKey(ref, allMeshServices)
			if selectedMeshService != nil {
				selectedMeshServices = append(selectedMeshServices, selectedMeshService)
			} else {
				// MeshService for referenced k8s Service not found
				return nil, KubeServiceNotFound(ref.GetName(), ref.GetNamespace(), ref.GetCluster())
			}
		}
	default:
		return nil, eris.Errorf("ServiceSelector has unexpected type %T", selectorType)
	}
	return selectedMeshServices, nil
}

func getMeshServiceByServiceKey(
	selectedRef *core_types.ResourceRef,
	meshServices []*discovery_v1alpha1.MeshService,
) *discovery_v1alpha1.MeshService {
	for _, meshService := range meshServices {
		kubeServiceRef := meshService.Spec.GetKubeService().GetRef()
		if selectedRef.GetName() == kubeServiceRef.GetName() &&
			selectedRef.GetNamespace() == kubeServiceRef.GetNamespace() &&
			selectedRef.GetCluster() == kubeServiceRef.GetCluster() {
			return meshService
		}
	}
	return nil
}

func getMeshServicesBySelectorNamespace(
	selectors map[string]string,
	namespaces []string,
	clusters []string,
	meshServices []*discovery_v1alpha1.MeshService,
) []*discovery_v1alpha1.MeshService {
	var selectedMeshServices []*discovery_v1alpha1.MeshService
	for _, meshService := range meshServices {
		kubeService := meshService.Spec.GetKubeService()
		if kubeServiceMatches(selectors, namespaces, clusters, kubeService) {
			selectedMeshServices = append(selectedMeshServices, meshService)
		}
	}
	return selectedMeshServices
}

/* For a k8s Service to match:
1) If labels is specified, all labels must exist on the k8s Service
2) If namespaces is specified, the k8s must be in one of those namespaces
3) The k8s Service must exist in the specified cluster. If cluster is empty, select across all clusters.
*/
func kubeServiceMatches(
	labels map[string]string,
	namespaces []string,
	clusters []string,
	kubeService *types.KubeService,
) bool {
	if len(namespaces) > 0 && !stringutils.ContainsString(kubeService.GetRef().GetNamespace(), namespaces) {
		return false
	}
	for k, v := range labels {
		serviceLabelValue, ok := kubeService.GetLabels()[k]
		if !ok || serviceLabelValue != v {
			return false
		}
	}
	if len(clusters) > 0 && !stringutils.ContainsString(kubeService.GetRef().GetCluster(), clusters) {
		return false
	}
	return true
}

func convertToPointerSlice(meshServices []discovery_v1alpha1.MeshService) []*discovery_v1alpha1.MeshService {
	pointerSlice := make([]*discovery_v1alpha1.MeshService, 0, len(meshServices))
	for _, meshService := range meshServices {
		meshService := meshService
		pointerSlice = append(pointerSlice, &meshService)
	}
	return pointerSlice
}
