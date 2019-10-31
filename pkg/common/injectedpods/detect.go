package injectedpods

import (
	"context"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
)

type DetectInjectedPod func(ctx context.Context, pod *kubernetes.Pod) (cluster, namespace string, isInjected bool)

type InjectedPods map[string]map[string]kubernetes.PodList

type detector struct {
	DetectInjectedPod
}

func NewDetector(f DetectInjectedPod) *detector {
	return &detector{DetectInjectedPod: f}
}

func (d detector) DetectInjectedPods(ctx context.Context, pods kubernetes.PodList) InjectedPods {
	injectedPods := make(map[string]map[string]kubernetes.PodList)
	for _, pod := range pods {
		clusterName, discoveryNamespace, ok := d.DetectInjectedPod(ctx, pod)
		if ok {
			if injectedPods[clusterName] == nil {
				injectedPods[clusterName] = make(map[string]kubernetes.PodList)
			}
			injectedPods[clusterName][discoveryNamespace] = append(injectedPods[clusterName][discoveryNamespace], pod)
		}
	}
	return injectedPods
}
