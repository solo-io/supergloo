package zephyr_networking_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	zephyr_networking "github.com/solo-io/mesh-projects/pkg/clients/zephyr/networking"
	mock_controller_runtime "github.com/solo-io/mesh-projects/test/mocks/controller-runtime"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("mesh group client", func() {
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
			mockClient *mock_controller_runtime.MockClient
			mgClient   zephyr_networking.MeshGroupClient
		)

		BeforeEach(func() {
			mockClient = mock_controller_runtime.NewMockClient(ctrl)
			mgClient = zephyr_networking.NewMeshGroupClient(mockClient)
		})

		Context("get", func() {

			It("can call get with the proper args, and return err", func() {
				mg := &v1alpha1.MeshGroup{}
				resourceName := client.ObjectKey{
					Name:      "name",
					Namespace: "namespace",
				}
				mockClient.EXPECT().Get(ctx, resourceName, mg).Return(testErr)
				_, err := mgClient.Get(ctx, resourceName.Name, resourceName.Namespace)
				Expect(err).To(Equal(testErr))
			})

			It("can call get with the proper args, and return non-err", func() {
				mg := &v1alpha1.MeshGroup{}
				resourceName := client.ObjectKey{
					Name:      "name",
					Namespace: "namespace",
				}
				mockClient.EXPECT().Get(ctx, resourceName, mg).Return(nil)
				response, err := mgClient.Get(ctx, resourceName.Name, resourceName.Namespace)
				Expect(err).NotTo(HaveOccurred())
				Expect(response).NotTo(BeNil())
			})

		})

		Context("list", func() {

			It("can call list with the proper args, and return err", func() {
				mg := &v1alpha1.MeshGroupList{}
				listOptions := v1.ListOptions{}
				mockClient.EXPECT().List(ctx, mg, &client.ListOptions{Raw: &listOptions}).Return(testErr)
				_, err := mgClient.List(ctx, listOptions)
				Expect(err).To(Equal(testErr))
			})

			It("can call get with the proper args, and return non-err", func() {
				mg := &v1alpha1.MeshGroupList{}
				listOptions := v1.ListOptions{}
				mockClient.EXPECT().List(ctx, mg, &client.ListOptions{Raw: &listOptions}).Return(nil)
				response, err := mgClient.List(ctx, listOptions)
				Expect(err).NotTo(HaveOccurred())
				Expect(response).NotTo(BeNil())
			})
		})

	})
})
