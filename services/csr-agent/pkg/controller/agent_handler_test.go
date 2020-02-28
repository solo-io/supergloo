package csr_agent_controller_test

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
	mock_security "github.com/solo-io/mesh-projects/services/common/processors/security/mocks"
	csr_agent_controller "github.com/solo-io/mesh-projects/services/csr-agent/pkg/controller"
	test_logging "github.com/solo-io/mesh-projects/test/logging"
	mock_controller_runtime "github.com/solo-io/mesh-projects/test/mocks/controller-runtime"
	"go.uber.org/zap/zapcore"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

var _ = Describe("csr-agent controller", func() {
	var (
		ctrl       *gomock.Controller
		ctx        context.Context
		csrClient  *mock_security_config.MockMeshGroupCertificateSigningRequestClient
		csrHandler controller.MeshGroupCertificateSigningRequestEventHandler
		predicate  *mock_controller_runtime.MockPredicate
		processor  *mock_security.MockMeshGroupCertificateSigningRequestProcessor
		testLogger *test_logging.TestLogger

		testErr = eris.New("hello")
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		csrClient = mock_security_config.NewMockMeshGroupCertificateSigningRequestClient(ctrl)
		predicate = mock_controller_runtime.NewMockPredicate(ctrl)
		processor = mock_security.NewMockMeshGroupCertificateSigningRequestProcessor(ctrl)
		testLogger = test_logging.NewTestLogger()
		ctx = contextutils.WithExistingLogger(context.TODO(), testLogger.Logger())

		csrHandler = csr_agent_controller.NewCsrAgentEventHandler(ctx, csrClient, processor, predicate)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("unexpected events", func() {
		var (
			csr = &securityv1alpha1.MeshGroupCertificateSigningRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "name",
					Namespace: "namespace",
				},
			}
		)

		It("delete", func() {
			Expect(csrHandler.Delete(csr)).NotTo(HaveOccurred())
			testLogger.EXPECT().NumEntries(1).LastEntry().Level(zapcore.DebugLevel).
				HaveMessage(csr_agent_controller.UnexpectedEventMsg).
				Have(logging.EventTypeKey, logging.DeleteEvent.String())

		})

		It("generic", func() {
			Expect(csrHandler.Generic(csr)).NotTo(HaveOccurred())
			testLogger.EXPECT().NumEntries(1).LastEntry().Level(zapcore.DebugLevel).
				HaveMessage(csr_agent_controller.UnexpectedEventMsg).
				Have(logging.EventTypeKey, logging.GenericEvent.String())

		})

	})

	Context("create", func() {

		It("will return nil and log skip if predicate is false", func() {
			csr := &securityv1alpha1.MeshGroupCertificateSigningRequest{}
			predicate.EXPECT().
				Create(event.CreateEvent{Object: csr, Meta: csr}).
				Return(false)

			Expect(csrHandler.Create(csr)).NotTo(HaveOccurred())

			testLogger.EXPECT().
				LastEntry().
				Level(zapcore.DebugLevel).
				HaveMessage("skipping event")
		})

		It("will log an error and set status frrom processor", func() {
			csr := &securityv1alpha1.MeshGroupCertificateSigningRequest{
				Status: security_types.MeshGroupCertificateSigningRequestStatus{
					ComputedStatus: &core_types.ComputedStatus{
						Status:  core_types.ComputedStatus_INVALID,
						Message: testErr.Error(),
					},
				},
			}

			predicate.EXPECT().
				Create(event.CreateEvent{Object: csr, Meta: csr}).
				Return(true)

			processor.EXPECT().
				ProcessCreate(ctx, csr).
				Return(csr.Status)

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

		It("will return nil and log skip if predicate is false", func() {
			csr := &securityv1alpha1.MeshGroupCertificateSigningRequest{}
			predicate.EXPECT().
				Update(event.UpdateEvent{
					MetaOld:   csr,
					ObjectOld: csr,
					MetaNew:   csr,
					ObjectNew: csr,
				}).
				Return(false)

			Expect(csrHandler.Update(csr, csr)).NotTo(HaveOccurred())

			testLogger.EXPECT().
				LastEntry().
				Level(zapcore.DebugLevel).
				HaveMessage("skipping event")
		})

		It("will log an error and set status frrom processor", func() {
			csr := &securityv1alpha1.MeshGroupCertificateSigningRequest{
				Status: security_types.MeshGroupCertificateSigningRequestStatus{
					ComputedStatus: &core_types.ComputedStatus{
						Status:  core_types.ComputedStatus_INVALID,
						Message: testErr.Error(),
					},
				},
			}

			predicate.EXPECT().
				Update(event.UpdateEvent{
					MetaOld:   csr,
					ObjectOld: csr,
					MetaNew:   csr,
					ObjectNew: csr,
				}).
				Return(true)

			processor.EXPECT().
				ProcessUpdate(ctx, csr, csr).
				Return(csr.Status)

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
