package cert_signer_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/contextutils"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_security "github.com/solo-io/service-mesh-hub/pkg/api/security.smh.solo.io/v1alpha1"
	smh_security_types "github.com/solo-io/service-mesh-hub/pkg/api/security.smh.solo.io/v1alpha1/types"
	mock_certgen "github.com/solo-io/service-mesh-hub/pkg/csr/certgen/mocks"
	cert_secrets "github.com/solo-io/service-mesh-hub/pkg/csr/certgen/secrets"
	cert_signer "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/security/cert-signer"
	mock_cert_signer "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/security/cert-signer/mocks"
	. "github.com/solo-io/service-mesh-hub/test/logging"
	mock_security_config "github.com/solo-io/service-mesh-hub/test/mocks/clients/security.smh.solo.io/v1alpha1"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("csr processor", func() {
	var (
		ctrl         *gomock.Controller
		ctx          context.Context
		testLogger   *TestLogger
		mgCertClient *mock_cert_signer.MockVirtualMeshCertClient
		csrClient    *mock_security_config.MockVirtualMeshCertificateSigningRequestClient
		signer       *mock_certgen.MockSigner
		csrProcessor cert_signer.VirtualMeshCSRSigner

		testErr = eris.New("hello")
	)

	BeforeEach(func() {
		testLogger = NewTestLogger()
		ctx = contextutils.WithExistingLogger(context.TODO(), testLogger.Logger())
		ctrl = gomock.NewController(GinkgoT())
		mgCertClient = mock_cert_signer.NewMockVirtualMeshCertClient(ctrl)
		csrClient = mock_security_config.NewMockVirtualMeshCertificateSigningRequestClient(ctrl)
		signer = mock_certgen.NewMockSigner(ctrl)
		csrProcessor = cert_signer.NewVirtualMeshCSRSigner(mgCertClient, csrClient, signer)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("should process", func() {
		It("will return false if cert is nil and request is denied", func() {
			obj := &smh_security.VirtualMeshCertificateSigningRequest{
				Spec: smh_security_types.VirtualMeshCertificateSigningRequestSpec{},
				Status: smh_security_types.VirtualMeshCertificateSigningRequestStatus{
					ThirdPartyApproval: &smh_security_types.VirtualMeshCertificateSigningRequestStatus_ThirdPartyApprovalWorkflow{
						ApprovalStatus: smh_security_types.VirtualMeshCertificateSigningRequestStatus_ThirdPartyApprovalWorkflow_DENIED,
					},
				},
			}
			Expect(csrProcessor.Sign(ctx, obj)).To(BeNil())
		})

		It("will return false if CSR data has len 0", func() {
			obj := &smh_security.VirtualMeshCertificateSigningRequest{
				Spec: smh_security_types.VirtualMeshCertificateSigningRequestSpec{},
				Status: smh_security_types.VirtualMeshCertificateSigningRequestStatus{
					ThirdPartyApproval: &smh_security_types.VirtualMeshCertificateSigningRequestStatus_ThirdPartyApprovalWorkflow{
						ApprovalStatus: smh_security_types.VirtualMeshCertificateSigningRequestStatus_ThirdPartyApprovalWorkflow_APPROVED,
					},
				},
			}
			Expect(csrProcessor.Sign(ctx, obj)).To(BeNil())
		})

		It("will return false if virtual mesh is nil", func() {
			obj := &smh_security.VirtualMeshCertificateSigningRequest{
				Spec: smh_security_types.VirtualMeshCertificateSigningRequestSpec{
					CsrData: []byte("hello"),
				},
				Status: smh_security_types.VirtualMeshCertificateSigningRequestStatus{
					ThirdPartyApproval: &smh_security_types.VirtualMeshCertificateSigningRequestStatus_ThirdPartyApprovalWorkflow{
						ApprovalStatus: smh_security_types.VirtualMeshCertificateSigningRequestStatus_ThirdPartyApprovalWorkflow_APPROVED,
					},
				},
			}
			Expect(csrProcessor.Sign(ctx, obj)).To(BeNil())
		})

		It("will return false if certs are filled in ", func() {
			obj := &smh_security.VirtualMeshCertificateSigningRequest{
				Spec: smh_security_types.VirtualMeshCertificateSigningRequestSpec{
					CsrData:        []byte("hello"),
					VirtualMeshRef: &smh_core_types.ResourceRef{},
				},
				Status: smh_security_types.VirtualMeshCertificateSigningRequestStatus{
					ThirdPartyApproval: &smh_security_types.VirtualMeshCertificateSigningRequestStatus_ThirdPartyApprovalWorkflow{
						ApprovalStatus: smh_security_types.VirtualMeshCertificateSigningRequestStatus_ThirdPartyApprovalWorkflow_APPROVED,
					},
					Response: &smh_security_types.VirtualMeshCertificateSigningRequestStatus_Response{
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
			csr := &smh_security.VirtualMeshCertificateSigningRequest{
				Spec: smh_security_types.VirtualMeshCertificateSigningRequestSpec{
					VirtualMeshRef: &smh_core_types.ResourceRef{},
					CsrData:        []byte("csr-data"),
				},
			}

			mgCertClient.EXPECT().
				GetRootCaBundle(ctx, csr.Spec.GetVirtualMeshRef()).
				Return(nil, testErr)

			status := csrProcessor.Sign(ctx, csr)
			Expect(status).To(Equal(&smh_security_types.VirtualMeshCertificateSigningRequestStatus{
				ComputedStatus: &smh_core_types.Status{
					State:   smh_core_types.Status_INVALID,
					Message: cert_signer.VirtualMeshTrustBundleNotFoundMsg(testErr, csr.Spec.GetVirtualMeshRef()).Error(),
				},
			}))
		})

		It("will return an error if cert cannot be generated from CSR", func() {
			csr := &smh_security.VirtualMeshCertificateSigningRequest{
				TypeMeta:   k8s_meta_types.TypeMeta{},
				ObjectMeta: k8s_meta_types.ObjectMeta{},
				Spec: smh_security_types.VirtualMeshCertificateSigningRequestSpec{
					VirtualMeshRef: &smh_core_types.ResourceRef{},
					CsrData:        []byte("csr-data"),
				},
				Status: smh_security_types.VirtualMeshCertificateSigningRequestStatus{},
			}

			rootCaData := &cert_secrets.RootCAData{
				PrivateKey: []byte("private-key"),
				RootCert:   []byte("root-cert"),
			}

			mgCertClient.EXPECT().
				GetRootCaBundle(ctx, csr.Spec.GetVirtualMeshRef()).
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
			Expect(status).To(Equal(&smh_security_types.VirtualMeshCertificateSigningRequestStatus{
				ComputedStatus: &smh_core_types.Status{
					State:   smh_core_types.Status_INVALID,
					Message: cert_signer.FailedToSignCertError(testErr).Error(),
				},
			}))
		})

		It("will return an error if cert cannot be generated from CSR", func() {
			csr := &smh_security.VirtualMeshCertificateSigningRequest{
				TypeMeta:   k8s_meta_types.TypeMeta{},
				ObjectMeta: k8s_meta_types.ObjectMeta{},
				Spec: smh_security_types.VirtualMeshCertificateSigningRequestSpec{
					VirtualMeshRef: &smh_core_types.ResourceRef{},
					CsrData:        []byte("csr-data"),
				},
				Status: smh_security_types.VirtualMeshCertificateSigningRequestStatus{},
			}

			rootCaData := &cert_secrets.RootCAData{
				PrivateKey: []byte("private-key"),
				RootCert:   []byte("root-cert"),
			}

			mgCertClient.EXPECT().
				GetRootCaBundle(ctx, csr.Spec.GetVirtualMeshRef()).
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
			Expect(status).To(Equal(&smh_security_types.VirtualMeshCertificateSigningRequestStatus{
				Response: &smh_security_types.VirtualMeshCertificateSigningRequestStatus_Response{
					CaCertificate:   newCert,
					RootCertificate: rootCaData.RootCert,
				},
				ComputedStatus: &smh_core_types.Status{
					State: smh_core_types.Status_ACCEPTED,
				},
			}))
		})
	})

})
