package preprocess

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
	InvalidSelectorErr          = eris.New("Incorrectly configured Selector: only one of (labels + namespaces) or (refs) may be declared.")
	ClusterSelectorNotSupported = eris.New("Remote destination cluster selection is not currently supported.")
	KubeServiceNotFound         = func(name, namespace string) error {
		return eris.Errorf("Kubernetes Service with name: %s, namespace: %s not found", name, namespace)
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

func NewMeshServiceSelector(meshServiceClient zephyr_discovery.MeshServiceClient) MeshServiceSelector {
	return &meshServiceSelector{meshServiceClient: meshServiceClient}
}

// List all MeshServices and filter for the ones associated with the k8s Services specified in the selector
func (m *meshServiceSelector) GetMatchingMeshServices(
	ctx context.Context,
	selector *core_types.Selector,
) ([]*discovery_v1alpha1.MeshService, error) {
	var selectedMeshServices []*discovery_v1alpha1.MeshService
	meshServiceList, err := m.meshServiceClient.List(ctx)
	if err != nil {
		return nil, err
	}
	allMeshServices := convertToPointerSlice(meshServiceList.Items)
	// select all MeshServices
	if selector == nil {
		selectedMeshServices = allMeshServices
	} else if selector.GetCluster().GetValue() != "" {
		return nil, ClusterSelectorNotSupported
	} else if selector.GetRefs() != nil && (selector.GetLabels() != nil ||
		selector.GetNamespaces() != nil ||
		selector.GetCluster().GetValue() != "") {
		return nil, InvalidSelectorErr
	} else if selector.GetRefs() != nil {
		// select by Service ResourceRef
		for _, ref := range selector.GetRefs() {
			if ref.GetCluster().GetValue() == "" {
				return nil, MustProvideClusterName(ref)
			}
			selectedMeshService := getMeshServiceByServiceKey(
				ref,
				allMeshServices)
			if selectedMeshService != nil {
				selectedMeshServices = append(selectedMeshServices, selectedMeshService)
			} else {
				// MeshService for referenced k8s Service not found
				return nil, KubeServiceNotFound(ref.GetName(), ref.GetNamespace())
			}
		}
	} else {
		selectedMeshServices = getMeshServicesBySelectorNamespace(selector.GetLabels(), selector.GetNamespaces(), allMeshServices)
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
			selectedRef.GetCluster().GetValue() == kubeServiceRef.GetCluster().GetValue() {
			return meshService
		}
	}
	return nil
}

func getMeshServicesBySelectorNamespace(
	selectors map[string]string,
	namespaces []string,
	meshServices []*discovery_v1alpha1.MeshService,
) []*discovery_v1alpha1.MeshService {
	var selectedMeshServices []*discovery_v1alpha1.MeshService
	for _, meshService := range meshServices {
		kubeService := meshService.Spec.GetKubeService()
		if kubeServiceMatches(selectors, namespaces, kubeService) {
			selectedMeshServices = append(selectedMeshServices, meshService)
		}
	}
	return selectedMeshServices
}

// All selectors must exist. If namespaces is populated, only one namespace must match
func kubeServiceMatches(
	selectors map[string]string,
	namespaces []string,
	kubeService *types.KubeService,
) bool {
	if len(namespaces) > 0 && !stringutils.ContainsString(kubeService.GetRef().GetNamespace(), namespaces) {
		return false
	}
	for k, v := range selectors {
		serviceLabelValue, ok := kubeService.GetLabels()[k]
		if !ok || serviceLabelValue != v {
			return false
		}
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
