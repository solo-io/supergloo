package utils

import (
	"github.com/pkg/errors"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func UpstreamsForSelector(selector *v1.PodSelector, allUpstreams gloov1.UpstreamList) (gloov1.UpstreamList, error) {
	if selector == nil {
		return allUpstreams, nil
	}
	var selectedUpstreams gloov1.UpstreamList

	switch selector := selector.SelectorType.(type) {
	case *v1.PodSelector_LabelSelector_:
		// true if an upstream exists whose selector falls within the rr's selector
		// and the host in question is that upstream's host
		for _, us := range allUpstreams {

			upstreamLabels := GetLabelsForUpstream(us)
			labelsMatch := labels.SelectorFromSet(selector.LabelSelector.LabelsToMatch).Matches(labels.Set(upstreamLabels))
			if !labelsMatch {
				continue
			}

			selectedUpstreams = append(selectedUpstreams, us)
		}

	case *v1.PodSelector_UpstreamSelector_:
		for _, ref := range selector.UpstreamSelector.Upstreams {
			us, err := allUpstreams.Find(ref.Strings())
			if err != nil {
				return nil, err
			}
			selectedUpstreams = append(selectedUpstreams, us)
		}
	case *v1.PodSelector_NamespaceSelector_:
		for _, us := range allUpstreams {
			namespaceForUpstream := GetNamespaceForUpstream(us)
			var inSelectedNamespace bool
			for _, ns := range selector.NamespaceSelector.Namespaces {
				if ns == namespaceForUpstream {
					inSelectedNamespace = true
					break
				}
			}
			if !inSelectedNamespace {
				continue
			}

			selectedUpstreams = append(selectedUpstreams, us)
		}
	}
	return selectedUpstreams, nil
}

func PodsForSelector(selector *v1.PodSelector, upstreams gloov1.UpstreamList, allPods kubernetes.PodList) (kubernetes.PodList, error) {
	if selector == nil {
		return allPods, nil
	}
	var selectedPods kubernetes.PodList

	switch selectorType := selector.SelectorType.(type) {
	case *v1.PodSelector_LabelSelector_:
		// true if an upstream exists whose selector falls within the rr's selector
		// and the host in question is that upstream's host
		for _, pod := range allPods {

			upstreamLabels := pod.Labels
			labelsMatch := labels.SelectorFromSet(selectorType.LabelSelector.LabelsToMatch).Matches(labels.Set(upstreamLabels))
			if !labelsMatch {
				continue
			}

			selectedPods = append(selectedPods, pod)
		}

	case *v1.PodSelector_UpstreamSelector_:
		selectedUpstreams, err := UpstreamsForSelector(selector, upstreams)
		if err != nil {
			return nil, errors.Wrap(err, "getting upstreams for pods")
		}
		return PodsForUpstreams(selectedUpstreams, allPods), nil
	case *v1.PodSelector_NamespaceSelector_:
		for _, pod := range allPods {
			var podInSelectedNamespace bool
			for _, ns := range selectorType.NamespaceSelector.Namespaces {
				namespaceForUpstream := pod.Namespace
				if ns == namespaceForUpstream {
					podInSelectedNamespace = true
					break
				}
			}
			if !podInSelectedNamespace {
				continue
			}

			selectedPods = append(selectedPods, pod)
		}
	}
	return selectedPods, nil
}

type namespacedSelector struct {
	namespace string
	selector  map[string]string
}

func PodsForUpstreams(upstreams gloov1.UpstreamList, allPods kubernetes.PodList) kubernetes.PodList {
	var selectedPods kubernetes.PodList
	var selectors []namespacedSelector
	for _, us := range upstreams {
		kubeUs, ok := us.UpstreamSpec.UpstreamType.(*gloov1.UpstreamSpec_Kube)
		if !ok {
			continue
		}

		// This upstream refers to a service without selectors. These services are used to abstract backends which are
		// not pods. Skip it because otherwise the nil selector will match any pod in the loop below.
		// (see https://kubernetes.io/docs/concepts/services-networking/service/#services-without-selectors)
		if kubeUs.Kube.Selector == nil {
			continue
		}

		selectors = append(selectors, namespacedSelector{namespace: kubeUs.Kube.ServiceNamespace, selector: kubeUs.Kube.Selector})
	}
	for _, pod := range allPods {
		var includedInSelector bool
		for _, selector := range selectors {
			if pod.Namespace != selector.namespace {
				continue
			}
			if labels.SelectorFromSet(selector.selector).Matches(labels.Set(pod.Labels)) {
				includedInSelector = true
				break
			}
		}
		if includedInSelector {
			selectedPods = append(selectedPods, pod)
		}
	}
	return selectedPods
}
