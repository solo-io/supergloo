package utils

import (
	"fmt"
	"os"
	"strings"

	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
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
