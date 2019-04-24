package utils

import (
	"regexp"
	"strings"

	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
)

func ImageVersion(image string) (string, error) {
	regex := regexp.MustCompile("([0-9]+[.][0-9]+[.][0-9]+$)")
	imageTag := regex.FindString(image)
	if imageTag == "" {
		return "", errors.Errorf("unable to find image version for image: %s", image)
	}
	return imageTag, nil
}

func FilerPodsByNamePrefix(pods kubernetes.PodList, namePrefix string) kubernetes.PodList {
	var result kubernetes.PodList
	for _, pod := range pods {
		if strings.Contains(pod.Name, namePrefix) {
			result = append(result, pod)
		}
	}
	return result
}
