package utils

import (
	"fmt"
	"os"
	"strings"

	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/translator/utils"
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

func InjectedPodsByNamespace(pods v1.PodList, proxyContainerName string) v1.PodsByNamespace {
	result := make(v1.PodsByNamespace)
	for _, pod := range pods {
		if isInjectedPodRunning(pod, proxyContainerName) {
			result.Add(pod)
		}
	}
	return result
}

func isInjectedPodRunning(pod *v1.Pod, proxyContainerName string) bool {
	for _, container := range pod.Spec.Containers {
		if container.Name == proxyContainerName &&
			pod.Status.Phase == corev1.PodRunning {
			return true
		}
	}
	return false

}

func GetUpstreamsForInjectedPods(pods v1.PodList, upstreams gloov1.UpstreamList) gloov1.UpstreamList {
	var result gloov1.UpstreamList
	for _, us := range upstreams {
		podsForUpstream := utils.PodsForUpstreams(gloov1.UpstreamList{us}, pods)
		if len(podsForUpstream) > 0 {
			result = append(result, us)
		}
	}
	return result
}
