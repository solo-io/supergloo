package utils

import (
	"fmt"
	"os"
	"strings"

	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	corev1 "k8s.io/api/core/v1"
)

const SelectorPrefix = "discovered_by"

func MeshWriteNamespace() string {
	if writeNamespace := os.Getenv("POD_NAMESPACE"); writeNamespace != "" {
		return writeNamespace
	}
	return "supergloo-system"
}

func GetVersionFromPodWithMatchers(pod *v1.Pod, podStringMatchers []string) (string, error) {
	containers := pod.Spec.Containers
	for _, container := range containers {
		if stringContainsAll(podStringMatchers, container.Image) {
			return ImageVersion(container.Image)
		}
	}
	return "", errors.Errorf("unable to find matching container from pod")
}

func stringContainsAll(podStringMatchers []string, matchString string) bool {
	for _, substr := range podStringMatchers {
		if !strings.Contains(matchString, substr) {
			return false
		}
	}
	return true
}

func BasicMeshInfo(mainPod *v1.Pod, discoverySelector map[string]string, meshType string) *v1.Mesh {
	mesh := &v1.Mesh{
		Metadata: core.Metadata{
			Namespace: MeshWriteNamespace(),
			Name:      fmt.Sprintf("%s-%s", meshType, mainPod.Namespace),
			Labels:    discoverySelector,
		},
	}
	return mesh
}

type NamespaceListFilterFunc = func(namespace *v1.KubeNamespace) bool

func FilterNamespaces(namespaces v1.KubeNamespaceList, filterFunc NamespaceListFilterFunc) v1.KubeNamespaceList {
	var result v1.KubeNamespaceList
	for _, namespace := range namespaces {
		if filterFunc(namespace) {
			result = append(result, namespace)
		}
	}
	return result
}

type InjectedPodFilterFunc = func(pod *v1.Pod, namespace *v1.KubeNamespace) bool

// TODO(EItanya): figure out a heuristic for when a singular pod has been injected
func GetInjectedPods(namespaces v1.KubeNamespaceList, pods v1.PodList,
	filterFunc InjectedPodFilterFunc) v1.PodsByNamespace {
	result := make(v1.PodsByNamespace)
	for _, namespace := range namespaces {
		result[namespace.Name] = v1.PodList{}
		for _, pod := range pods {
			if pod.Namespace == namespace.Name && filterFunc(pod, namespace) {
				result.Add(pod)
			}
		}

	}
	return result
}

func InjectedPodsByProxyContainerName(proxyContainerName string) InjectedPodFilterFunc {
	return func(pod *v1.Pod, namespace *v1.KubeNamespace) bool {
		for _, container := range pod.Spec.Containers {
			if container.Name == proxyContainerName &&
				pod.Status.Phase == corev1.PodRunning {
				return true
			}
		}
		return false
	}
}

func InjectionNamespaceSelector(injectedPods v1.PodsByNamespace) *v1.PodSelector {
	var namespaces []string
	for namespace := range injectedPods {
		namespaces = append(namespaces, namespace)
	}
	// Assume all pods in namespace have been injected,
	// so use pod namespace selector to get upstreams
	return &v1.PodSelector{
		SelectorType: &v1.PodSelector_NamespaceSelector_{
			NamespaceSelector: &v1.PodSelector_NamespaceSelector{
				Namespaces: namespaces,
			},
		},
	}
}
