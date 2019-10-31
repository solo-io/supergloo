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
