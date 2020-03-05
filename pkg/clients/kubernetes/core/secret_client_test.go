package kubernetes_core_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	kubernetes_core "github.com/solo-io/mesh-projects/pkg/clients/kubernetes/core"
	mock_controller_runtime "github.com/solo-io/mesh-projects/test/mocks/controller-runtime"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("secrets client", func() {
	var (
		ctrl *gomock.Controller
		ctx  context.Context

		testErr = eris.New("hello")
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("client impl", func() {
		var (
			mockClient   *mock_controller_runtime.MockClient
			secretClient kubernetes_core.SecretsClient
		)

		BeforeEach(func() {
			mockClient = mock_controller_runtime.NewMockClient(ctrl)
			secretClient = kubernetes_core.NewSecretsClient(mockClient)
		})

		It("can call update with the proper args", func() {
			secret := &corev1.Secret{}
			mockClient.EXPECT().Update(ctx, secret).Return(testErr)
			Expect(secretClient.Update(ctx, secret)).To(Equal(testErr))
		})

		Context("get", func() {

			It("can call get with the proper args, and return err", func() {
				secret := &corev1.Secret{}
				resourceName := client.ObjectKey{
					Name:      "name",
					Namespace: "namespace",
				}
				mockClient.EXPECT().Get(ctx, resourceName, secret).Return(testErr)
				_, err := secretClient.Get(ctx, resourceName.Name, resourceName.Namespace)
				Expect(err).To(Equal(testErr))
			})

			It("can call get with the proper args, and return non-err", func() {
				secret := &corev1.Secret{}
				resourceName := client.ObjectKey{
					Name:      "name",
					Namespace: "namespace",
				}
				mockClient.EXPECT().Get(ctx, resourceName, secret).Return(nil)
				response, err := secretClient.Get(ctx, resourceName.Name, resourceName.Namespace)
				Expect(err).NotTo(HaveOccurred())
				Expect(response).NotTo(BeNil())
			})

		})

		Context("list", func() {

			It("can call list with the proper args, and return err", func() {
				secret := &corev1.SecretList{}
				listOptions := v1.ListOptions{}
				mockClient.EXPECT().List(ctx, secret, &client.ListOptions{Raw: &listOptions}).Return(testErr)
				_, err := secretClient.List(ctx, listOptions)
				Expect(err).To(Equal(testErr))
			})

			It("can call get with the proper args, and return non-err", func() {
				secret := &corev1.SecretList{}
				listOptions := v1.ListOptions{}
				mockClient.EXPECT().List(ctx, secret, &client.ListOptions{Raw: &listOptions}).Return(nil)
				response, err := secretClient.List(ctx, listOptions)
				Expect(err).NotTo(HaveOccurred())
				Expect(response).NotTo(BeNil())
			})
		})

	})
})
