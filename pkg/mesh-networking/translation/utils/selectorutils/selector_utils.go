package selectorutils

import (
	"github.com/solo-io/go-utils/stringutils"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/solo-io/skv2/pkg/ezkube"
)

func SelectorMatchesWorkload(selectors []*v1alpha2.WorkloadSelector, workload *discoveryv1alpha2.Workload,
	matchingMeshes map[string]bool, matchingVirtualMeshes map[string]bool, meshToVirtualMesh map[string]string) bool {

	var selectorMatches bool
	if len(selectors) == 0 {
		selectorMatches = true
	} else {
		for _, selector := range selectors {
			kubeWorkload := workload.Spec.GetKubernetes()
			if kubeWorkload != nil {
				if kubeWorkloadMatches(
					selector.GetLabels(),
					selector.GetNamespaces(),
					kubeWorkload,
				) {
					selectorMatches = true
					break
				}
			}
		}
	}

	return selectorMatches && meshMatches(workload.Spec.GetMesh(), matchingMeshes, matchingVirtualMeshes, meshToVirtualMesh)
}

func IdentityMatchesWorkload(selectors []*v1alpha2.IdentitySelector, workload *discoveryv1alpha2.Workload,
	matchingMeshes map[string]bool, matchingVirtualMeshes map[string]bool, meshToVirtualMesh map[string]string) bool {

	var selectorMatches bool
	if len(selectors) == 0 {
		selectorMatches = true
	} else {
		for _, selector := range selectors {
			kubeWorkload := workload.Spec.GetKubernetes()
			if kubeWorkload != nil {
				if kubeWorkloadIdentityMatcher := selector.GetKubeIdentityMatcher(); kubeWorkloadIdentityMatcher != nil {
					namespaces := kubeWorkloadIdentityMatcher.GetNamespaces()
					clusters := kubeWorkloadIdentityMatcher.GetClusters()
					if len(namespaces) > 0 && !stringutils.ContainsString(kubeWorkload.GetController().GetNamespace(), namespaces) {
						return false
					}
					if len(clusters) > 0 && !stringutils.ContainsString(kubeWorkload.GetController().GetClusterName(), clusters) {
						return false
					}
					selectorMatches = true
					break
				}
				if kubeWorkloadRefs := selector.GetKubeServiceAccountRefs(); kubeWorkloadRefs != nil {
					for _, ref := range kubeWorkloadRefs.GetServiceAccounts() {
						if ref.GetName() == kubeWorkload.GetServiceAccountName() &&
							ref.GetNamespace() == kubeWorkload.GetController().GetNamespace() &&
							ref.GetClusterName() == kubeWorkload.GetController().GetClusterName() {
							return true
						}
					}
					return false
				}
			}
		}
	}

	return selectorMatches && meshMatches(workload.Spec.GetMesh(), matchingMeshes, matchingVirtualMeshes, meshToVirtualMesh)
}

func SelectorMatchesService(selectors []*v1alpha2.TrafficTargetSelector, service *discoveryv1alpha2.TrafficTarget) bool {
	if len(selectors) == 0 {
		return true
	}

	for _, selector := range selectors {
		kubeService := service.Spec.GetKubeService()
		if kubeService != nil {
			if kubeServiceMatcher := selector.KubeServiceMatcher; kubeServiceMatcher != nil {
				if kubeServiceMatches(
					kubeServiceMatcher.Labels,
					kubeServiceMatcher.Namespaces,
					kubeServiceMatcher.Clusters,
					kubeService,
				) {
					return true
				}
			}
			if kubeServiceRefs := selector.KubeServiceRefs; kubeServiceRefs != nil {
				if refsContain(
					kubeServiceRefs.Services,
					kubeService.Ref,
				) {
					return true
				}
			}
		}
	}

	return false
}

/* For a k8s Workload to match:
1) If labels is specified, all labels must exist on the k8s Workload
2) If namespaces is specified, the k8s workload must be in one of those namespaces
*/
func kubeWorkloadMatches(
	labels map[string]string,
	namespaces []string,
	kubeWorkload *discoveryv1alpha2.WorkloadSpec_KubernetesWorkload,
) bool {
	if len(namespaces) > 0 && !stringutils.ContainsString(kubeWorkload.GetController().GetNamespace(), namespaces) {
		return false
	}
	for k, v := range labels {
		serviceLabelValue, ok := kubeWorkload.GetPodLabels()[k]
		if !ok || serviceLabelValue != v {
			return false
		}
	}
	return true
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
	kubeService *discoveryv1alpha2.TrafficTargetSpec_KubeService,
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
	if len(clusters) > 0 && !stringutils.ContainsString(kubeService.GetRef().GetClusterName(), clusters) {
		return false
	}
	return true
}

// Returns true if the given mesh either matches one of the given matchingMeshes, or it is in a virtual mesh that
// matches one of the given matchingVirtualMeshes
func meshMatches(meshRef *v1.ObjectRef, matchingMeshes map[string]bool, matchingVirtualMeshes map[string]bool,
	meshToVirtualMesh map[string]string) bool {
	meshKey := sets.Key(meshRef)
	if matchingMeshes[meshKey] {
		return true
	}
	if virtualMeshRefKey, ok := meshToVirtualMesh[meshKey]; ok {
		return matchingVirtualMeshes[virtualMeshRefKey]
	}
	return false
}

func refsContain(refs []*v1.ClusterObjectRef, targetRef *v1.ClusterObjectRef) bool {
	for _, ref := range refs {
		if ezkube.ClusterRefsMatch(targetRef, ref) {
			return true
		}
	}
	return false
}
