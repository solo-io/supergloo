package selectorutils

import (
	"github.com/solo-io/go-utils/stringutils"
	discoveryv1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/solo-io/skv2/pkg/ezkube"
)

func SelectorMatchesService(selectors []*v1alpha1.ServiceSelector, service *discoveryv1alpha1.MeshService) bool {
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
			if kubeServiceRefs := selector.KubeServiceRefs; kubeServiceRefs != nil && kubeService != nil {
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

/* For a k8s Service to match:
1) If labels is specified, all labels must exist on the k8s Service
2) If namespaces is specified, the k8s must be in one of those namespaces
3) The k8s Service must exist in the specified cluster. If cluster is empty, select across all clusters.
*/
func kubeServiceMatches(
	labels map[string]string,
	namespaces []string,
	clusters []string,
	kubeService *discoveryv1alpha1.MeshServiceSpec_KubeService,
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

func refsContain(refs []*v1.ClusterObjectRef, targetRef *v1.ClusterObjectRef) bool {
	for _, ref := range refs {
		if refsEqual(targetRef, ref) {
			return true
		}
	}
	return false
}

func refsEqual(ref1, ref2 ezkube.ClusterResourceId) bool {
	return ref1.GetClusterName() == ref2.GetClusterName() &&
		ref1.GetNamespace() == ref2.GetNamespace() &&
		ref1.GetName() == ref2.GetName()
}
