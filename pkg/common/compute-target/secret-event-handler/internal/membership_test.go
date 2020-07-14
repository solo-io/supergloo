package internal_watcher_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/testutils"
	compute_target "github.com/solo-io/service-mesh-hub/pkg/common/compute-target"
	mock_compute_target "github.com/solo-io/service-mesh-hub/pkg/common/compute-target/mocks"
	. "github.com/solo-io/service-mesh-hub/pkg/common/compute-target/secret-event-handler/internal"
	k8s_core_types "k8s.io/api/core/v1"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("multicluster-watcher", func() {

	var (
		ctrl                                *gomock.Controller
		ctx                                 context.Context
		mockComputeTargetCredentialsHandler *mock_compute_target.MockComputeTargetCredentialsHandler
		computeTargetMembershipHandler      *ComputeTargetMembershipHandler
		secret                              *k8s_core_types.Secret
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		secret = &k8s_core_types.Secret{
			ObjectMeta: k8s_meta_types.ObjectMeta{Name: "secret-name"},
		}
		mockComputeTargetCredentialsHandler = mock_compute_target.NewMockComputeTargetCredentialsHandler(ctrl)
		computeTargetMembershipHandler = NewComputeTargetMembershipHandler(
			[]compute_target.ComputeTargetCredentialsHandler{
				mockComputeTargetCredentialsHandler,
			},
		)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should handle compute target addition", func() {
		mockComputeTargetCredentialsHandler.
			EXPECT().
			ComputeTargetAdded(ctx, secret).
			Return(nil)
		resync, err := computeTargetMembershipHandler.ComputeTargetSecretAdded(ctx, secret)
		Expect(err).To(BeNil())
		Expect(resync).To(BeFalse())
	})

	It("should handle compute target addition with error", func() {
		err := eris.New("compute target add error")
		mockComputeTargetCredentialsHandler.
			EXPECT().
			ComputeTargetAdded(ctx, secret).
			Return(err)
		resync, err := computeTargetMembershipHandler.ComputeTargetSecretAdded(ctx, secret)
		Expect(resync).To(BeFalse())
		Expect(err).To(testutils.HaveInErrorChain(ComputeTargetAddError(err, secret.GetName())))
	})

	It("should handle compute target removal", func() {
		mockComputeTargetCredentialsHandler.
			EXPECT().
			ComputeTargetRemoved(ctx, secret).
			Return(nil)
		resync, err := computeTargetMembershipHandler.ComputeTargetSecretRemoved(ctx, secret)
		Expect(err).To(BeNil())
		Expect(resync).To(BeFalse())
	})

	It("should handle compute target removal with error", func() {
		err := eris.New("compute target remove error")
		mockComputeTargetCredentialsHandler.
			EXPECT().
			ComputeTargetRemoved(ctx, secret).
			Return(err)
		resync, err := computeTargetMembershipHandler.ComputeTargetSecretRemoved(ctx, secret)
		Expect(resync).To(BeFalse())
		Expect(err).To(testutils.HaveInErrorChain(ComputeTargetRemoveError(err, secret.GetName())))
	})
})
