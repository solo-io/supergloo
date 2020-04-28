package internal_watcher_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/testutils"
	mesh_platform "github.com/solo-io/service-mesh-hub/services/common/mesh-platform"
	mock_mesh_platform "github.com/solo-io/service-mesh-hub/services/common/mesh-platform/mocks"
	. "github.com/solo-io/service-mesh-hub/services/common/mesh-platform/secret-event-handler/internal"
	k8s_core_types "k8s.io/api/core/v1"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("multicluster-watcher", func() {

	var (
		ctrl                          *gomock.Controller
		ctx                           context.Context
		mockMeshPlatformCredsHandler  *mock_mesh_platform.MockMeshPlatformCredentialsHandler
		meshPlatformMembershipHandler *MeshPlatformMembershipHandler
		secret                        *k8s_core_types.Secret
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		secret = &k8s_core_types.Secret{
			ObjectMeta: k8s_meta_types.ObjectMeta{Name: "secret-name"},
		}
		mockMeshPlatformCredsHandler = mock_mesh_platform.NewMockMeshPlatformCredentialsHandler(ctrl)
		meshPlatformMembershipHandler = NewMeshPlatformMembershipHandler(
			[]mesh_platform.MeshPlatformCredentialsHandler{
				mockMeshPlatformCredsHandler,
			},
		)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should handle mesh platform addition", func() {
		mockMeshPlatformCredsHandler.
			EXPECT().
			MeshPlatformAdded(ctx, secret).
			Return(nil)
		resync, err := meshPlatformMembershipHandler.MeshPlatformSecretAdded(ctx, secret)
		Expect(err).To(BeNil())
		Expect(resync).To(BeFalse())
	})

	It("should handle mesh platform addition with error", func() {
		err := eris.New("mesh platform add error")
		mockMeshPlatformCredsHandler.
			EXPECT().
			MeshPlatformAdded(ctx, secret).
			Return(err)
		resync, err := meshPlatformMembershipHandler.MeshPlatformSecretAdded(ctx, secret)
		Expect(resync).To(BeFalse())
		Expect(err).To(testutils.HaveInErrorChain(PlatformAddError(err, secret.GetName())))
	})

	It("should handle mesh platform removal", func() {
		mockMeshPlatformCredsHandler.
			EXPECT().
			MeshPlatformRemoved(ctx, secret).
			Return(nil)
		resync, err := meshPlatformMembershipHandler.MeshPlatformSecretRemoved(ctx, secret)
		Expect(err).To(BeNil())
		Expect(resync).To(BeFalse())
	})

	It("should handle mesh platform removal with error", func() {
		err := eris.New("mesh platform remove error")
		mockMeshPlatformCredsHandler.
			EXPECT().
			MeshPlatformRemoved(ctx, secret).
			Return(err)
		resync, err := meshPlatformMembershipHandler.MeshPlatformSecretRemoved(ctx, secret)
		Expect(resync).To(BeFalse())
		Expect(err).To(testutils.HaveInErrorChain(PlatformRemoveError(err, secret.GetName())))
	})
})
