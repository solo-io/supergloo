package utils

import (
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
	"k8s.io/apimachinery/pkg/labels"
)

func UpstreamsForPods(pods kubernetes.PodList, allUpstreams gloov1.UpstreamList) gloov1.UpstreamList {
	var upstreamsForPods gloov1.UpstreamList
	for _, us := range allUpstreams {
		kubeUs, ok := us.UpstreamSpec.UpstreamType.(*gloov1.UpstreamSpec_Kube)
		if !ok {
			continue
		}

		selector := kubeUs.Kube.Selector
		// This upstream refers to a service without selectors. These services are used to abstract backends which are
		// not pods. Skip it because otherwise the nil selector will match any pod in the loop below.
		// (see https://kubernetes.io/docs/concepts/services-networking/service/#services-without-selectors)
		if selector == nil {
			continue
		}

		for _, pod := range pods {
			if kubeUs.Kube.ServiceNamespace == pod.Namespace && labels.SelectorFromSet(selector).Matches(labels.Set(pod.Labels)) {
				upstreamsForPods = append(upstreamsForPods, us)
				break
			}
		}
	}

	return upstreamsForPods
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
