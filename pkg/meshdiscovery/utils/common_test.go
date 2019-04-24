package utils

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	skkube "github.com/solo-io/solo-kit/pkg/api/external/kubernetes/pod"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
	kubev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("common utils for discovery", func() {

	var (
		istioNamespace = "istio-system"
	)

	var constructPod = func(container kubev1.Container, namespace string) *kubernetes.Pod {

		pod := &kubev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      "istio-pilot",
			},
			Spec: kubev1.PodSpec{
				Containers: []kubev1.Container{
					container,
				},
			},
		}
		return skkube.FromKubePod(pod)
	}

	Context("get version from pod", func() {
		It("errors when no pilot container is found", func() {
			container := kubev1.Container{
				Image: "istio-",
			}
			pod := constructPod(container, istioNamespace)
			_, err := GetVersionFromPodWithMatchers(pod, []string{"istio", "pilot"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unable to find matching container from pod"))
		})
		It("errors when no version is found in image name", func() {
			container := kubev1.Container{
				Image: "istio-pilot",
			}
			pod := constructPod(container, istioNamespace)
			_, err := GetVersionFromPodWithMatchers(pod, []string{"istio", "pilot"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unable to find image version for image"))
		})
		It("fails when image is the incorrect format", func() {
			container := kubev1.Container{
				Image: "istio-pilot:10.6",
			}
			pod := constructPod(container, istioNamespace)
			_, err := GetVersionFromPodWithMatchers(pod, []string{"istio", "pilot"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unable to find image version for image"))
		})
		It("errors when no version is found in image name", func() {
			container := kubev1.Container{
				Image: "istio-pilot:1.0.6",
			}
			pod := constructPod(container, istioNamespace)
			version, err := GetVersionFromPodWithMatchers(pod, []string{"istio", "pilot"})
			Expect(err).NotTo(HaveOccurred())
			Expect(version).To(Equal("1.0.6"))
		})
	})
})
