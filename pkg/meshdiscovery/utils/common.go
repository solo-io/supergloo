package utils

import (
	"fmt"
	"os"
	"strings"

	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/translator/utils"
	corev1 "k8s.io/api/core/v1"
)

const (
	SelectorDiscoveredByPrefix = "discovered_by"
	SelectorCreatedByPrefix    = "created_by"
	SelectorCreatedByValue     = "mesh-discovery"
)

func MeshWriteNamespace() string {
	if writeNamespace := os.Getenv("POD_NAMESPACE"); writeNamespace != "" {
		return writeNamespace
	}
	return "supergloo-system"
}

func GetVersionFromPodWithMatchers(pod *kubernetes.Pod, podStringMatchers []string) (string, error) {
	containers := pod.Spec.Containers
	for _, container := range containers {
		if StringContainsAll(podStringMatchers, container.Image) {
			return ImageVersion(container.Image)
		}
	}
	return "", errors.Errorf("unable to find matching container from pod")
}

func StringContainsAll(podStringMatchers []string, matchString string) bool {
	for _, substr := range podStringMatchers {
		if !strings.Contains(matchString, substr) {
			return false
		}
	}
	return true
}

func BasicMeshInfo(meshNamespace string, discoverySelector map[string]string, meshType string) *v1.Mesh {
	mesh := &v1.Mesh{
		Metadata: core.Metadata{
			Namespace: MeshWriteNamespace(),
			Name:      fmt.Sprintf("%s-%s", meshType, meshNamespace),
			Labels:    discoverySelector,
		},
	}
	return mesh
}

func InjectedPodsByNamespace(pods kubernetes.PodList, proxyContainerName string) kubernetes.PodsByNamespace {
	result := make(kubernetes.PodsByNamespace)
	for _, pod := range pods {
		if isInjectedPodRunning(pod, proxyContainerName) {
			result.Add(pod)
		}
	}
	return result
}

func isInjectedPodRunning(pod *kubernetes.Pod, proxyContainerName string) bool {
	for _, container := range pod.Spec.Containers {
		if container.Name == proxyContainerName &&
			pod.Status.Phase == corev1.PodRunning {
			return true
		}
	}
	return false

}

func GetUpstreamsForInjectedPods(pods kubernetes.PodList, upstreams gloov1.UpstreamList) gloov1.UpstreamList {
	var result gloov1.UpstreamList
	for _, us := range upstreams {
		podsForUpstream := utils.PodsForUpstreams(gloov1.UpstreamList{us}, pods)
		if len(podsForUpstream) > 0 {
			result = append(result, us)
		}
	}
	return result
}
