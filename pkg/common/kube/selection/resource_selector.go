package selection

import (
	"context"

	kubernetes_apps_providers "github.com/solo-io/external-apis/pkg/api/k8s/apps/v1/providers"

	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/stringutils"
	core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/multicluster"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type NotFoundError struct {
	inner error
}

func (e *NotFoundError) Error() string {
	return e.inner.Error()
}
func (e *NotFoundError) Unwrap() error {
	return e.inner
}
func (e *NotFoundError) Is(err error) bool {
	if _, ok := err.(*NotFoundError); ok {
		return true
	}
	return false
}
func (e *NotFoundError) As(target interface{}) bool {
	if t, ok := target.(**NotFoundError); ok {
		*t = e
		return true
	}
	return false
}
func NewNotFoundError(inner error) *NotFoundError {
	return &NotFoundError{
		inner: inner,
	}
}

var (
	KubeServiceNotFound = func(name, namespace, cluster string) error {
		return NewNotFoundError(eris.Errorf("Kubernetes Service with name: %s, namespace: %s, cluster: %s not found", name, namespace, cluster))
	}
	MultipleMeshServicesFound = func(name, namespace, clusterName string) error {
		return NewNotFoundError(eris.Errorf("Multiple MeshServices found with labels %s=%s, %s=%s, %s=%s",
			kube.KUBE_SERVICE_NAME, name,
			kube.KUBE_SERVICE_NAMESPACE, namespace,
			kube.COMPUTE_TARGET, clusterName))
	}
	MultipleMeshWorkloadsFound = func(name, namespace, clusterName string) error {
		return NewNotFoundError(eris.Errorf("Multiple MeshWorkloads found with labels %s=%s, %s=%s, %s=%s",
			kube.KUBE_CONTROLLER_NAME, name,
			kube.KUBE_CONTROLLER_NAMESPACE, namespace,
			kube.COMPUTE_TARGET, clusterName))
	}
	MeshServiceNotFound = func(name, namespace, clusterName string) error {
		return NewNotFoundError(eris.Errorf("No MeshService found with labels %s=%s, %s=%s, %s=%s",
			kube.KUBE_SERVICE_NAME, name,
			kube.KUBE_SERVICE_NAMESPACE, namespace,
			kube.COMPUTE_TARGET, clusterName))
	}
	MeshWorkloadNotFound = func(name, namespace, clusterName string) error {
		return NewNotFoundError(eris.Errorf("No MeshWorkloads found with labels %s=%s, %s=%s, %s=%s",
			kube.KUBE_CONTROLLER_NAME, name,
			kube.KUBE_CONTROLLER_NAMESPACE, namespace,
			kube.COMPUTE_TARGET, clusterName))
	}
	MustProvideClusterName = func(ref *core_types.ResourceRef) error {
		return eris.Errorf("Must provide cluster name in ref %+v", ref)
	}
	MissingComputeTargetLabel = func(resourceName string) error {
		return eris.Errorf("Resource '%s' does not have a "+kube.COMPUTE_TARGET+" label", resourceName)
	}
)

func NewBaseResourceSelector() BaseResourceSelector {
	return &baseResourceSelector{}
}

func NewResourceSelector(
	meshServiceReader smh_discovery.MeshServiceReader,
	meshWorkloadClient smh_discovery.MeshWorkloadReader,
	deploymentClientFactory kubernetes_apps_providers.DeploymentClientFactory,
	dynamicClientGetter multicluster.DynamicClientGetter,
) ResourceSelector {
	return &resourceSelector{
		BaseResourceSelector:    NewBaseResourceSelector(),
		meshServiceReader:       meshServiceReader,
		meshWorkloadClient:      meshWorkloadClient,
		deploymentClientFactory: deploymentClientFactory,
		dynamicClientGetter:     dynamicClientGetter,
	}
}

type baseResourceSelector struct{}

type resourceSelector struct {
	BaseResourceSelector
	meshServiceReader       smh_discovery.MeshServiceReader
	meshWorkloadClient      smh_discovery.MeshWorkloadReader
	deploymentClientFactory kubernetes_apps_providers.DeploymentClientFactory
	dynamicClientGetter     multicluster.DynamicClientGetter
}

func (r *baseResourceSelector) FindMeshServiceByRefSelector(
	meshServices []*smh_discovery.MeshService,
	kubeServiceName string,
	kubeServiceNamespace string,
	kubeServiceCluster string,
) *smh_discovery.MeshService {
	for _, meshService := range meshServices {
		matchesCriteria := meshService.Labels[kube.KUBE_SERVICE_NAME] == kubeServiceName &&
			meshService.Labels[kube.KUBE_SERVICE_NAMESPACE] == kubeServiceNamespace &&
			meshService.Labels[kube.COMPUTE_TARGET] == kubeServiceCluster

		if matchesCriteria {
			return meshService
		}
	}

	return nil
}

