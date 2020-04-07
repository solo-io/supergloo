package kubernetes_apps_test

import (
	"context"
	"errors"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	kubernetes_apps "github.com/solo-io/service-mesh-hub/pkg/clients/kubernetes/apps"
	mock_controller_runtime "github.com/solo-io/service-mesh-hub/test/mocks/controller-runtime"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("DeploymentClient", func() {
	var (
		deploymentClient kubernetes_apps.DeploymentClient
		mockKubeClient   *mock_controller_runtime.MockClient
		ctx              context.Context
		ctrl             *gomock.Controller
	)
	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockKubeClient = mock_controller_runtime.NewMockClient(ctrl)
		deploymentClient = kubernetes_apps.NewControllerRuntimeDeploymentClient(mockKubeClient)
		ctx = context.TODO()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should get deployment", func() {
		objectKey := client.ObjectKey{Namespace: "foo", Name: "bar"}
		mockKubeClient.EXPECT().Get(ctx, objectKey, &appsv1.Deployment{}).Return(nil)
		_, err := deploymentClient.Get(ctx, objectKey)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should return error", func() {
		namespace := "foo"
		name := "bar"
		objectKey := client.ObjectKey{Namespace: namespace, Name: name}
		mockKubeClient.EXPECT().Get(ctx, objectKey, &appsv1.Deployment{}).Return(errors.New(""))
		_, err := deploymentClient.Get(ctx, objectKey)
		Expect(err).To(HaveOccurred())
	})
})
