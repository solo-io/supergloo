package webhook

import (
	"context"

	"github.com/solo-io/supergloo/pkg/webhook/clients"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/webhook/test"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("handle AdmissionReview requests", func() {

	var (
		testData *test.ResourcesForTest
		mockClientLabelSelector,
		mockClientNamespaceSelector,
		mockClientMeshInjectDisabled,
		mockClientIstio,
		mockClientMeshNoConfigMap,
		mockClientMeshNoSelector *clients.MockWebhookResourceClient
	)

	BeforeEach(func() {
		ctrl := gomock.NewController(T)
		defer ctrl.Finish()

		RegisterSidecarInjectionHandler()

		testData = test.GetTestResources(clients.Codecs.UniversalDeserializer())
		configMap := testData.OneContOneInitContPatch.AsStruct

		mockClientLabelSelector = buildMock(ctrl, configMap, testData.AppMeshInjectEnabledLabelSelector)
		mockClientNamespaceSelector = buildMock(ctrl, configMap, testData.AppMeshInjectEnabledNamespaceSelector)
		mockClientMeshInjectDisabled = buildMock(ctrl, configMap, testData.AppMeshInjectDisabled)
		mockClientMeshNoConfigMap = buildMock(ctrl, configMap, testData.AppMeshNoConfigMap)
		mockClientMeshNoSelector = buildMock(ctrl, configMap, testData.AppMeshNoSelector)
		mockClientIstio = buildMock(ctrl, configMap, testData.IstioMesh)
	})

	It("correctly patches pod that matches injection label selector", func() {
		clients.SetClientSet(mockClientLabelSelector)

		response, err := admit(context.TODO(), testData.MatchingPod.ToRequest())
		Expect(err).NotTo(HaveOccurred())

		Expect(response.Allowed).To(BeTrue())
		pt := admissionv1beta1.PatchTypeJSONPatch
		Expect(response.PatchType).To(BeEquivalentTo(&pt))
		Expect(len(response.Patch)).NotTo(BeZero())

		patchedPod := test.GetPatchedPod(testData.MatchingPod.AsJsonString, response.Patch)

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

	It("correctly patches pod that matches injection namespace selector", func() {
		clients.SetClientSet(mockClientNamespaceSelector)

		response, err := admit(context.TODO(), testData.MatchingPod.ToRequest())
		Expect(err).NotTo(HaveOccurred())

		Expect(response.Allowed).To(BeTrue())
		pt := admissionv1beta1.PatchTypeJSONPatch
		Expect(response.PatchType).To(BeEquivalentTo(&pt))
		Expect(len(response.Patch)).NotTo(BeZero())

		patchedPod := test.GetPatchedPod(testData.MatchingPod.AsJsonString, response.Patch)

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

	It("uses the default config map if the mesh is missing the SidecarPatchConfigMap field", func() {
		clients.SetClientSet(mockClientMeshNoConfigMap)

		response, err := admit(context.TODO(), testData.MatchingPod.ToRequest())
		Expect(err).NotTo(HaveOccurred())
		Expect(response.Allowed).To(BeTrue())
		pt := admissionv1beta1.PatchTypeJSONPatch
		Expect(response.PatchType).To(BeEquivalentTo(&pt))
		Expect(len(response.Patch)).NotTo(BeZero())
	})

	It("does not patch pod that does not match injection selector", func() {
		clients.SetClientSet(mockClientLabelSelector)

		response, err := admit(context.TODO(), testData.NonMatchingPod.ToRequest())
		Expect(err).NotTo(HaveOccurred())
		Expect(response.Allowed).To(BeTrue())
		Expect(response.PatchType).To(BeNil())
		Expect(response.Patch).To(BeNil())
	})

	It("does not patch pods when auto-injection is disabled for the mesh", func() {
		clients.SetClientSet(mockClientMeshInjectDisabled)

		response, err := admit(context.TODO(), testData.MatchingPod.ToRequest())
		Expect(err).NotTo(HaveOccurred())
		Expect(response.Allowed).To(BeTrue())
		Expect(response.PatchType).To(BeNil())
		Expect(response.Patch).To(BeNil())
	})

	It("does not patch pods when mesh is not of type AWS App Mesh", func() {
		clients.SetClientSet(mockClientIstio)

		response, err := admit(context.TODO(), testData.MatchingPod.ToRequest())
		Expect(err).NotTo(HaveOccurred())
		Expect(response.Allowed).To(BeTrue())
		Expect(response.PatchType).To(BeNil())
		Expect(response.Patch).To(BeNil())
	})

	It("fails if auto-injection is enabled but the mesh is missing the InjectionSelector field", func() {
		clients.SetClientSet(mockClientMeshNoSelector)

		_, err := admit(context.TODO(), testData.MatchingPod.ToRequest())
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("auto-injection enabled but no selector for mesh"))
	})

	It("fails if the container in the candidate pod has containers that do not specify any containerPorts", func() {
		clients.SetClientSet(mockClientLabelSelector)

		_, err := admit(context.TODO(), testData.MatchingPodWithoutPorts.ToRequest())
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("no containerPorts for container"))
	})
})

func buildMock(ctrl *gomock.Controller, configMapToReturn *corev1.ConfigMap, meshesToReturn ...*v1.Mesh) *clients.MockWebhookResourceClient {
	slice := []*v1.Mesh(meshesToReturn)
	mockClient := clients.NewMockWebhookResourceClient(ctrl)
	mockClient.EXPECT().ListMeshes(gomock.Any(), gomock.Any()).Return(slice, nil).AnyTimes()
	mockClient.EXPECT().GetConfigMap("supergloo-system", "sidecar-injector").Return(configMapToReturn, nil).AnyTimes()
	mockClient.EXPECT().GetSuperglooNamespace().Return("supergloo-system", nil).AnyTimes()
	return mockClient
}
