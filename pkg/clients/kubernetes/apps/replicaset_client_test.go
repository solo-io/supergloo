package kubernetes_apps_test

import (
	"context"
	"errors"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	kubernetes_apps "github.com/solo-io/mesh-projects/pkg/clients/kubernetes/apps"
	mock_controller_runtime "github.com/solo-io/mesh-projects/test/mocks/controller-runtime"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("ReplicaSetClient", func() {
	var (
		replicaSetClient kubernetes_apps.ReplicaSetClient
		mockKubeClient   *mock_controller_runtime.MockClient
		ctx              context.Context
		ctrl             *gomock.Controller
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockKubeClient = mock_controller_runtime.NewMockClient(ctrl)
		replicaSetClient = kubernetes_apps.NewReplicaSetClient(mockKubeClient)
		ctx = context.TODO()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should get replicaSet", func() {
		objectKey := client.ObjectKey{Namespace: "foo", Name: "bar"}
		mockKubeClient.EXPECT().Get(ctx, objectKey, &appsv1.ReplicaSet{}).Return(nil)
		_, err := replicaSetClient.Get(ctx, objectKey)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should return error", func() {
		namespace := "foo"
		name := "bar"
		objectKey := client.ObjectKey{Namespace: namespace, Name: name}
		mockKubeClient.EXPECT().Get(ctx, objectKey, &appsv1.ReplicaSet{}).Return(errors.New(""))
		_, err := replicaSetClient.Get(ctx, objectKey)
		Expect(err).To(HaveOccurred())
	})
})
