package translation_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/agent/input"
	mock_certagent "github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/agent/output/certagent/mocks"
	certificatesv1 "github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/certificates/agent/translation"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("CertAgentTranslator", func() {
	var (
		ctrl *gomock.Controller
		ctx  context.Context

		mockOutput *mock_certagent.MockBuilder
	)

	BeforeEach(func() {
		ctrl, ctx = gomock.WithContext(context.Background(), GinkgoT())

		mockOutput = mock_certagent.NewMockBuilder(ctrl)
	})

	AfterEach(func() {
		ctrl.Finish()
	})
	Context("IssuedCertiticatePending", func() {

		It("Will create the provate key secret, and return CSR bytes", func() {
			translator := translation.NewCertAgentTranslator()

			// Cannot match secret, and CSR directly as private key and CSR bytes are not idempotent

			issuedCertiticate := &certificatesv1.IssuedCertificate{}
			inputSnap := input.NewInputSnapshotManualBuilder("hello").
				Build()

			mockOutput.EXPECT().
				AddSecrets(gomock.Any()).
				Do(func(secret *corev1.Secret) {

				})

			csrBytes, err := translator.IssuedCertiticatePending(ctx, issuedCertiticate, inputSnap, mockOutput)
			Expect(csrBytes).NotTo(BeNil())
			Expect(err).NotTo(HaveOccurred())
		})

	})

	Context("IssuedCertiticateRequested", func() {

	})

	Context("IssuedCertiticateIssued", func() {

	})
})
