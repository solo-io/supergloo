package reconciliation_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/issuer/input"
	mock_input "github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/issuer/input/mocks"
	certificatesv1 "github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/v1"
	v1sets "github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/v1/sets"
	"github.com/solo-io/gloo-mesh/pkg/certificates/issuer/reconciliation"
	"github.com/solo-io/gloo-mesh/pkg/certificates/issuer/translation"
	mock_translation "github.com/solo-io/gloo-mesh/pkg/certificates/issuer/translation/mocks"
	"github.com/solo-io/gloo-mesh/test/matchers"
	skv2_matchers "github.com/solo-io/skv2/test/matchers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("CertIssueReconciler", func() {
	var (
		ctrl           *gomock.Controller
		ctx            context.Context
		mockTranslator *mock_translation.MockTranslator
		mockBuilder    *mock_input.MockBuilder
	)

	BeforeEach(func() {
		ctrl, ctx = gomock.WithContext(context.Background(), GinkgoT())
		mockTranslator = mock_translation.NewMockTranslator(ctrl)
		mockBuilder = mock_input.NewMockBuilder(ctrl)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("will do nothing if workflow has finished", func() {
		reconcileFunc := reconciliation.NewCertificateRequestReconciler(
			ctx,
			mockBuilder,
			func(ctx context.Context, snapshot input.Snapshot) error {
				return nil
			},
			mockTranslator,
		)

		certRequest := &certificatesv1.CertificateRequest{
			ObjectMeta: metav1.ObjectMeta{
				Generation: 2,
			},
			Spec: certificatesv1.CertificateRequestSpec{
				CertificateSigningRequest: []byte("hello"),
			},
			Status: certificatesv1.CertificateRequestStatus{
				ObservedGeneration: 2,
				State:              certificatesv1.CertificateRequestStatus_FINISHED,
			},
		}

		mockBuilder.EXPECT().
			BuildSnapshot(gomock.Any(), "cert-issuer", input.BuildOptions{}).
			Return(input.NewSnapshot(
				"hello",
				v1sets.NewIssuedCertificateSet(),
				v1sets.NewCertificateRequestSet(certRequest),
			), nil)
		_, err := reconcileFunc(nil)

		Expect(err).NotTo(HaveOccurred())

	})

	It("will set csr output from translator to CertificateRequest status", func() {
		reconcileFunc := reconciliation.NewCertificateRequestReconciler(
			ctx,
			mockBuilder,
			func(ctx context.Context, snapshot input.Snapshot) error {
				return nil
			},
			mockTranslator,
		)

		certRequest := &certificatesv1.CertificateRequest{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "issued-cert",
				Namespace:  "ns",
				Generation: 2,
			},
			Spec: certificatesv1.CertificateRequestSpec{
				CertificateSigningRequest: []byte("hello"),
			},
		}

		issuedCert := &certificatesv1.IssuedCertificate{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "issued-cert",
				Namespace: "ns",
			},
		}

		mockBuilder.EXPECT().
			BuildSnapshot(gomock.Any(), "cert-issuer", input.BuildOptions{}).
			Return(input.NewSnapshot(
				"hello",
				v1sets.NewIssuedCertificateSet(issuedCert),
				v1sets.NewCertificateRequestSet(certRequest),
			), nil)

		output := &translation.Output{
			SignedCertificate: []byte("cert"),
			SigningRootCa:     []byte("ca"),
		}

		expectedCertRequest := certRequest.DeepCopy()
		expectedCertRequest.Status.State = certificatesv1.CertificateRequestStatus_PENDING
		expectedCertRequest.Status.ObservedGeneration = 2

		mockTranslator.EXPECT().
			Translate(
				gomock.Any(),
				matchers.GomockMatchPublicFields(expectedCertRequest),
				matchers.GomockMatchPublicFields(issuedCert),
			).
			Return(output, nil)
		_, err := reconcileFunc(nil)
		Expect(err).NotTo(HaveOccurred())

		Expect(&certRequest.Status).To(skv2_matchers.MatchProto(&certificatesv1.CertificateRequestStatus{
			ObservedGeneration: 2,
			State:              certificatesv1.CertificateRequestStatus_FINISHED,
			SignedCertificate:  output.SignedCertificate,
			SigningRootCa:      output.SigningRootCa,
		}))

	})

	It("will set csr error from translator to CertificateRequest status", func() {
		reconcileFunc := reconciliation.NewCertificateRequestReconciler(
			ctx,
			mockBuilder,
			func(ctx context.Context, snapshot input.Snapshot) error {
				return nil
			},
			mockTranslator,
		)

		certRequest := &certificatesv1.CertificateRequest{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "issued-cert",
				Namespace:  "ns",
				Generation: 2,
			},
			Spec: certificatesv1.CertificateRequestSpec{
				CertificateSigningRequest: []byte("hello"),
			},
			Status: certificatesv1.CertificateRequestStatus{
				ObservedGeneration: 1,
				State:              certificatesv1.CertificateRequestStatus_FINISHED,
			},
		}

		issuedCert := &certificatesv1.IssuedCertificate{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "issued-cert",
				Namespace: "ns",
			},
		}

		mockBuilder.EXPECT().
			BuildSnapshot(gomock.Any(), "cert-issuer", input.BuildOptions{}).
			Return(input.NewSnapshot(
				"hello",
				v1sets.NewIssuedCertificateSet(issuedCert),
				v1sets.NewCertificateRequestSet(certRequest),
			), nil)

		expectedCertRequest := certRequest.DeepCopy()
		expectedCertRequest.Status.State = certificatesv1.CertificateRequestStatus_PENDING
		expectedCertRequest.Status.ObservedGeneration = 2

		mockTranslator.EXPECT().
			Translate(
				gomock.Any(),
				matchers.GomockMatchPublicFields(expectedCertRequest),
				matchers.GomockMatchPublicFields(issuedCert),
			).
			Return(nil, eris.New("hello"))

		_, err := reconcileFunc(nil)
		Expect(err).NotTo(HaveOccurred())

		Expect(&certRequest.Status).To(skv2_matchers.MatchProto(&certificatesv1.CertificateRequestStatus{
			ObservedGeneration: 2,
			State:              certificatesv1.CertificateRequestStatus_FAILED,
			Error:              "failed to translate certificate request + issued certificate: hello",
		}))

	})
})
