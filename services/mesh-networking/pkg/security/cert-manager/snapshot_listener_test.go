package cert_manager_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	"github.com/solo-io/go-utils/contextutils"
	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	networking_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	networking_types "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	mock_zephyr_networking "github.com/solo-io/mesh-projects/pkg/clients/zephyr/networking/mocks"
	"github.com/solo-io/mesh-projects/services/mesh-networking/pkg/multicluster/snapshot"
	cert_manager "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/security/cert-manager"
	mock_cert_manager "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/security/cert-manager/mocks"
	test_logging "github.com/solo-io/mesh-projects/test/logging"
	"go.uber.org/zap/zapcore"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("snapshot listener", func() {
	var (
		ctrl                *gomock.Controller
		ctx                 context.Context
		testLogger          *test_logging.TestLogger
		csrProcessor        *mock_cert_manager.MockMeshGroupCertificateManager
		meshGroupClient     *mock_zephyr_networking.MockMeshGroupClient
		csrSnapshotListener cert_manager.GroupMgcsrSnapshotListener
	)

	BeforeEach(func() {
		testLogger = test_logging.NewTestLogger()
		ctx = contextutils.WithExistingLogger(context.TODO(), testLogger.Logger())
		ctrl = gomock.NewController(GinkgoT())
		csrProcessor = mock_cert_manager.NewMockMeshGroupCertificateManager(ctrl)
		meshGroupClient = mock_zephyr_networking.NewMockMeshGroupClient(ctrl)
		csrSnapshotListener = cert_manager.NewGroupMgcsrSnapshotListener(csrProcessor, meshGroupClient)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("will do nothing if there are no updated mesh groups", func() {
		snap := &snapshot.MeshNetworkingSnapshot{}
		csrSnapshotListener.Sync(ctx, snap)
		testLogger.EXPECT().
			LastEntry().
			Level(zapcore.DebugLevel).
			HaveMessage(cert_manager.NoMeshGroupsChangedMessage)
	})

	It("will process all create events in order", func() {
		mg1 := &networking_v1alpha1.MeshGroup{
			TypeMeta:   metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{},
			Spec:       networking_types.MeshGroupSpec{},
			Status: networking_types.MeshGroupStatus{
				CertificateStatus: &core_types.ComputedStatus{
					Status: core_types.ComputedStatus_ACCEPTED,
				},
			},
		}
		mg2 := &networking_v1alpha1.MeshGroup{
			TypeMeta:   metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{},
			Spec:       networking_types.MeshGroupSpec{},
			Status: networking_types.MeshGroupStatus{
				CertificateStatus: &core_types.ComputedStatus{
					Status: core_types.ComputedStatus_INVALID,
				},
			},
		}
		snap := &snapshot.MeshNetworkingSnapshot{
			MeshGroups: []*networking_v1alpha1.MeshGroup{mg1, mg2},
		}
		csrProcessor.EXPECT().InitializeCertificateForMeshGroup(ctx, mg1).Return(mg1.Status)
		csrProcessor.EXPECT().InitializeCertificateForMeshGroup(ctx, mg2).Return(mg2.Status)

		meshGroupClient.EXPECT().UpdateStatus(ctx, mg1).Return(nil)
		meshGroupClient.EXPECT().UpdateStatus(ctx, mg2).Return(nil)
		csrSnapshotListener.Sync(ctx, snap)
	})
})
