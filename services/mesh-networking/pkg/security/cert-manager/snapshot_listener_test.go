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
		csrProcessor        *mock_cert_manager.MockVirtualMeshCertificateManager
		virtualMeshClient   *mock_zephyr_networking.MockVirtualMeshClient
		csrSnapshotListener cert_manager.VMCSRSnapshotListener
	)

	BeforeEach(func() {
		testLogger = test_logging.NewTestLogger()
		ctx = contextutils.WithExistingLogger(context.TODO(), testLogger.Logger())
		ctrl = gomock.NewController(GinkgoT())
		csrProcessor = mock_cert_manager.NewMockVirtualMeshCertificateManager(ctrl)
		virtualMeshClient = mock_zephyr_networking.NewMockVirtualMeshClient(ctrl)
		csrSnapshotListener = cert_manager.NewVMCSRSnapshotListener(csrProcessor, virtualMeshClient)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("will do nothing if there are no updated virtual meshes", func() {
		snap := &snapshot.MeshNetworkingSnapshot{}
		csrSnapshotListener.Sync(ctx, snap)
		testLogger.EXPECT().
			LastEntry().
			Level(zapcore.DebugLevel).
			HaveMessage(cert_manager.NoVirtualMeshsChangedMessage)
	})

	It("will process all create events in order", func() {
		vm1 := &networking_v1alpha1.VirtualMesh{
			TypeMeta:   metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{},
			Spec:       networking_types.VirtualMeshSpec{},
			Status: networking_types.VirtualMeshStatus{
				CertificateStatus: &core_types.ComputedStatus{
					Status: core_types.ComputedStatus_ACCEPTED,
				},
			},
		}
		vm2 := &networking_v1alpha1.VirtualMesh{
			TypeMeta:   metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{},
			Spec:       networking_types.VirtualMeshSpec{},
			Status: networking_types.VirtualMeshStatus{
				CertificateStatus: &core_types.ComputedStatus{
					Status: core_types.ComputedStatus_INVALID,
				},
			},
		}
		snap := &snapshot.MeshNetworkingSnapshot{
			VirtualMeshes: []*networking_v1alpha1.VirtualMesh{vm1, vm2},
		}
		csrProcessor.EXPECT().InitializeCertificateForVirtualMesh(ctx, vm1).Return(vm1.Status)
		csrProcessor.EXPECT().InitializeCertificateForVirtualMesh(ctx, vm2).Return(vm2.Status)

		virtualMeshClient.EXPECT().UpdateStatus(ctx, vm1).Return(nil)
		virtualMeshClient.EXPECT().UpdateStatus(ctx, vm2).Return(nil)
		csrSnapshotListener.Sync(ctx, snap)
	})
})
