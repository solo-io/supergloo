package selector

import (
	"context"

	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/stringutils"
	core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	kubernetes_apps "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/apps/v1"
	"github.com/solo-io/service-mesh-hub/pkg/clients"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/compute-target/k8s"
	"github.com/solo-io/service-mesh-hub/services/common/constants"
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
			constants.COMPUTE_TARGET, clusterName)
	}
	MultipleMeshWorkloadsFound = func(name, namespace, clusterName string) error {
		return eris.Errorf("Multiple MeshWorkloads found with labels %s=%s, %s=%s, %s=%s",
			constants.KUBE_CONTROLLER_NAME, name,
			constants.KUBE_CONTROLLER_NAMESPACE, namespace,
			constants.COMPUTE_TARGET, clusterName)
	}
	MeshServiceNotFound = func(name, namespace, clusterName string) error {
		return eris.Errorf("No MeshService found with labels %s=%s, %s=%s, %s=%s",
			constants.KUBE_SERVICE_NAME, name,
			constants.KUBE_SERVICE_NAMESPACE, namespace,
			constants.COMPUTE_TARGET, clusterName)
	}
	MeshWorkloadNotFound = func(name, namespace, clusterName string) error {
		return eris.Errorf("No MeshWorkloads found with labels %s=%s, %s=%s, %s=%s",
			constants.KUBE_CONTROLLER_NAME, name,
			constants.KUBE_CONTROLLER_NAMESPACE, namespace,
			constants.COMPUTE_TARGET, clusterName)
	}
	MustProvideClusterName = func(ref *core_types.ResourceRef) error {
		return eris.Errorf("Must provide cluster name in ref %+v", ref)
	}
	MissingComputeTargetLabel = func(resourceName string) error {
		return eris.Errorf("Resource '%s' does not have a "+constants.COMPUTE_TARGET+" label", resourceName)
	}
)

func NewResourceSelector(
	meshServiceClient zephyr_discovery.MeshServiceClient,
	meshWorkloadClient zephyr_discovery.MeshWorkloadClient,
	deploymentClientFactory kubernetes_apps.DeploymentClientFactory,
	dynamicClientGetter mc_manager.DynamicClientGetter,
) ResourceSelector {
	return &resourceSelector{
		meshServiceClient:       meshServiceClient,
		meshWorkloadClient:      meshWorkloadClient,
		deploymentClientFactory: deploymentClientFactory,
		dynamicClientGetter:     dynamicClientGetter,
	}
}

type resourceSelector struct {
	meshServiceClient       zephyr_discovery.MeshServiceClient
	meshWorkloadClient      zephyr_discovery.MeshWorkloadClient
	deploymentClientFactory kubernetes_apps.DeploymentClientFactory
	dynamicClientGetter     mc_manager.DynamicClientGetter
}

func (b *resourceSelector) FindMeshServiceByRefSelector(
	meshServices []*zephyr_discovery.MeshService,
	kubeServiceName string,
	kubeServiceNamespace string,
	kubeServiceCluster string,
) *zephyr_discovery.MeshService {
	for _, meshService := range meshServices {
		matchesCriteria := meshService.Labels[constants.KUBE_SERVICE_NAME] == kubeServiceName &&
			meshService.Labels[constants.KUBE_SERVICE_NAMESPACE] == kubeServiceNamespace &&
			meshService.Labels[constants.COMPUTE_TARGET] == kubeServiceCluster

		if matchesCriteria {
			return meshService
		}
	}

	return nil
}

