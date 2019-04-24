package utils

import (
	"regexp"

	"github.com/solo-io/go-utils/errors"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

func ImageVersion(image string) (string, error) {
	regex := regexp.MustCompile("([0-9]+[.][0-9]+[.][0-9]+$)")
	imageTag := regex.FindString(image)
	if imageTag == "" {
		return "", errors.Errorf("unable to find image version for image: %s", image)
	}
	return imageTag, nil
}

func SelectRunningPods(pods v1.PodList) v1.PodList {
	var result v1.PodList
	for _, pod := range pods {
		podRunning := true
		for _, stat := range pod.Status.ContainerStatuses {
			if stat.State.Running == nil {
				podRunning = false
				break
			}
		}
		if podRunning {
			result = append(result, pod)
		}
	}
	return result
}