func (r *resourceSelector) GetAllMeshServiceByRefSelector(
	ctx context.Context,
	kubeServiceName string,
	kubeServiceNamespace string,
	kubeServiceCluster string,
) (*smh_discovery.MeshService, error) {
	if kubeServiceCluster == "" {
		return nil, MustProvideClusterName(&core_types.ResourceRef{Name: kubeServiceName, Namespace: kubeServiceNamespace})
	}
	destinationKey := client.MatchingLabels{
		kube.KUBE_SERVICE_NAME:      kubeServiceName,
		kube.KUBE_SERVICE_NAMESPACE: kubeServiceNamespace,
		kube.COMPUTE_TARGET:         kubeServiceCluster,
	}
	meshServiceList, err := r.meshServiceReader.ListMeshService(ctx, destinationKey)
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
func (r *resourceSelector) GetAllMeshServicesByServiceSelector(
	ctx context.Context,
	selector *core_types.ServiceSelector,
) ([]*smh_discovery.MeshService, error) {
	meshServiceList, err := r.meshServiceReader.ListMeshService(ctx)
	if err != nil {
		return nil, err
	}
	allMeshServices := convertServicesToPointerSlice(meshServiceList.Items)

	return r.FilterMeshServicesByServiceSelector(allMeshServices, selector)
}

func (r *baseResourceSelector) FilterMeshServicesByServiceSelector(
	meshServices []*smh_discovery.MeshService,
	selector *core_types.ServiceSelector,
) ([]*smh_discovery.MeshService, error) {
	var selectedMeshServices []*smh_discovery.MeshService

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

func (r *resourceSelector) GetMeshWorkloadsByIdentitySelector(
	ctx context.Context,
	identitySelector *core_types.IdentitySelector,
) ([]*smh_discovery.MeshWorkload, error) {
	meshWorkloadList, err := r.meshWorkloadClient.ListMeshWorkload(ctx)
	if err != nil {
		return nil, err
	}

	if identitySelector == nil {
		return convertWorkloadsToPointerSlice(meshWorkloadList.Items), nil
	}

	var matches []*smh_discovery.MeshWorkload
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

func (r *resourceSelector) GetMeshWorkloadsByWorkloadSelector(
	ctx context.Context,
	workloadSelector *core_types.WorkloadSelector,
) ([]*smh_discovery.MeshWorkload, error) {
	meshWorkloadList, err := r.meshWorkloadClient.ListMeshWorkload(ctx)
	if err != nil {
		return nil, err
	}

	// if a selector was not provided or if both of its field are empty, accept everything
	if workloadSelector == nil || (len(workloadSelector.Labels) == 0 && len(workloadSelector.Namespaces) == 0) {
		return convertWorkloadsToPointerSlice(meshWorkloadList.Items), nil
	}

	var matches []*smh_discovery.MeshWorkload

	// for each mesh workload we know about:
	//   - load its deployment
	//   - check whether the deployment labels match the selector labels
	//   - check whether the deployment namespace matches the selector namespaces
	for _, meshWorkloadIter := range meshWorkloadList.Items {
		meshWorkload := meshWorkloadIter // careful not to close over the loop var

		clusterName := meshWorkload.Labels[kube.COMPUTE_TARGET]
		if clusterName == "" {
			return nil, MissingComputeTargetLabel(meshWorkload.GetName())
		}

		dynamicClient, err := r.dynamicClientGetter.GetClientForCluster(ctx, clusterName)
		if err != nil {
			return nil, err
		}

		deploymentClient := r.deploymentClientFactory(dynamicClient)

		workloadController, err := deploymentClient.GetDeployment(ctx, ResourceRefToObjectKey(meshWorkload.Spec.GetKubeController().GetKubeControllerRef()))
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

func (r *resourceSelector) GetMeshWorkloadByRefSelector(
	ctx context.Context,
	podEventWatcherName string,
	podEventWatcherNamespace string,
	podEventWatcherCluster string,
) (*smh_discovery.MeshWorkload, error) {
	if podEventWatcherCluster == "" {
		return nil, MustProvideClusterName(&core_types.ResourceRef{Name: podEventWatcherName, Namespace: podEventWatcherNamespace})
	}
	destinationKey := client.MatchingLabels{
		kube.KUBE_CONTROLLER_NAME:      podEventWatcherName,
		kube.KUBE_CONTROLLER_NAMESPACE: podEventWatcherNamespace,
		kube.COMPUTE_TARGET:            podEventWatcherCluster,
	}
	meshWorkloadList, err := r.meshWorkloadClient.ListMeshWorkload(ctx, destinationKey)
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
	meshServices []*smh_discovery.MeshService,
) *smh_discovery.MeshService {
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
	meshServices []*smh_discovery.MeshService,
) []*smh_discovery.MeshService {
	var selectedMeshServices []*smh_discovery.MeshService
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

func convertServicesToPointerSlice(meshServices []smh_discovery.MeshService) []*smh_discovery.MeshService {
	pointerSlice := make([]*smh_discovery.MeshService, 0, len(meshServices))
	for _, meshService := range meshServices {
		meshService := meshService
		pointerSlice = append(pointerSlice, &meshService)
	}
	return pointerSlice
}

func convertWorkloadsToPointerSlice(meshWorkloads []smh_discovery.MeshWorkload) []*smh_discovery.MeshWorkload {
	pointerSlice := []*smh_discovery.MeshWorkload{}
	for _, meshWorkloadIter := range meshWorkloads {
		meshWorkload := meshWorkloadIter
		pointerSlice = append(pointerSlice, &meshWorkload)
	}
	return pointerSlice
}
