package test

import (
	jsonpatch "github.com/evanphx/json-patch"
	"github.com/ghodss/yaml"
	corev1 "k8s.io/api/core/v1"

	. "github.com/onsi/gomega"
)

func GetPatchedPod(podYamlString string, patchBytes []byte) *corev1.Pod {
	patch, err := jsonpatch.DecodePatch(patchBytes)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	podJson, err := yaml.YAMLToJSON([]byte(podYamlString))
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	patchedPodBytes, err := patch.Apply(podJson)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	patchedPod := &corev1.Pod{}
	ExpectWithOffset(1, yaml.Unmarshal(patchedPodBytes, patchedPod)).NotTo(HaveOccurred())
	return patchedPod
}
