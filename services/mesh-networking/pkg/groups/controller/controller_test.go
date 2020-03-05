package group_controller_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	mg_controller "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1/controller"
	"github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	group_controller "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/groups/controller"
	mock_group_controller "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/groups/controller/mocks"
)

var _ = Describe("controller", func() {
	var (
		ctrl      *gomock.Controller
		validator *mock_group_controller.MockMeshGroupValidator
		handler   mg_controller.MeshGroupEventHandler
		ctx       context.Context

		testErr = eris.New("hello")
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		validator = mock_group_controller.NewMockMeshGroupValidator(ctrl)
		ctx = context.TODO()
		handler = group_controller.NewMeshGroupEventHandler(ctx, validator)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("create", func() {
		It("will return an error if the validator returns an unknown status", func() {
			mg := &v1alpha1.MeshGroup{}
			validator.EXPECT().Validate(ctx, mg).Return(types.MeshGroupStatus{
				Config: types.MeshGroupStatus_PROCESSING_ERROR,
			}, testErr)
			err := handler.Create(mg)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(testErr))
		})

		It("will return nil if the validator returns an invalid status", func() {
			mg := &v1alpha1.MeshGroup{}
			validator.EXPECT().Validate(ctx, mg).Return(types.MeshGroupStatus{
				Config: types.MeshGroupStatus_INVALID,
			}, testErr)
			err := handler.Create(mg)
			Expect(err).NotTo(HaveOccurred())
		})

		It("will return nil if the validator returns an valid status", func() {
			mg := &v1alpha1.MeshGroup{}
			validator.EXPECT().Validate(ctx, mg).Return(types.MeshGroupStatus{
				Config: types.MeshGroupStatus_VALID,
			}, nil)
			err := handler.Create(mg)
			Expect(err).NotTo(HaveOccurred())
		})
	})
	Context("update", func() {
		It("will return an error if the validator returns an unknown status", func() {
			mg := &v1alpha1.MeshGroup{}
			validator.EXPECT().Validate(ctx, mg).Return(types.MeshGroupStatus{
				Config: types.MeshGroupStatus_PROCESSING_ERROR,
			}, testErr)
			err := handler.Update(nil, mg)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(testErr))
		})

		It("will return nil if the validator returns an invalid status", func() {
			mg := &v1alpha1.MeshGroup{}
			validator.EXPECT().Validate(ctx, mg).Return(types.MeshGroupStatus{
				Config: types.MeshGroupStatus_INVALID,
			}, testErr)
			err := handler.Update(nil, mg)
			Expect(err).NotTo(HaveOccurred())
		})

		It("will return nil if the validator returns an valid status", func() {
			mg := &v1alpha1.MeshGroup{}
			validator.EXPECT().Validate(ctx, mg).Return(types.MeshGroupStatus{
				Config: types.MeshGroupStatus_VALID,
			}, nil)
			err := handler.Update(nil, mg)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
