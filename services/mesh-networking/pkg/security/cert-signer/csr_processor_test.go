package cert_signer_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/contextutils"
	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1"
	security_types "github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1/types"
	mock_security_config "github.com/solo-io/mesh-projects/pkg/clients/zephyr/security/mocks"
	mock_certgen "github.com/solo-io/mesh-projects/pkg/security/certgen/mocks"
	cert_secrets "github.com/solo-io/mesh-projects/pkg/security/secrets"
	cert_signer "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/security/cert-signer"
	mock_cert_signer "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/security/cert-signer/mocks"
	. "github.com/solo-io/mesh-projects/test/logging"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("csr processor", func() {
	var (
		ctrl         *gomock.Controller
		ctx          context.Context
		testLogger   *TestLogger
		mgCertClient *mock_cert_signer.MockMeshGroupCertClient
		csrClient    *mock_security_config.MockMeshGroupCSRClient
		signer       *mock_certgen.MockSigner
		csrProcessor cert_signer.MeshGroupCSRSigner

		testErr = eris.New("hello")
	)

	BeforeEach(func() {
		testLogger = NewTestLogger()
		ctx = contextutils.WithExistingLogger(context.TODO(), testLogger.Logger())
		ctrl = gomock.NewController(GinkgoT())
		mgCertClient = mock_cert_signer.NewMockMeshGroupCertClient(ctrl)
		csrClient = mock_security_config.NewMockMeshGroupCSRClient(ctrl)
		signer = mock_certgen.NewMockSigner(ctrl)
		csrProcessor = cert_signer.NewMeshGroupCSRSigner(mgCertClient, csrClient, signer)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("should process", func() {
		It("will return false if cert is nil and request is denied", func() {
			obj := &v1alpha1.MeshGroupCertificateSigningRequest{
				Spec: security_types.MeshGroupCertificateSigningRequestSpec{},
				Status: security_types.MeshGroupCertificateSigningRequestStatus{
					ThirdPartyApproval: &security_types.ThirdPartyApprovalWorkflow{
						ApprovalStatus: security_types.ThirdPartyApprovalWorkflow_DENIED,
					},
				},
			}
			Expect(csrProcessor.Sign(ctx, obj)).To(BeNil())
		})

		It("will return false if CSR data has len 0", func() {
			obj := &v1alpha1.MeshGroupCertificateSigningRequest{
				Spec: security_types.MeshGroupCertificateSigningRequestSpec{},
				Status: security_types.MeshGroupCertificateSigningRequestStatus{
					ThirdPartyApproval: &security_types.ThirdPartyApprovalWorkflow{
						ApprovalStatus: security_types.ThirdPartyApprovalWorkflow_APPROVED,
					},
				},
			}
			Expect(csrProcessor.Sign(ctx, obj)).To(BeNil())
		})

		It("will return false if mesh group is nil", func() {
			obj := &v1alpha1.MeshGroupCertificateSigningRequest{
				Spec: security_types.MeshGroupCertificateSigningRequestSpec{
					CsrData: []byte("hello"),
				},
				Status: security_types.MeshGroupCertificateSigningRequestStatus{
					ThirdPartyApproval: &security_types.ThirdPartyApprovalWorkflow{
						ApprovalStatus: security_types.ThirdPartyApprovalWorkflow_APPROVED,
					},
				},
			}
			Expect(csrProcessor.Sign(ctx, obj)).To(BeNil())
		})

		It("will return false if certs are filled in ", func() {
			obj := &v1alpha1.MeshGroupCertificateSigningRequest{
				Spec: security_types.MeshGroupCertificateSigningRequestSpec{
					CsrData:      []byte("hello"),
					MeshGroupRef: &core_types.ResourceRef{},
				},
				Status: security_types.MeshGroupCertificateSigningRequestStatus{
					ThirdPartyApproval: &security_types.ThirdPartyApprovalWorkflow{
						ApprovalStatus: security_types.ThirdPartyApprovalWorkflow_APPROVED,
					},
					Response: &security_types.MeshGroupCertificateSigningResponse{
						CaCertificate:   []byte("hello"),
						RootCertificate: []byte("hello"),
					},
				},
			}

			Expect(csrProcessor.Sign(ctx, obj)).To(BeNil())
		})
	})

	Context("process", func() {

		It("will return an error if root ca bundle cannot be found", func() {
			csr := &v1alpha1.MeshGroupCertificateSigningRequest{
				Spec: security_types.MeshGroupCertificateSigningRequestSpec{
					MeshGroupRef: &core_types.ResourceRef{},
					CsrData:      []byte("csr-data"),
				},
			}

			mgCertClient.EXPECT().
				GetRootCaBundle(ctx, csr.Spec.GetMeshGroupRef()).
				Return(nil, testErr)

			status := csrProcessor.Sign(ctx, csr)
			Expect(status).To(Equal(&security_types.MeshGroupCertificateSigningRequestStatus{
				ComputedStatus: &core_types.ComputedStatus{
					Status:  core_types.ComputedStatus_INVALID,
					Message: cert_signer.MeshGroupTrustBundleNotFoundMsg(testErr, csr.Spec.GetMeshGroupRef()).Error(),
				},
			}))
		})

		It("will return an error if cert cannot be generated from CSR", func() {
			csr := &v1alpha1.MeshGroupCertificateSigningRequest{
				TypeMeta:   metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{},
				Spec: security_types.MeshGroupCertificateSigningRequestSpec{
					MeshGroupRef: &core_types.ResourceRef{},
					CsrData:      []byte("csr-data"),
				},
				Status: security_types.MeshGroupCertificateSigningRequestStatus{},
			}

			rootCaData := &cert_secrets.RootCaData{
				CertAndKeyData: cert_secrets.CertAndKeyData{
					CertChain:  nil,
					PrivateKey: []byte("private-key"),
					RootCert:   []byte("root-key"),
				},
			}

			mgCertClient.EXPECT().
				GetRootCaBundle(ctx, csr.Spec.GetMeshGroupRef()).
				Return(rootCaData, nil)

			signer.EXPECT().
				GenCertFromEncodedCSR(
					csr.Spec.GetCsrData(),
					rootCaData.RootCert,
					rootCaData.PrivateKey,
					nil,
					gomock.Any(),
					true,
				).Return(nil, testErr)

			status := csrProcessor.Sign(ctx, csr)
			Expect(status).To(Equal(&security_types.MeshGroupCertificateSigningRequestStatus{
				ComputedStatus: &core_types.ComputedStatus{
					Status:  core_types.ComputedStatus_INVALID,
					Message: cert_signer.FailedToSignCertError(testErr).Error(),
				},
			}))
		})

		It("will return an error if cert cannot be generated from CSR", func() {
			csr := &v1alpha1.MeshGroupCertificateSigningRequest{
				TypeMeta:   metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{},
				Spec: security_types.MeshGroupCertificateSigningRequestSpec{
					MeshGroupRef: &core_types.ResourceRef{},
					CsrData:      []byte("csr-data"),
				},
				Status: security_types.MeshGroupCertificateSigningRequestStatus{},
			}

			rootCaData := &cert_secrets.RootCaData{
				CertAndKeyData: cert_secrets.CertAndKeyData{
					CertChain:  nil,
					PrivateKey: []byte("private-key"),
					RootCert:   []byte("root-key"),
				},
			}

			mgCertClient.EXPECT().
				GetRootCaBundle(ctx, csr.Spec.GetMeshGroupRef()).
				Return(rootCaData, nil)

			newCert := []byte("new-cert")
			signer.EXPECT().
				GenCertFromEncodedCSR(
					csr.Spec.GetCsrData(),
					rootCaData.RootCert,
					rootCaData.PrivateKey,
					nil,
					gomock.Any(),
					true,
				).Return(newCert, nil)

			status := csrProcessor.Sign(ctx, csr)
			Expect(status).To(Equal(&security_types.MeshGroupCertificateSigningRequestStatus{
				Response: &security_types.MeshGroupCertificateSigningResponse{
					CaCertificate:   newCert,
					RootCertificate: rootCaData.RootCert,
				},
				ComputedStatus: &core_types.ComputedStatus{
					Status: core_types.ComputedStatus_ACCEPTED,
				},
			}))
		})
	})

})
