package translation_test

import (
	"context"
	"crypto/x509"
	"encoding/pem"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/agent/input"
	mock_certagent "github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/agent/output/certagent/mocks"
	certificatesv1 "github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/v1"
	v1 "github.com/solo-io/gloo-mesh/pkg/api/common.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/certificates/agent/translation"
	"github.com/solo-io/gloo-mesh/pkg/certificates/common/secrets"
	skv2corev1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/solo-io/skv2/pkg/ezkube"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

		It("will do nothing if no IssuedCertSecret is not present", func() {

			translator := translation.NewCertAgentTranslator()

			issuedCertiticate := &certificatesv1.IssuedCertificate{}
			inputSnap := input.NewInputSnapshotManualBuilder("hello").
				Build()

			csrBytes, err := translator.IssuedCertiticatePending(ctx, issuedCertiticate, inputSnap, mockOutput)
			Expect(err).NotTo(HaveOccurred())
			Expect(csrBytes).To(BeNil())
		})

		It("Will create the private key secret, and return CSR bytes", func() {
			translator := translation.NewCertAgentTranslator()

			issuedCertiticate := &certificatesv1.IssuedCertificate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "name",
					Namespace: "namespace",
				},
				Spec: certificatesv1.IssuedCertificateSpec{
					IssuedCertificateSecret: &skv2corev1.ObjectRef{},
					CertOptions: &v1.CommonCertOptions{
						OrgName: "istio",
					},
				},
			}
			inputSnap := input.NewInputSnapshotManualBuilder("hello").
				Build()

			mockOutput.EXPECT().
				AddSecrets(gomock.Any()).
				Do(func(secret *corev1.Secret) {
					Expect(secret.ObjectMeta).To(Equal(metav1.ObjectMeta{
						Name:      issuedCertiticate.Name,
						Namespace: issuedCertiticate.Namespace,
						Labels: map[string]string{
							"agent.certificates.mesh.gloo.solo.io": "gloo-mesh",
						},
					}))
					Expect(secret.Type).To(Equal(translation.PrivateKeySecretType()))
					pemByt, _ := pem.Decode(secret.Data["private-key"])
					Expect(pemByt.Type).To(Equal("RSA PRIVATE KEY"))
				})

			csrBytes, err := translator.IssuedCertiticatePending(ctx, issuedCertiticate, inputSnap, mockOutput)
			Expect(err).NotTo(HaveOccurred())

			Expect(csrBytes).NotTo(BeNil())
			pemByt, _ := pem.Decode(csrBytes)
			csr, err := x509.ParseCertificateRequest(pemByt.Bytes)
			Expect(err).NotTo(HaveOccurred())
			Expect(csr.Subject.Organization).To(ConsistOf("istio"))
			Expect(csr.Extensions)
		})

	})

	Context("IssuedCertiticateRequested", func() {

		It("will do nothing if no IssuedCertSecret is not present", func() {

			translator := translation.NewCertAgentTranslator()

			issuedCertiticate := &certificatesv1.IssuedCertificate{}
			csr := &certificatesv1.CertificateRequest{}
			inputSnap := input.NewInputSnapshotManualBuilder("hello").
				Build()

			continueIterating, err := translator.IssuedCertificateRequested(ctx, issuedCertiticate, csr, inputSnap, mockOutput)
			Expect(err).NotTo(HaveOccurred())
			Expect(continueIterating).To(BeFalse())
		})

		var (
			privateKeySecret = &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "name",
					Namespace: "namespace",
				},
				Data: map[string][]byte{
					"private-key": []byte("hello"),
				},
			}

			issuedCertiticate = &certificatesv1.IssuedCertificate{
				ObjectMeta: privateKeySecret.ObjectMeta,
				Spec: certificatesv1.IssuedCertificateSpec{
					IssuedCertificateSecret: &skv2corev1.ObjectRef{
						Name:      "issued",
						Namespace: "cert",
					},
				},
			}
		)

		It("will re-add csr and secret if csr is pending", func() {
			translator := translation.NewCertAgentTranslator()

			csr := &certificatesv1.CertificateRequest{
				Status: certificatesv1.CertificateRequestStatus{
					State: certificatesv1.CertificateRequestStatus_PENDING,
				},
			}

			inputSnap := input.NewInputSnapshotManualBuilder("hello").
				AddSecrets([]*corev1.Secret{privateKeySecret}).
				Build()

			mockOutput.EXPECT().AddSecrets(privateKeySecret)
			mockOutput.EXPECT().AddCertificateRequests(csr)

			continueIterating, err := translator.IssuedCertificateRequested(ctx, issuedCertiticate, csr, inputSnap, mockOutput)
			Expect(err).NotTo(HaveOccurred())
			Expect(continueIterating).To(BeTrue())
		})

		It("will save certs to issued cert if csr is finished", func() {
			translator := translation.NewCertAgentTranslator()

			csr := &certificatesv1.CertificateRequest{
				Status: certificatesv1.CertificateRequestStatus{
					State:             certificatesv1.CertificateRequestStatus_FINISHED,
					SignedCertificate: []byte("I'm a signing cert"),
					SigningRootCa:     []byte("I'm a root ca"),
				},
			}

			inputSnap := input.NewInputSnapshotManualBuilder("hello").
				AddSecrets([]*corev1.Secret{privateKeySecret}).
				Build()

			mockOutput.EXPECT().
				AddSecrets(gomock.Any()).
				Do(func(secret *corev1.Secret) {
					Expect(secret.ObjectMeta).To(Equal(metav1.ObjectMeta{
						Name:      issuedCertiticate.Spec.IssuedCertificateSecret.Name,
						Namespace: issuedCertiticate.Spec.IssuedCertificateSecret.Namespace,
						Labels: map[string]string{
							"agent.certificates.mesh.gloo.solo.io": "gloo-mesh",
						},
					}))
					Expect(secret.Type).To(Equal(translation.IssuedCertificateSecretType()))
					intCaData := secrets.IntermediateCADataFromSecretData(secret.Data)
					Expect(intCaData.CaPrivateKey).To(Equal([]byte("hello")))
					Expect(intCaData.RootCert).To(Equal([]byte("I'm a root ca")))
					Expect(intCaData.CaCert).To(Equal([]byte("I'm a signing cert")))
				})

			continueIterating, err := translator.IssuedCertificateRequested(ctx, issuedCertiticate, csr, inputSnap, mockOutput)
			Expect(err).NotTo(HaveOccurred())
			Expect(continueIterating).To(BeTrue())
		})

	})

	Context("IssuedCertiticateIssued", func() {

		It("will do nothing if no IssuedCertSecret is not present", func() {

			translator := translation.NewCertAgentTranslator()

			issuedCertiticate := &certificatesv1.IssuedCertificate{}
			inputSnap := input.NewInputSnapshotManualBuilder("hello").
				Build()

			continueIterating, err := translator.IssuedCertificateIssued(ctx, issuedCertiticate, inputSnap, mockOutput)
			Expect(err).NotTo(HaveOccurred())
			Expect(continueIterating).To(BeFalse())
		})

		It("Will create the provate key secret, and return CSR bytes", func() {
			translator := translation.NewCertAgentTranslator()

			issuedCertSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "issued",
					Namespace: "cert",
				},
			}

			issuedCertiticate := &certificatesv1.IssuedCertificate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "name",
					Namespace: "namespace",
				},
				Spec: certificatesv1.IssuedCertificateSpec{
					IssuedCertificateSecret: ezkube.MakeObjectRef(issuedCertSecret),
					CertOptions: &v1.CommonCertOptions{
						OrgName: "istio",
					},
				},
			}
			inputSnap := input.NewInputSnapshotManualBuilder("hello").
				AddSecrets([]*corev1.Secret{issuedCertSecret}).
				Build()

			mockOutput.EXPECT().
				AddSecrets(issuedCertSecret)

			continueIterating, err := translator.IssuedCertificateIssued(ctx, issuedCertiticate, inputSnap, mockOutput)
			Expect(err).NotTo(HaveOccurred())
			Expect(continueIterating).To(BeTrue())
		})

	})
})
