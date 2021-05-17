package selectorutils

import (
	"context"

	commonv1 "github.com/solo-io/gloo-mesh/pkg/api/common.mesh.gloo.solo.io/v1"
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/stringutils"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/solo-io/skv2/pkg/ezkube"
)

func SelectorMatchesWorkload(ctx context.Context, selectors []*commonv1.WorkloadSelector, workload *discoveryv1.Workload) bool {
	if len(selectors) == 0 {
		return true
	}

	kubeWorkload := workload.Spec.GetKubernetes()
	if kubeWorkload == nil {
		contextutils.LoggerFrom(ctx).DPanicf("Missing Kubernetes workload data for Workload %s", sets.Key(workload))
		return false
	}

	for _, selector := range selectors {

		// matcher-based selection
		matcher := selector.GetKubeWorkloadMatcher()
		if kubeWorkloadMatches(matcher.GetLabels(), matcher.GetNamespaces(), matcher.GetClusters(), kubeWorkload) {
			return true
		}
	}

	return false
}

func IdentityMatchesWorkload(selectors []*commonv1.IdentitySelector, workload *discoveryv1.Workload) bool {
	if len(selectors) == 0 {
		return true
	}

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
				return true
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

	return false
}

func SelectorMatchesDestination(selectors []*commonv1.DestinationSelector, destination *discoveryv1.Destination) bool {
	if len(selectors) == 0 {
		return true
	}

	for _, selector := range selectors {
		kubeService := destination.Spec.GetKubeService()
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

// Return true if any WorkloadSelector selects the specified clusterName
func WorkloadSelectorContainsCluster(selectors []*commonv1.WorkloadSelector, clusterName string) bool {
	if len(selectors) == 0 {
		return true
	}

	for _, selector := range selectors {
		clusters := selector.GetKubeWorkloadMatcher().Clusters

		if len(clusters) == 0 || stringutils.ContainsString(clusterName, clusters) {
			return true
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
	clusters []string,
	kubeWorkload *discoveryv1.WorkloadSpec_KubernetesWorkload,
) bool {
	if len(namespaces) > 0 && !stringutils.ContainsString(kubeWorkload.GetController().GetNamespace(), namespaces) {
		return false
	}
	if len(clusters) > 0 && !stringutils.ContainsString(kubeWorkload.GetController().GetClusterName(), clusters) {
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
	kubeService *discoveryv1.DestinationSpec_KubeService,
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
		if ezkube.ClusterRefsMatch(targetRef, ref) {
			return true
		}
	}
	return false
}
