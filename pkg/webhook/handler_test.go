package webhook

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/webhook/test"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ = Describe("handle AdmissionReview requests", func() {

	var (
		mockClient,
		mockClientMeshInjectDisabled,
		mockClientIstio,
		mockClientMeshNoConfigMap,
		mockClientMeshNoSelector *MockwebhookResourceClient
		pod       *corev1.Pod
		configMap *corev1.ConfigMap
	)

	BeforeEach(func() {
		ctrl := gomock.NewController(T)
		defer ctrl.Finish()

		deserializer := Codecs.UniversalDeserializer()

		pod = &corev1.Pod{}
		_, _, err := deserializer.Decode([]byte(test.MatchingPod), nil, pod)
		Expect(err).NotTo(HaveOccurred())

		configMap = &corev1.ConfigMap{}
		_, _, err = deserializer.Decode([]byte(test.ConfigMap), nil, configMap)
		Expect(err).NotTo(HaveOccurred())

		mockClient = NewMockwebhookResourceClient(ctrl)
		mockClient.EXPECT().ListMeshes(gomock.Any(), gomock.Any()).Return(v1.MeshList{test.AppMeshInjectEnabled}, nil).AnyTimes()
		mockClient.EXPECT().GetConfigMap(gomock.Any(), gomock.Any()).Return(configMap, nil).AnyTimes()

		mockClientMeshInjectDisabled = NewMockwebhookResourceClient(ctrl)
		mockClientMeshInjectDisabled.EXPECT().ListMeshes(gomock.Any(), gomock.Any()).Return(v1.MeshList{test.AppMeshInjectDisabled}, nil).AnyTimes()
		mockClientMeshInjectDisabled.EXPECT().GetConfigMap(gomock.Any(), gomock.Any()).Return(configMap, nil).AnyTimes()

		mockClientIstio = NewMockwebhookResourceClient(ctrl)
		mockClientIstio.EXPECT().ListMeshes(gomock.Any(), gomock.Any()).Return(v1.MeshList{test.IstioMesh}, nil).AnyTimes()
		mockClientIstio.EXPECT().GetConfigMap(gomock.Any(), gomock.Any()).Return(configMap, nil).AnyTimes()

		mockClientMeshNoConfigMap = NewMockwebhookResourceClient(ctrl)
		mockClientMeshNoConfigMap.EXPECT().ListMeshes(gomock.Any(), gomock.Any()).Return(v1.MeshList{test.AppMeshNoConfigMap}, nil).AnyTimes()
		mockClientMeshNoConfigMap.EXPECT().GetConfigMap(gomock.Any(), gomock.Any()).Return(configMap, nil).AnyTimes()

		mockClientMeshNoSelector = NewMockwebhookResourceClient(ctrl)
		mockClientMeshNoSelector.EXPECT().ListMeshes(gomock.Any(), gomock.Any()).Return(v1.MeshList{test.AppMeshNoSelector}, nil).AnyTimes()
		mockClientMeshNoSelector.EXPECT().GetConfigMap(gomock.Any(), gomock.Any()).Return(configMap, nil).AnyTimes()
	})

	It("correctly patches pod that matches injection selector", func() {
		setClientSet(mockClient)

		response, err := admit(context.TODO(), buildRequest(test.MatchingPod))
		Expect(err).NotTo(HaveOccurred())

		Expect(response.Allowed).To(BeTrue())
		pt := admissionv1beta1.PatchTypeJSONPatch
		Expect(response.PatchType).To(BeEquivalentTo(&pt))
		Expect(len(response.Patch)).NotTo(BeZero())

		patchedPod := test.GetPatchedPod(test.MatchingPod, response.Patch)

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

	It("does not patch pod that does not match injection selector", func() {
		setClientSet(mockClient)

		response, err := admit(context.TODO(), buildRequest(test.NonMatchingPod))
		Expect(err).NotTo(HaveOccurred())
		Expect(response.Allowed).To(BeTrue())
		Expect(response.PatchType).To(BeNil())
		Expect(response.Patch).To(BeNil())
	})

	It("does not patch pods when auto-injection is disabled for the mesh", func() {
		setClientSet(mockClientMeshInjectDisabled)

		response, err := admit(context.TODO(), buildRequest(test.MatchingPod))
		Expect(err).NotTo(HaveOccurred())
		Expect(response.Allowed).To(BeTrue())
		Expect(response.PatchType).To(BeNil())
		Expect(response.Patch).To(BeNil())
	})

	It("does not patch pods when mesh is not of type AWS App Mesh", func() {
		setClientSet(mockClientIstio)

		response, err := admit(context.TODO(), buildRequest(test.MatchingPod))
		Expect(err).NotTo(HaveOccurred())
		Expect(response.Allowed).To(BeTrue())
		Expect(response.PatchType).To(BeNil())
		Expect(response.Patch).To(BeNil())
	})

	It("fails if auto-injection is enabled but the mesh is missing the SidecarPatchConfigMap field", func() {
		setClientSet(mockClientMeshNoConfigMap)

		_, err := admit(context.TODO(), buildRequest(test.MatchingPod))
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("SidecarPatchConfigMap is nil for mesh"))

	})

	It("fails if auto-injection is enabled but the mesh is missing the InjectionSelector field", func() {
		setClientSet(mockClientMeshNoSelector)

		_, err := admit(context.TODO(), buildRequest(test.MatchingPod))
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("auto-injection enabled but no selector for mesh"))
	})

})

func buildRequest(pod string) admissionv1beta1.AdmissionReview {
	return admissionv1beta1.AdmissionReview{
		Request: &admissionv1beta1.AdmissionRequest{
			Resource: metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"},
			Object: runtime.RawExtension{
				Raw: []byte(pod),
			},
		},
	}
}
