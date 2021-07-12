package reconciliation

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	v1sets "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/sets"
	"github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/agent/input"
	mock_certagent "github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/agent/output/certagent/mocks"
	certificatesv1 "github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/v1"
	mock_podbouncer "github.com/solo-io/gloo-mesh/pkg/certificates/agent/reconciliation/pod-bouncer/mocks"
	mock_translation "github.com/solo-io/gloo-mesh/pkg/certificates/agent/translation/mocks"
	"github.com/solo-io/skv2/pkg/ezkube"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("CertAgentReconciler", func() {
	var (
		ctx  context.Context
		ctrl *gomock.Controller

		mockTranslator *mock_translation.MockTranslator
		mockPodBouncer *mock_podbouncer.MockPodBouncer
		mockOutput     *mock_certagent.MockBuilder
	)

	BeforeEach(func() {
		ctrl, ctx = gomock.WithContext(context.Background(), GinkgoT())

		mockTranslator = mock_translation.NewMockTranslator(ctrl)
		mockPodBouncer = mock_podbouncer.NewMockPodBouncer(ctrl)
		mockOutput = mock_certagent.NewMockBuilder(ctrl)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("IssuedCertificatePending", func() {
		var (
			issuedCert *certificatesv1.IssuedCertificate
			csr        *certificatesv1.CertificateRequest
			csrBytes   []byte
		)
		BeforeEach(func() {

			issuedCert = &certificatesv1.IssuedCertificate{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "hello",
					Namespace:  "world",
					Generation: 3,
				},
				Spec: certificatesv1.IssuedCertificateSpec{},
				Status: certificatesv1.IssuedCertificateStatus{
					State:              certificatesv1.IssuedCertificateStatus_FINISHED,
					ObservedGeneration: 2,
				},
			}

			csrBytes = []byte("I'm a CSR")

			csr = &certificatesv1.CertificateRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      issuedCert.GetName(),
					Namespace: issuedCert.GetNamespace(),
					Labels: map[string]string{
						"agent.certificates.mesh.gloo.solo.io": "gloo-mesh",
					},
				},
				Spec: certificatesv1.CertificateRequestSpec{
					CertificateSigningRequest: csrBytes,
				},
			}

		})

		It("Create CSR if translator Pending func returns properly", func() {
			reconciler := &certAgentReconciler{
				ctx:        ctx,
				podBouncer: mockPodBouncer,
				translator: mockTranslator,
			}

			inputSnap := input.NewInputSnapshotManualBuilder("hello").
				Build()

			mockTranslator.EXPECT().
				IssuedCertiticatePending(gomock.Any(), issuedCert, inputSnap, mockOutput).
				Return(csrBytes, nil)

			mockOutput.EXPECT().AddCertificateRequests(csr)

			mockTranslator.EXPECT().
				ShouldProcess(ctx, issuedCert).
				Return(true)

			err := reconciler.reconcileIssuedCertificate(issuedCert, inputSnap, mockOutput)
			Expect(err).NotTo(HaveOccurred())
			Expect(issuedCert.Status.State).To(Equal(certificatesv1.IssuedCertificateStatus_REQUESTED))
		})

		It("Will do nothing if should Process is false", func() {

			reconciler := &certAgentReconciler{
				ctx:        ctx,
				podBouncer: mockPodBouncer,
				translator: mockTranslator,
			}

			inputSnap := input.NewInputSnapshotManualBuilder("hello").
				Build()

			mockTranslator.EXPECT().
				ShouldProcess(ctx, issuedCert).
				Return(false)

			err := reconciler.reconcileIssuedCertificate(issuedCert, inputSnap, mockOutput)
			Expect(err).NotTo(HaveOccurred())
			Expect(issuedCert.Status.State).To(Equal(certificatesv1.IssuedCertificateStatus_FINISHED))
		})

	})

	Context("IssuedCertificateRequested", func() {
		var (
			issuedCert *certificatesv1.IssuedCertificate
			csr        *certificatesv1.CertificateRequest
			csrBytes   []byte
		)
		BeforeEach(func() {
			issuedCert = &certificatesv1.IssuedCertificate{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "hello",
					Namespace:  "world",
					Generation: 2,
				},
				Spec: certificatesv1.IssuedCertificateSpec{},
				Status: certificatesv1.IssuedCertificateStatus{
					State:              certificatesv1.IssuedCertificateStatus_REQUESTED,
					ObservedGeneration: 2,
				},
			}

			csrBytes = []byte("I'm a CSR")

			csr = &certificatesv1.CertificateRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      issuedCert.GetName(),
					Namespace: issuedCert.GetNamespace(),
					Labels: map[string]string{
						"agent.certificates.mesh.gloo.solo.io": "gloo-mesh",
					},
				},
				Spec: certificatesv1.CertificateRequestSpec{
					CertificateSigningRequest: csrBytes,
				},
			}
		})

		It("Find CSR and pass into translator when cert is requested", func() {

			reconciler := &certAgentReconciler{
				ctx:        ctx,
				podBouncer: mockPodBouncer,
				translator: mockTranslator,
			}

			inputSnap := input.NewInputSnapshotManualBuilder("hello").
				AddCertificateRequests([]*certificatesv1.CertificateRequest{csr}).
				Build()

			mockTranslator.EXPECT().
				ShouldProcess(gomock.Any(), issuedCert).
				Return(true)

			mockTranslator.EXPECT().
				IssuedCertificateRequested(gomock.Any(), issuedCert, csr, inputSnap, mockOutput).
				Return(false, nil)

			err := reconciler.reconcileIssuedCertificate(issuedCert, inputSnap, mockOutput)
			Expect(err).NotTo(HaveOccurred())
			Expect(issuedCert.Status.State).To(Equal(certificatesv1.IssuedCertificateStatus_ISSUED))
		})

		It("Will not update status when translator.ShouldProcess == false", func() {

			reconciler := &certAgentReconciler{
				ctx:        ctx,
				podBouncer: mockPodBouncer,
				translator: mockTranslator,
			}

			inputSnap := input.NewInputSnapshotManualBuilder("hello").
				AddCertificateRequests([]*certificatesv1.CertificateRequest{csr}).
				Build()

			mockTranslator.EXPECT().
				ShouldProcess(gomock.Any(), issuedCert).
				Return(false)

			err := reconciler.reconcileIssuedCertificate(issuedCert, inputSnap, mockOutput)
			Expect(err).NotTo(HaveOccurred())
			Expect(issuedCert.Status.State).To(Equal(certificatesv1.IssuedCertificateStatus_REQUESTED))
		})
	})

	Context("IssuedCertificateIssued", func() {
		var (
			issuedCert *certificatesv1.IssuedCertificate
			pbd        *certificatesv1.PodBounceDirective
		)

		BeforeEach(func() {
			pbd = &certificatesv1.PodBounceDirective{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hello",
					Namespace: "world",
				},
				Spec: certificatesv1.PodBounceDirectiveSpec{},
			}

			issuedCert = &certificatesv1.IssuedCertificate{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "hello",
					Namespace:  "world",
					Generation: 2,
				},
				Spec: certificatesv1.IssuedCertificateSpec{
					PodBounceDirective: ezkube.MakeObjectRef(pbd),
				},
				Status: certificatesv1.IssuedCertificateStatus{
					State:              certificatesv1.IssuedCertificateStatus_ISSUED,
					ObservedGeneration: 2,
				},
			}
		})

		It("Will delete pods when cert has been issued", func() {

			reconciler := &certAgentReconciler{
				ctx:        ctx,
				podBouncer: mockPodBouncer,
				translator: mockTranslator,
			}

			pods := v1sets.NewPodSet(&corev1.Pod{})
			configMaps := v1sets.NewConfigMapSet(&corev1.ConfigMap{})
			secrets := v1sets.NewSecretSet(&corev1.Secret{})

			inputSnap := input.NewInputSnapshotManualBuilder("hello").
				AddPodBounceDirectives([]*certificatesv1.PodBounceDirective{pbd}).
				AddPods(pods.List()).
				AddSecrets(secrets.List()).
				AddConfigMaps(configMaps.List()).
				Build()

			mockTranslator.EXPECT().
				ShouldProcess(gomock.Any(), issuedCert).
				Return(true)

			mockTranslator.EXPECT().
				IssuedCertificateIssued(gomock.Any(), issuedCert, inputSnap, mockOutput).
				Return(nil)

			mockPodBouncer.EXPECT().
				BouncePods(gomock.Any(), pbd, pods, configMaps, secrets).
				Return(false, nil)

			err := reconciler.reconcileIssuedCertificate(issuedCert, inputSnap, mockOutput)
			Expect(err).NotTo(HaveOccurred())
			Expect(issuedCert.Status.State).To(Equal(certificatesv1.IssuedCertificateStatus_FINISHED))
		})

		It("Will not delete pods when translator.ShouldProcess==false", func() {

			reconciler := &certAgentReconciler{
				ctx:        ctx,
				podBouncer: mockPodBouncer,
				translator: mockTranslator,
			}

			pods := v1sets.NewPodSet(&corev1.Pod{})
			configMaps := v1sets.NewConfigMapSet(&corev1.ConfigMap{})
			secrets := v1sets.NewSecretSet(&corev1.Secret{})

			inputSnap := input.NewInputSnapshotManualBuilder("hello").
				AddPodBounceDirectives([]*certificatesv1.PodBounceDirective{pbd}).
				AddPods(pods.List()).
				AddSecrets(secrets.List()).
				AddConfigMaps(configMaps.List()).
				Build()

			mockTranslator.EXPECT().
				ShouldProcess(gomock.Any(), issuedCert).
				Return(false)

			err := reconciler.reconcileIssuedCertificate(issuedCert, inputSnap, mockOutput)
			Expect(err).NotTo(HaveOccurred())
			Expect(issuedCert.Status.State).To(Equal(certificatesv1.IssuedCertificateStatus_ISSUED))
		})
	})

	Context("IssuedCertificateFinished", func() {
		var (
			issuedCert       *certificatesv1.IssuedCertificate
			issuedCertSecret *corev1.Secret
		)

		BeforeEach(func() {

			issuedCertSecret = &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "issued",
					Namespace: "cert",
				},
			}

			issuedCert = &certificatesv1.IssuedCertificate{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "hello",
					Namespace:  "world",
					Generation: 2,
				},
				Spec: certificatesv1.IssuedCertificateSpec{
					IssuedCertificateSecret: ezkube.MakeObjectRef(issuedCertSecret),
				},
				Status: certificatesv1.IssuedCertificateStatus{
					State:              certificatesv1.IssuedCertificateStatus_FINISHED,
					ObservedGeneration: 2,
				},
			}
		})

		It("Will do nothing if no error happens", func() {

			reconciler := &certAgentReconciler{
				ctx:        ctx,
				podBouncer: mockPodBouncer,
				translator: mockTranslator,
			}

			pods := v1sets.NewPodSet(&corev1.Pod{})
			configMaps := v1sets.NewConfigMapSet(&corev1.ConfigMap{})
			secrets := v1sets.NewSecretSet(issuedCertSecret)

			inputSnap := input.NewInputSnapshotManualBuilder("hello").
				AddPods(pods.List()).
				AddSecrets(secrets.List()).
				AddConfigMaps(configMaps.List()).
				Build()

			mockTranslator.EXPECT().
				ShouldProcess(gomock.Any(), issuedCert).
				Return(true)

			mockTranslator.EXPECT().
				IssuedCertificateFinished(gomock.Any(), issuedCert, inputSnap, mockOutput).
				Return(nil)

			err := reconciler.reconcileIssuedCertificate(issuedCert, inputSnap, mockOutput)
			Expect(err).NotTo(HaveOccurred())
			Expect(issuedCert.Status.State).To(Equal(certificatesv1.IssuedCertificateStatus_FINISHED))
		})

		It("Will set status to failed if error happens", func() {

			reconciler := &certAgentReconciler{
				ctx:        ctx,
				podBouncer: mockPodBouncer,
				translator: mockTranslator,
			}

			pods := v1sets.NewPodSet(&corev1.Pod{})
			configMaps := v1sets.NewConfigMapSet(&corev1.ConfigMap{})
			secrets := v1sets.NewSecretSet(&corev1.Secret{})

			inputSnap := input.NewInputSnapshotManualBuilder("hello").
				AddPods(pods.List()).
				AddSecrets(secrets.List()).
				AddConfigMaps(configMaps.List()).
				Build()

			mockTranslator.EXPECT().
				ShouldProcess(gomock.Any(), issuedCert).
				Return(true)

			mockTranslator.EXPECT().
				IssuedCertificateFinished(gomock.Any(), issuedCert, inputSnap, mockOutput).
				Return(eris.New("hello"))

			err := reconciler.reconcileIssuedCertificate(issuedCert, inputSnap, mockOutput)
			Expect(err).NotTo(HaveOccurred())
			Expect(issuedCert.Status.State).To(Equal(certificatesv1.IssuedCertificateStatus_FAILED))
		})

	})
})
