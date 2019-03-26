package webhook

import (
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/webhook/test"

	corev1 "k8s.io/api/core/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("create pod sidecar patches", func() {

	var (
		pod *corev1.Pod
		configMap,
		emptyPatchConfigMap,
		twoEntryConfigMap,
		noContainerConfigMap,
		noInitContainerConfigMap *corev1.ConfigMap
		mesh *v1.Mesh
	)

	BeforeEach(func() {
		deserializer := Codecs.UniversalDeserializer()
		pod = &corev1.Pod{}
		_, _, err := deserializer.Decode([]byte(test.MatchingPod), nil, pod)
		Expect(err).NotTo(HaveOccurred())

		configMap = &corev1.ConfigMap{}
		_, _, err = deserializer.Decode([]byte(test.ConfigMap), nil, configMap)
		Expect(err).NotTo(HaveOccurred())

		noContainerConfigMap = &corev1.ConfigMap{}
		_, _, err = deserializer.Decode([]byte(test.NoContainerPatch), nil, noContainerConfigMap)
		Expect(err).NotTo(HaveOccurred())

		noInitContainerConfigMap = &corev1.ConfigMap{}
		_, _, err = deserializer.Decode([]byte(test.NoInitContainerPatch), nil, noInitContainerConfigMap)
		Expect(err).NotTo(HaveOccurred())

		emptyPatchConfigMap = &corev1.ConfigMap{}
		_, _, err = deserializer.Decode([]byte(test.EmptyPatch), nil, emptyPatchConfigMap)
		Expect(err).NotTo(HaveOccurred())

		twoEntryConfigMap = &corev1.ConfigMap{}
		_, _, err = deserializer.Decode([]byte(test.TwoEntryPatch), nil, emptyPatchConfigMap)
		Expect(err).NotTo(HaveOccurred())

		mesh = test.AppMeshInjectEnabled
	})

	It("correctly creates a patch adding one container and one initContainer", func() {

		patchBytes, err := buildSidecarPatch(pod, configMap, mesh)
		Expect(err).NotTo(HaveOccurred())

		patchedPod := test.GetPatchedPod(test.MatchingPod, patchBytes)

		// Check containers
		Expect(patchedPod.Spec.Containers).To(HaveLen(2))
		envoyContainer := patchedPod.Spec.Containers[1]
		Expect(envoyContainer.Name).To(BeEquivalentTo("envoy"))
		Expect(envoyContainer.Image).To(BeEquivalentTo("111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-envoy:v1.8.0.2-beta"))
		Expect(*envoyContainer.SecurityContext.RunAsUser).To(BeEquivalentTo(int64(1337)))
		Expect(envoyContainer.Env[0]).To(BeEquivalentTo(corev1.EnvVar{
			Name:  "APPMESH_VIRTUAL_NODE_NAME",
			Value: "mesh/test-mesh/virtualNode/testrunner-vn",
		}))
		Expect(envoyContainer.Env[2]).To(BeEquivalentTo(corev1.EnvVar{
			Name:  "AWS_REGION",
			Value: "us-east-1",
		}))

		// Check initContainers
		Expect(patchedPod.Spec.InitContainers).To(HaveLen(1))
		initContainer := patchedPod.Spec.InitContainers[0]
		Expect(initContainer.Name).To(BeEquivalentTo("proxyinit"))
		Expect(initContainer.Image).To(BeEquivalentTo("111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-proxy-route-manager:latest"))
		Expect((*initContainer.SecurityContext.Capabilities).Add[0]).To(BeEquivalentTo("NET_ADMIN"))
		Expect(initContainer.Env[4]).To(BeEquivalentTo(corev1.EnvVar{
			Name:  "APPMESH_APP_PORTS",
			Value: "1234",
		}))
	})

	It("correctly creates a patch adding one initContainer", func() {
		patchBytes, err := buildSidecarPatch(pod, noContainerConfigMap, mesh)
		Expect(err).NotTo(HaveOccurred())

		patchedPod := test.GetPatchedPod(test.MatchingPod, patchBytes)

		// Check containers
		Expect(patchedPod.Spec.Containers).To(HaveLen(1))
		envoyContainer := patchedPod.Spec.Containers[0]
		Expect(envoyContainer.Name).To(BeEquivalentTo("testrunner"))

		// Check initContainers
		Expect(patchedPod.Spec.InitContainers).To(HaveLen(1))
		initContainer := patchedPod.Spec.InitContainers[0]
		Expect(initContainer.Name).To(BeEquivalentTo("proxyinit"))
		Expect(initContainer.Image).To(BeEquivalentTo("111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-proxy-route-manager:latest"))
	})

	It("correctly creates a patch adding one container", func() {
		patchBytes, err := buildSidecarPatch(pod, noInitContainerConfigMap, mesh)
		Expect(err).NotTo(HaveOccurred())

		patchedPod := test.GetPatchedPod(test.MatchingPod, patchBytes)

		// Check containers
		Expect(patchedPod.Spec.Containers).To(HaveLen(2))
		envoyContainer := patchedPod.Spec.Containers[1]
		Expect(envoyContainer.Name).To(BeEquivalentTo("envoy"))
		Expect(envoyContainer.Image).To(BeEquivalentTo("111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-envoy:v1.8.0.2-beta"))

		// Check initContainers
		Expect(patchedPod.Spec.InitContainers).To(HaveLen(0))
	})

	It("fails if the config map does not contain any data", func() {
		_, err := buildSidecarPatch(pod, emptyPatchConfigMap, mesh)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("expected exactly 1 entry in config map"))
	})

	It("fails if the config map contains more than one data entry", func() {
		_, err := buildSidecarPatch(pod, twoEntryConfigMap, mesh)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("expected exactly 1 entry in config map"))
	})
})