func (b *resourceSelector) GetAllMeshServiceByRefSelector(
	ctx context.Context,
	kubeServiceName string,
	kubeServiceNamespace string,
	kubeServiceCluster string,
) (*zephyr_discovery.MeshService, error) {
	if kubeServiceCluster == "" {
		return nil, MustProvideClusterName(&core_types.ResourceRef{Name: kubeServiceName, Namespace: kubeServiceNamespace})
	}
	destinationKey := client.MatchingLabels{
		constants.KUBE_SERVICE_NAME:      kubeServiceName,
		constants.KUBE_SERVICE_NAMESPACE: kubeServiceNamespace,
		constants.COMPUTE_TARGET:         kubeServiceCluster,
	}
	meshServiceList, err := b.meshServiceClient.ListMeshService(ctx, destinationKey)
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
func (b *resourceSelector) GetAllMeshServicesByServiceSelector(
	ctx context.Context,
	selector *core_types.ServiceSelector,
) ([]*zephyr_discovery.MeshService, error) {
	meshServiceList, err := b.meshServiceClient.ListMeshService(ctx)
	if err != nil {
		return nil, err
	}
	allMeshServices := convertServicesToPointerSlice(meshServiceList.Items)

	return b.FilterMeshServicesByServiceSelector(allMeshServices, selector)
}

func (b *resourceSelector) FilterMeshServicesByServiceSelector(
	meshServices []*zephyr_discovery.MeshService,
	selector *core_types.ServiceSelector,
) ([]*zephyr_discovery.MeshService, error) {
	var selectedMeshServices []*zephyr_discovery.MeshService

	// select all MeshServices
	if selector.GetServiceSelectorType() == nil {
		return meshServices, nil
	}
	switch selectorType := selector.GetServiceSelectorType().(type) {
	case *core_types.ServiceSelector_Matcher_:
		selectedMeshServices = getMeshServicesBySelectorNamespace(
			selector.GetMatcher().GetLabels(),
			selector.GetMatcher().GetNamespaces(),
			selector.GetMatcher().GetClusters(),
			meshServices,
		)
	case *core_types.ServiceSelector_ServiceRefs_:
		for _, ref := range selector.GetServiceRefs().GetServices() {
			if ref.GetCluster() == "" {
				return nil, MustProvideClusterName(ref)
			}
			selectedMeshService := getMeshServiceByServiceKey(ref, meshServices)
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

func (b *resourceSelector) GetMeshWorkloadsByIdentitySelector(
	ctx context.Context,
	identitySelector *core_types.IdentitySelector,
) ([]*zephyr_discovery.MeshWorkload, error) {
	meshWorkloadList, err := b.meshWorkloadClient.ListMeshWorkload(ctx)
	if err != nil {
		return nil, err
	}

	if identitySelector == nil {
		return convertWorkloadsToPointerSlice(meshWorkloadList.Items), nil
	}

	var matches []*zephyr_discovery.MeshWorkload
	for _, workloadIter := range meshWorkloadList.Items {
		workload := workloadIter // careful not to close over the loop var
		switch identitySelector.GetIdentitySelectorType().(type) {
		case *core_types.IdentitySelector_Matcher_:
			namespaces := identitySelector.GetMatcher().GetNamespaces()
			clusters := identitySelector.GetMatcher().GetClusters()

			namespaceMatches := len(namespaces) == 0 || stringutils.ContainsString(workload.Spec.GetKubeController().GetKubeControllerRef().GetNamespace(), namespaces)
			clusterMatches := len(clusters) == 0 || stringutils.ContainsString(workload.Spec.GetKubeController().GetKubeControllerRef().GetCluster(), clusters)

			if namespaceMatches && clusterMatches {
				matches = append(matches, &workload)
			}

		case *core_types.IdentitySelector_ServiceAccountRefs_:
			for _, ref := range identitySelector.GetServiceAccountRefs().GetServiceAccounts() {
				if ref.GetCluster() == "" {
					return nil, MustProvideClusterName(ref)
				}

				if ref.GetNamespace() == workload.Spec.GetKubeController().GetKubeControllerRef().GetNamespace() && ref.GetName() == workload.Spec.GetKubeController().GetServiceAccountName() {
					matches = append(matches, &workload)
				}
			}
		default:
			return nil, eris.Errorf("IdentitySelector has unexpected type %T", identitySelector)
		}
	}

	return matches, nil
}

func (b *resourceSelector) GetMeshWorkloadsByWorkloadSelector(
	ctx context.Context,
	workloadSelector *core_types.WorkloadSelector,
) ([]*zephyr_discovery.MeshWorkload, error) {
	meshWorkloadList, err := b.meshWorkloadClient.ListMeshWorkload(ctx)
	if err != nil {
		return nil, err
	}

	// if a selector was not provided or if both of its field are empty, accept everything
	if workloadSelector == nil || (len(workloadSelector.Labels) == 0 && len(workloadSelector.Namespaces) == 0) {
		return convertWorkloadsToPointerSlice(meshWorkloadList.Items), nil
	}

	var matches []*zephyr_discovery.MeshWorkload

	// for each mesh workload we know about:
	//   - load its deployment
	//   - check whether the deployment labels match the selector labels
	//   - check whether the deployment namespace matches the selector namespaces
	for _, meshWorkloadIter := range meshWorkloadList.Items {
		meshWorkload := meshWorkloadIter // careful not to close over the loop var

		clusterName := meshWorkload.Labels[constants.COMPUTE_TARGET]
		if clusterName == "" {
			return nil, MissingComputeTargetLabel(meshWorkload.GetName())
		}

		dynamicClient, err := b.dynamicClientGetter.GetClientForCluster(ctx, clusterName)
		if err != nil {
			return nil, err
		}

		deploymentClient := b.deploymentClientFactory(dynamicClient)

		workloadController, err := deploymentClient.GetDeployment(ctx, clients.ResourceRefToObjectKey(meshWorkload.Spec.GetKubeController().GetKubeControllerRef()))
		if err != nil {
			return nil, err
		}

		// consider the selector labels a subset of the controller labels if:
		//   the selector did not provide labels, OR
		//   every label in the selector appears with the same value in the controller
		labelsAreSubset := true
		if len(workloadSelector.Labels) > 0 {
			if len(workloadController.Labels) == 0 {
				labelsAreSubset = false
			} else {
				for k, v := range workloadSelector.Labels {
					if workloadController.Labels[k] != v {
						labelsAreSubset = false
						break
					}
				}
			}
		}

		namespaceMatches := len(workloadSelector.Namespaces) == 0 || stringutils.ContainsString(workloadController.GetNamespace(), workloadSelector.Namespaces)

		if labelsAreSubset && namespaceMatches {
			matches = append(matches, &meshWorkload)
		}
	}

	return matches, nil
}

func (b *resourceSelector) GetMeshWorkloadByRefSelector(
	ctx context.Context,
	podEventWatcherName string,
	podEventWatcherNamespace string,
	podEventWatcherCluster string,
) (*zephyr_discovery.MeshWorkload, error) {
	if podEventWatcherCluster == "" {
		return nil, MustProvideClusterName(&core_types.ResourceRef{Name: podEventWatcherName, Namespace: podEventWatcherNamespace})
	}
	destinationKey := client.MatchingLabels{
		constants.KUBE_CONTROLLER_NAME:      podEventWatcherName,
		constants.KUBE_CONTROLLER_NAMESPACE: podEventWatcherNamespace,
		constants.COMPUTE_TARGET:            podEventWatcherCluster,
	}
	meshWorkloadList, err := b.meshWorkloadClient.ListMeshWorkload(ctx, destinationKey)
	if err != nil {
		return nil, err
	}
	// there should only be a single MeshService with the kube Service name/namespace/cluster key
	if len(meshWorkloadList.Items) > 1 {
		return nil, MultipleMeshWorkloadsFound(podEventWatcherName, podEventWatcherNamespace, podEventWatcherCluster)
	} else if len(meshWorkloadList.Items) < 1 {
		return nil, MeshWorkloadNotFound(podEventWatcherName, podEventWatcherNamespace, podEventWatcherCluster)
	}
	return &meshWorkloadList.Items[0], nil
}

func getMeshServiceByServiceKey(
	selectedRef *core_types.ResourceRef,
	meshServices []*zephyr_discovery.MeshService,
) *zephyr_discovery.MeshService {
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
	meshServices []*zephyr_discovery.MeshService,
) []*zephyr_discovery.MeshService {
	var selectedMeshServices []*zephyr_discovery.MeshService
	for _, meshService := range meshServices {
		kubeService := meshService.Spec.GetKubeService()
		if KubeServiceMatches(selectors, namespaces, clusters, kubeService) {
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
func KubeServiceMatches(
	labels map[string]string,
	namespaces []string,
	clusters []string,
	kubeService *types.MeshServiceSpec_KubeService,
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

func convertServicesToPointerSlice(meshServices []zephyr_discovery.MeshService) []*zephyr_discovery.MeshService {
	pointerSlice := make([]*zephyr_discovery.MeshService, 0, len(meshServices))
	for _, meshService := range meshServices {
		meshService := meshService
		pointerSlice = append(pointerSlice, &meshService)
	}
	return pointerSlice
}

func convertWorkloadsToPointerSlice(meshWorkloads []zephyr_discovery.MeshWorkload) []*zephyr_discovery.MeshWorkload {
	pointerSlice := []*zephyr_discovery.MeshWorkload{}
	for _, meshWorkloadIter := range meshWorkloads {
		meshWorkload := meshWorkloadIter
		pointerSlice = append(pointerSlice, &meshWorkload)
	}
	return pointerSlice
}
