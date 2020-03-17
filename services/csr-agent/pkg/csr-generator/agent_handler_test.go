package csr_generator_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/contextutils"
	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	securityv1alpha1 "github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1/controller"
	security_types "github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1/types"
	mock_security_config "github.com/solo-io/mesh-projects/pkg/clients/zephyr/security/mocks"
	"github.com/solo-io/mesh-projects/pkg/logging"
	csr_generator "github.com/solo-io/mesh-projects/services/csr-agent/pkg/csr-generator"
	mock_csr_generator "github.com/solo-io/mesh-projects/services/csr-agent/pkg/csr-generator/mocks"
	test_logging "github.com/solo-io/mesh-projects/test/logging"
	"go.uber.org/zap/zapcore"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("csr-agent controller", func() {
	var (
		ctrl       *gomock.Controller
		ctx        context.Context
		csrClient  *mock_security_config.MockVirtualMeshCSRClient
		csrHandler controller.VirtualMeshCertificateSigningRequestEventHandler
		processor  *mock_csr_generator.MockVirtualMeshCSRProcessor
		testLogger *test_logging.TestLogger

		testErr = eris.New("hello")
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		csrClient = mock_security_config.NewMockVirtualMeshCSRClient(ctrl)
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
			csr = &securityv1alpha1.VirtualMeshCertificateSigningRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "name",
					Namespace: "namespace",
				},
			}
		)

		It("delete", func() {
			Expect(csrHandler.Delete(csr)).NotTo(HaveOccurred())
			testLogger.EXPECT().NumEntries(1).LastEntry().Level(zapcore.DebugLevel).
				HaveMessage(csr_generator.UnexpectedEventMsg).
				Have(logging.EventTypeKey, logging.DeleteEvent.String())

		})

		It("generic", func() {
			Expect(csrHandler.Generic(csr)).NotTo(HaveOccurred())
			testLogger.EXPECT().NumEntries(1).LastEntry().Level(zapcore.DebugLevel).
				HaveMessage(csr_generator.UnexpectedEventMsg).
				Have(logging.EventTypeKey, logging.GenericEvent.String())

		})

	})

	Context("create", func() {

		It("will return nil and log if event is not processed", func() {
			csr := &securityv1alpha1.VirtualMeshCertificateSigningRequest{}

			processor.EXPECT().
				ProcessUpsert(ctx, csr).
				Return(nil)

			Expect(csrHandler.Create(csr)).NotTo(HaveOccurred())

			testLogger.EXPECT().
				LastEntry().
				Level(zapcore.DebugLevel).
				HaveMessage("csr event was not processed")
		})

		It("will log an error and set status frrom processor", func() {
			csr := &securityv1alpha1.VirtualMeshCertificateSigningRequest{
				Status: security_types.VirtualMeshCertificateSigningRequestStatus{
					ComputedStatus: &core_types.ComputedStatus{
						Status:  core_types.ComputedStatus_INVALID,
						Message: testErr.Error(),
					},
				},
			}

			processor.EXPECT().
				ProcessUpsert(ctx, csr).
				Return(&csr.Status)

			csrClient.EXPECT().
				UpdateStatus(ctx, csr).
				Return(testErr)

			Expect(csrHandler.Create(csr)).NotTo(HaveOccurred())

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
			csr := &securityv1alpha1.VirtualMeshCertificateSigningRequest{}

			processor.EXPECT().
				ProcessUpsert(ctx, csr).
				Return(nil)

			Expect(csrHandler.Update(csr, csr)).NotTo(HaveOccurred())

			testLogger.EXPECT().
				LastEntry().
				Level(zapcore.DebugLevel).
				HaveMessage("csr event was not processed")
		})

		It("will log an error and set status frrom processor", func() {
			csr := &securityv1alpha1.VirtualMeshCertificateSigningRequest{
				Status: security_types.VirtualMeshCertificateSigningRequestStatus{
					ComputedStatus: &core_types.ComputedStatus{
						Status:  core_types.ComputedStatus_INVALID,
						Message: testErr.Error(),
					},
				},
			}

			processor.EXPECT().
				ProcessUpsert(ctx, csr).
				Return(&csr.Status)

			csrClient.EXPECT().
				UpdateStatus(ctx, csr).
				Return(testErr)

			Expect(csrHandler.Update(csr, csr)).NotTo(HaveOccurred())

			testLogger.EXPECT().LastEntry().
				HaveError(testErr).
				Level(zapcore.ErrorLevel)

			testLogger.EXPECT().Entry(testLogger.NumLogEntries() - 2).
				Level(zapcore.DebugLevel).
				HaveError(testErr)
		})

	})

})
