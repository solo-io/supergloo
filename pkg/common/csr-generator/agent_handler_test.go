package csr_generator_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/contextutils"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_security "github.com/solo-io/service-mesh-hub/pkg/api/security.smh.solo.io/v1alpha1"
	smh_security_controller "github.com/solo-io/service-mesh-hub/pkg/api/security.smh.solo.io/v1alpha1/controller"
	mock_security_config "github.com/solo-io/service-mesh-hub/pkg/api/security.smh.solo.io/v1alpha1/mocks"
	smh_security_types "github.com/solo-io/service-mesh-hub/pkg/api/security.smh.solo.io/v1alpha1/types"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/common/container-runtime"
	csr_generator "github.com/solo-io/service-mesh-hub/pkg/common/csr-generator"
	mock_csr_generator "github.com/solo-io/service-mesh-hub/pkg/common/csr-generator/mocks"
	test_logging "github.com/solo-io/service-mesh-hub/test/logging"
	"go.uber.org/zap/zapcore"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("csr-agent controller", func() {
	var (
		ctrl       *gomock.Controller
		ctx        context.Context
		csrClient  *mock_security_config.MockVirtualMeshCertificateSigningRequestClient
		csrHandler smh_security_controller.VirtualMeshCertificateSigningRequestEventHandler
		processor  *mock_csr_generator.MockVirtualMeshCSRProcessor
		testLogger *test_logging.TestLogger

		testErr = eris.New("hello")
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		csrClient = mock_security_config.NewMockVirtualMeshCertificateSigningRequestClient(ctrl)
		processor = mock_csr_generator.NewMockVirtualMeshCSRProcessor(ctrl)
		testLogger = test_logging.NewTestLogger()
		ctx = contextutils.WithExistingLogger(context.TODO(), testLogger.Logger())

		csrHandler = csr_generator.NewVirtualMeshCSRDataSource(ctx, csrClient, processor)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("unexpected events", func() {
		var (
			csr = &smh_security.VirtualMeshCertificateSigningRequest{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "name",
					Namespace: "namespace",
				},
			}
		)

		It("delete", func() {
			Expect(csrHandler.DeleteVirtualMeshCertificateSigningRequest(csr)).NotTo(HaveOccurred())
			testLogger.EXPECT().NumEntries(1).LastEntry().Level(zapcore.DebugLevel).
				HaveMessage(csr_generator.UnexpectedEventMsg).
				Have(container_runtime.EventTypeKey, container_runtime.DeleteEvent.String())

		})

		It("generic", func() {
			Expect(csrHandler.GenericVirtualMeshCertificateSigningRequest(csr)).NotTo(HaveOccurred())
			testLogger.EXPECT().NumEntries(1).LastEntry().Level(zapcore.DebugLevel).
				HaveMessage(csr_generator.UnexpectedEventMsg).
				Have(container_runtime.EventTypeKey, container_runtime.GenericEvent.String())

		})

	})

	Context("create", func() {

		It("will return nil and log if event is not processed", func() {
			csr := &smh_security.VirtualMeshCertificateSigningRequest{}

			processor.EXPECT().
				ProcessUpsert(ctx, csr).
				Return(nil)

			Expect(csrHandler.CreateVirtualMeshCertificateSigningRequest(csr)).NotTo(HaveOccurred())

			testLogger.EXPECT().
				LastEntry().
				Level(zapcore.DebugLevel).
				HaveMessage("csr event was not processed")
		})

		It("will log an error and set status frrom processor", func() {
			csr := &smh_security.VirtualMeshCertificateSigningRequest{
				Status: smh_security_types.VirtualMeshCertificateSigningRequestStatus{
					ComputedStatus: &smh_core_types.Status{
						State:   smh_core_types.Status_INVALID,
						Message: testErr.Error(),
					},
				},
			}

			processor.EXPECT().
				ProcessUpsert(ctx, csr).
				Return(&csr.Status)

			csrClient.EXPECT().
				UpdateVirtualMeshCertificateSigningRequestStatus(ctx, csr).
				Return(testErr)

			Expect(csrHandler.CreateVirtualMeshCertificateSigningRequest(csr)).NotTo(HaveOccurred())

			testLogger.EXPECT().LastEntry().
				HaveError(testErr).
				Level(zapcore.ErrorLevel)

			testLogger.EXPECT().Entry(testLogger.NumLogEntries() - 2).
				Level(zapcore.DebugLevel).
				HaveError(testErr)
		})
	})

	Context("update", func() {

		It("will return nil and log if event is not processed", func() {
			csr := &smh_security.VirtualMeshCertificateSigningRequest{}

			processor.EXPECT().
				ProcessUpsert(ctx, csr).
				Return(nil)

			Expect(csrHandler.UpdateVirtualMeshCertificateSigningRequest(csr, csr)).NotTo(HaveOccurred())

			testLogger.EXPECT().
				LastEntry().
				Level(zapcore.DebugLevel).
				HaveMessage("csr event was not processed")
		})

		It("will log an error and set status frrom processor", func() {
			csr := &smh_security.VirtualMeshCertificateSigningRequest{
				Status: smh_security_types.VirtualMeshCertificateSigningRequestStatus{
					ComputedStatus: &smh_core_types.Status{
						State:   smh_core_types.Status_INVALID,
						Message: testErr.Error(),
					},
				},
			}

			processor.EXPECT().
				ProcessUpsert(ctx, csr).
				Return(&csr.Status)

			csrClient.EXPECT().
				UpdateVirtualMeshCertificateSigningRequestStatus(ctx, csr).
				Return(testErr)

			Expect(csrHandler.UpdateVirtualMeshCertificateSigningRequest(csr, csr)).NotTo(HaveOccurred())

			testLogger.EXPECT().LastEntry().
				HaveError(testErr).
				Level(zapcore.ErrorLevel)

			testLogger.EXPECT().Entry(testLogger.NumLogEntries() - 2).
				Level(zapcore.DebugLevel).
				HaveError(testErr)
		})

	})

})
