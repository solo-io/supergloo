package selectorutils

import (
	"github.com/rotisserie/eris"
	commonv1 "github.com/solo-io/gloo-mesh/pkg/api/common.mesh.gloo.solo.io/v1"
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	"github.com/solo-io/go-utils/stringutils"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/solo-io/skv2/pkg/ezkube"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

// Reserved value for object namespace selection.
// If a selector contains this value in its 'namespace' field, we match objects from any namespace
const allNamespaceObjectSelector = "*"

var (
	ObjectSelectorExpressionsAndLabelsWarning = eris.New("cannot use both labels and expressions within the " +
		"same selector")
	ObjectSelectorInvalidExpressionWarning = eris.New("the object selector expression is invalid")

	// Map connecting Objects expression operator values and Kubernetes expression operator string values.
	ObjectExpressionOperatorValues = map[commonv1.ObjectSelector_Expression_Operator]selection.Operator{
		commonv1.ObjectSelector_Expression_Equals:       selection.Equals,
		commonv1.ObjectSelector_Expression_DoubleEquals: selection.DoubleEquals,
		commonv1.ObjectSelector_Expression_NotEquals:    selection.NotEquals,
		commonv1.ObjectSelector_Expression_In:           selection.In,
		commonv1.ObjectSelector_Expression_NotIn:        selection.NotIn,
		commonv1.ObjectSelector_Expression_Exists:       selection.Exists,
		commonv1.ObjectSelector_Expression_DoesNotExist: selection.DoesNotExist,
		commonv1.ObjectSelector_Expression_GreaterThan:  selection.GreaterThan,
		commonv1.ObjectSelector_Expression_LessThan:     selection.LessThan,
	}
)

func SelectorMatchesWorkload(selectors []*commonv1.WorkloadSelector, workload *discoveryv1.Workload) bool {
	if len(selectors) == 0 {
		return true
	}

	for _, selector := range selectors {
		kubeWorkload := workload.Spec.GetKubernetes()

		kubeWorkloadMatcher := selector.GetKubeWorkloadMatcher()
		if kubeWorkload != nil {
			if kubeWorkloadMatches(
				kubeWorkloadMatcher.GetLabels(),
				kubeWorkloadMatcher.GetNamespaces(),
				kubeWorkloadMatcher.GetClusters(),
				kubeWorkload,
			) {
				return true
			}
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

// used in enterprise
func ValidateSelector(
	selector *commonv1.ObjectSelector,
) error {

	if len(selector.Labels) > 0 {
		// expressions and labels cannot be both specified at the same time
		if len(selector.Expressions) > 0 {
			return ObjectSelectorExpressionsAndLabelsWarning
		}
	}

	if len(selector.Expressions) > 0 {
		for _, expression := range selector.Expressions {
			if _, err := labels.NewRequirement(
				expression.Key,
				ObjectExpressionOperatorValues[expression.Operator],
				expression.Values); err != nil {
				return eris.Wrap(ObjectSelectorInvalidExpressionWarning, err.Error())
			}
		}
	}

	return nil
}

// TODO: validate the logic contained here
func SelectorMatchesObject(
	candidate ezkube.Object,
	selector *commonv1.ObjectSelector,
	ownerNamespace string,
) bool {
	type nsSelectorType int
	const (
		// Match objects in the owner namespace
		owner nsSelectorType = iota
		// Match objects in all namespaces watched by Gloo
		all
		// Match objects in the specified namespaces
		list
	)

	nsSelector := owner
	if len(selector.Namespaces) > 0 {
		nsSelector = list
	}
	for _, ns := range selector.Namespaces {
		if ns == allNamespaceObjectSelector {
			nsSelector = all
		}
	}

	rtLabels := labels.Set(candidate.GetLabels())

	if len(selector.Labels) > 0 {
		// expressions and labels cannot be both specified at the same time
		if len(selector.Expressions) > 0 {
			return false
		}

		labelSelector := labels.SelectorFromSet(selector.Labels)

		// Check whether labels match (strict equality)
		if selector.Labels != nil {
			if !labelSelector.Matches(rtLabels) {
				return false
			}
		}

	} else if len(selector.Expressions) > 0 {
		var requirements labels.Requirements
		for _, expression := range selector.Expressions {
			r, err := labels.NewRequirement(
				expression.Key,
				ObjectExpressionOperatorValues[expression.Operator],
				expression.Values)
			if err != nil {
				return false
			}
			requirements = append(requirements, *r)
		}
		// Check whether labels match (expression requirements)
		if requirements != nil {
			if !objectLabelsMatchRequirements(requirements, rtLabels) {
				return false
			}
		}
	}

	// Check whether namespace matches
	nsMatches := false
	switch nsSelector {
	case all:
		nsMatches = true
	case owner:
		nsMatches = candidate.GetName() == ownerNamespace
	case list:
		for _, ns := range selector.Namespaces {
			if ns == candidate.GetNamespace() {
				nsMatches = true
			}
		}
	}

	return nsMatches
}

// Asserts that the object labels matches all of the expression requirements (logical AND).
func objectLabelsMatchRequirements(requirements labels.Requirements, rtLabels labels.Set) bool {
	for _, r := range requirements {
		if !r.Matches(rtLabels) {
			return false
		}
	}
	return true
}
