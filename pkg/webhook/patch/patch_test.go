package patch

import (
	"encoding/json"

	"github.com/solo-io/supergloo/pkg/webhook/clients"
	"github.com/solo-io/supergloo/pkg/webhook/test"

	corev1 "k8s.io/api/core/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("create pod sidecar patches", func() {

	var testData *test.ResourcesForTest

	BeforeEach(func() {
		testData = test.GetTestResources(clients.Codecs.UniversalDeserializer())
	})

	It("correctly creates a patch adding one container and one initContainer", func() {
		patchOperations, err := BuildSidecarPatch(
			testData.MatchingPod.AsStruct,
			testData.OneContOneInitContPatch.AsStruct,
			testData.TemplateData)
		Expect(err).NotTo(HaveOccurred())

		patchBytes, err := json.Marshal(patchOperations)
		Expect(err).NotTo(HaveOccurred())

		patchedPod := test.GetPatchedPod(testData.MatchingPod.AsJsonString, patchBytes)

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
		patchOperations, err := BuildSidecarPatch(testData.MatchingPod.AsStruct, testData.NoContainerPatch.AsStruct, testData.TemplateData)
		Expect(err).NotTo(HaveOccurred())

		patchBytes, err := json.Marshal(patchOperations)
		Expect(err).NotTo(HaveOccurred())

		patchedPod := test.GetPatchedPod(testData.MatchingPod.AsJsonString, patchBytes)

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
		patchOperations, err := BuildSidecarPatch(testData.MatchingPod.AsStruct, testData.NoInitContainerPatch.AsStruct, testData.TemplateData)
		Expect(err).NotTo(HaveOccurred())

		patchBytes, err := json.Marshal(patchOperations)
		Expect(err).NotTo(HaveOccurred())

		patchedPod := test.GetPatchedPod(testData.MatchingPod.AsJsonString, patchBytes)

		// Check containers
		Expect(patchedPod.Spec.Containers).To(HaveLen(2))
		envoyContainer := patchedPod.Spec.Containers[1]
		Expect(envoyContainer.Name).To(BeEquivalentTo("envoy"))
		Expect(envoyContainer.Image).To(BeEquivalentTo("111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-envoy:v1.8.0.2-beta"))

		// Check initContainers
		Expect(patchedPod.Spec.InitContainers).To(HaveLen(0))
	})

	It("fails if the config map does not contain any data", func() {
		_, err := BuildSidecarPatch(testData.MatchingPod.AsStruct, testData.EmptyPatch.AsStruct, testData.TemplateData)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("expected exactly 1 entry in config map"))
	})

	It("fails if the config map contains more than one data entry", func() {
		_, err := BuildSidecarPatch(testData.MatchingPod.AsStruct, testData.TwoEntryPatch.AsStruct, testData.TemplateData)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("expected exactly 1 entry in config map"))
	})
})
