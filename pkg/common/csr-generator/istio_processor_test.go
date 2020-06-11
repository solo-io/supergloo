package csr_generator_test

import (
	"context"
	"strings"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_security "github.com/solo-io/service-mesh-hub/pkg/api/security.smh.solo.io/v1alpha1"
	smh_security_types "github.com/solo-io/service-mesh-hub/pkg/api/security.smh.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/common/csr/certgen"
	mock_certgen "github.com/solo-io/service-mesh-hub/pkg/common/csr/certgen/mocks"
	cert_secrets "github.com/solo-io/service-mesh-hub/pkg/common/csr/certgen/secrets"
	csr_generator "github.com/solo-io/service-mesh-hub/pkg/common/csr-generator"
	mock_csr_agent_controller "github.com/solo-io/service-mesh-hub/pkg/common/csr-generator/mocks"
	mock_kubernetes_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/kubernetes/core/v1"
	mock_security_config "github.com/solo-io/service-mesh-hub/test/mocks/clients/security.smh.solo.io/v1alpha1"
	pki_util "istio.io/istio/security/pkg/pki/util"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("csr processor", func() {
	var (
		ctrl              *gomock.Controller
		ctx               context.Context
		csrClient         *mock_security_config.MockVirtualMeshCertificateSigningRequestClient
		secretClient      *mock_kubernetes_core.MockSecretClient
		certClient        *mock_csr_agent_controller.MockCertClient
		signer            *mock_certgen.MockSigner
		istioCsrGenerator csr_generator.IstioCSRGenerator

		testErr = eris.New("hello")
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		csrClient = mock_security_config.NewMockVirtualMeshCertificateSigningRequestClient(ctrl)
		secretClient = mock_kubernetes_core.NewMockSecretClient(ctrl)
		certClient = mock_csr_agent_controller.NewMockCertClient(ctrl)
		signer = mock_certgen.NewMockSigner(ctrl)
		istioCsrGenerator = csr_generator.NewIstioCSRGenerator(csrClient, secretClient, certClient, signer)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("will return an error if secret key cannot be ensured", func() {
		csr := &smh_security.VirtualMeshCertificateSigningRequest{}
		certClient.EXPECT().
			EnsureSecretKey(ctx, csr).
			Return(nil, testErr)

		status := istioCsrGenerator.GenerateIstioCSR(ctx, csr)
		Expect(status).To(Equal(&smh_security_types.VirtualMeshCertificateSigningRequestStatus{
			ComputedStatus: &smh_core_types.Status{
				State:   smh_core_types.Status_INVALID,
				Message: csr_generator.FailedToRetrievePrivateKeyError(testErr).Error(),
			},
		}))
	})

	Context("no csr", func() {

		It("will return an error if csr cannot be generated", func() {
			csr := &smh_security.VirtualMeshCertificateSigningRequest{
				Spec: smh_security_types.VirtualMeshCertificateSigningRequestSpec{
					CertConfig: &smh_security_types.VirtualMeshCertificateSigningRequestSpec_CertConfig{
						Hosts: []string{"host1", "host2"},
						Org:   "Istio",
					},
				},
			}
			certData := &cert_secrets.IntermediateCAData{
				CaPrivateKey: []byte("ca-key"),
			}

			certClient.EXPECT().
				EnsureSecretKey(ctx, csr).
				Return(certData, nil)

			signer.EXPECT().
				GenCSRWithKey(pki_util.CertOptions{
					SignerPrivPem: certData.CaPrivateKey,
					Org:           csr.Spec.GetCertConfig().GetOrg(),
					Host:          strings.Join(csr.Spec.GetCertConfig().GetHosts(), ","),
				}).
				Return(nil, testErr)

			status := istioCsrGenerator.GenerateIstioCSR(ctx, csr)
			Expect(status).To(Equal(&smh_security_types.VirtualMeshCertificateSigningRequestStatus{
				ComputedStatus: &smh_core_types.Status{
					State:   smh_core_types.Status_INVALID,
					Message: csr_generator.FailedToGenerateCSRError(testErr).Error(),
				},
			}))
		})

		It("will return an error mgcsr cannot be updated with csr bytes", func() {
			csr := &smh_security.VirtualMeshCertificateSigningRequest{
				Spec: smh_security_types.VirtualMeshCertificateSigningRequestSpec{
					CertConfig: &smh_security_types.VirtualMeshCertificateSigningRequestSpec_CertConfig{
						Hosts: []string{"host1", "host2"},
						Org:   "Istio",
					},
				},
			}
			certData := &cert_secrets.IntermediateCAData{
				CaPrivateKey: []byte("ca-key"),
			}

			csrData := []byte("csr-data")

			certClient.EXPECT().
				EnsureSecretKey(ctx, csr).
				Return(certData, nil)

			signer.EXPECT().
				GenCSRWithKey(pki_util.CertOptions{
					SignerPrivPem: certData.CaPrivateKey,
					Org:           csr.Spec.GetCertConfig().GetOrg(),
					Host:          strings.Join(csr.Spec.GetCertConfig().GetHosts(), ","),
				}).
				Return(csrData, nil)

			matchCsr := csr.DeepCopy()
			matchCsr.Spec.CsrData = csrData
			csrClient.EXPECT().
				UpdateVirtualMeshCertificateSigningRequest(ctx, matchCsr).
				Return(testErr)

			status := istioCsrGenerator.GenerateIstioCSR(ctx, csr)
			Expect(status).To(Equal(&smh_security_types.VirtualMeshCertificateSigningRequestStatus{
				ComputedStatus: &smh_core_types.Status{
					State:   smh_core_types.Status_INVALID,
					Message: csr_generator.FailedToAddCsrToResource(testErr).Error(),
				},
			}))
		})

		It("will return an error mgcsr cannot be updated with csr bytes", func() {
			csr := &smh_security.VirtualMeshCertificateSigningRequest{
				Spec: smh_security_types.VirtualMeshCertificateSigningRequestSpec{
					CertConfig: &smh_security_types.VirtualMeshCertificateSigningRequestSpec_CertConfig{
						Hosts: []string{"host1", "host2"},
						Org:   "Istio",
					},
				},
			}
			certData := &cert_secrets.IntermediateCAData{
				CaPrivateKey: []byte("ca-key"),
			}

			csrData := []byte("csr-data")

			certClient.EXPECT().
				EnsureSecretKey(ctx, csr).
				Return(certData, nil)

			signer.EXPECT().
				GenCSRWithKey(pki_util.CertOptions{
					SignerPrivPem: certData.CaPrivateKey,
					Org:           csr.Spec.GetCertConfig().GetOrg(),
					Host:          strings.Join(csr.Spec.GetCertConfig().GetHosts(), ","),
				}).
				Return(csrData, nil)

			matchCsr := csr.DeepCopy()
			matchCsr.Spec.CsrData = csrData
			csrClient.EXPECT().
				UpdateVirtualMeshCertificateSigningRequest(ctx, matchCsr).
				Return(nil)

			status := istioCsrGenerator.GenerateIstioCSR(ctx, csr)
			Expect(status).To(Equal(&smh_security_types.VirtualMeshCertificateSigningRequestStatus{
				ComputedStatus: &smh_core_types.Status{
					State: smh_core_types.Status_ACCEPTED,
				},
			}))
		})

	})

	Context("with cert data", func() {

		It("will fail if istio CA cannot be updated", func() {
			csr := &smh_security.VirtualMeshCertificateSigningRequest{
				Spec: smh_security_types.VirtualMeshCertificateSigningRequestSpec{
					CsrData: []byte("csr-data"),
				},
				Status: smh_security_types.VirtualMeshCertificateSigningRequestStatus{
					Response: &smh_security_types.VirtualMeshCertificateSigningRequestStatus_Response{
						CaCertificate:   []byte("ca-cert"),
						RootCertificate: []byte("root-cert"),
					},
				},
			}
			certData := &cert_secrets.IntermediateCAData{}
			certClient.EXPECT().
				EnsureSecretKey(ctx, csr).
				Return(certData, nil)

			secretClient.EXPECT().
				GetSecret(ctx, client.ObjectKey{Name: csr_generator.IstioCaSecretName, Namespace: "istio-system"}).
				Return(nil, testErr)

			matchErr := csr_generator.FailedToUpdateCaError(testErr)
			status := istioCsrGenerator.GenerateIstioCSR(ctx, csr)
			Expect(status).To(Equal(&smh_security_types.VirtualMeshCertificateSigningRequestStatus{
				ComputedStatus: &smh_core_types.Status{
					State:   smh_core_types.Status_INVALID,
					Message: matchErr.Error(),
				},
				Response: csr.Status.GetResponse(),
			}))
		})

		It("will create a new resource if the secret doesn't exist", func() {
			csr := &smh_security.VirtualMeshCertificateSigningRequest{
				Spec: smh_security_types.VirtualMeshCertificateSigningRequestSpec{
					CsrData: []byte("csr-data"),
				},
				Status: smh_security_types.VirtualMeshCertificateSigningRequestStatus{
					Response: &smh_security_types.VirtualMeshCertificateSigningRequestStatus_Response{
						CaCertificate:   []byte("ca-cert"),
						RootCertificate: []byte("root-cert"),
					},
				},
			}
			certData := &cert_secrets.IntermediateCAData{}
			certClient.EXPECT().
				EnsureSecretKey(ctx, csr).
				Return(certData, nil)

			secretClient.EXPECT().
				GetSecret(ctx, client.ObjectKey{Name: csr_generator.IstioCaSecretName, Namespace: "istio-system"}).
				Return(nil, errors.NewNotFound(schema.GroupResource{}, "name"))

			certData.RootCert = csr.Status.GetResponse().GetRootCertificate()
			certData.CaCert = csr.Status.GetResponse().GetCaCertificate()
			certData.CertChain = certgen.AppendRootCerts(certData.CaCert, certData.RootCert)
			secretClient.EXPECT().
				CreateSecret(ctx, certData.BuildSecret(csr_generator.IstioCaSecretName, "istio-system")).
				Return(testErr)

			matchErr := csr_generator.FailedToUpdateCaError(testErr)
			status := istioCsrGenerator.GenerateIstioCSR(ctx, csr)
			Expect(status).To(Equal(&smh_security_types.VirtualMeshCertificateSigningRequestStatus{
				ComputedStatus: &smh_core_types.Status{
					State:   smh_core_types.Status_INVALID,
					Message: matchErr.Error(),
				},
				Response: csr.Status.GetResponse(),
			}))
		})

		It("won't try to update istio CA if it already exists, and the content is equal", func() {
			csr := &smh_security.VirtualMeshCertificateSigningRequest{
				Spec: smh_security_types.VirtualMeshCertificateSigningRequestSpec{
					CsrData: []byte("csr-data"),
				},
				Status: smh_security_types.VirtualMeshCertificateSigningRequestStatus{
					Response: &smh_security_types.VirtualMeshCertificateSigningRequestStatus_Response{
						CaCertificate:   []byte("ca-cert"),
						RootCertificate: []byte("root-cert"),
					},
				},
			}
			certData := &cert_secrets.IntermediateCAData{}
			certClient.EXPECT().
				EnsureSecretKey(ctx, csr).
				Return(certData, nil)

			certData.RootCert = csr.Status.GetResponse().GetRootCertificate()
			certData.CaCert = csr.Status.GetResponse().GetCaCertificate()
			certData.CertChain = certgen.AppendRootCerts(certData.CaCert, certData.RootCert)
			secretClient.EXPECT().
				GetSecret(ctx, client.ObjectKey{Name: csr_generator.IstioCaSecretName, Namespace: "istio-system"}).
				Return(certData.BuildSecret(csr_generator.IstioCaSecretName, "istio-system"), nil)

			status := istioCsrGenerator.GenerateIstioCSR(ctx, csr)
			Expect(status).To(Equal(&smh_security_types.VirtualMeshCertificateSigningRequestStatus{
				ComputedStatus: &smh_core_types.Status{
					State: smh_core_types.Status_ACCEPTED,
				},
				Response: csr.Status.GetResponse(),
			}))
		})

		It("will update cacerts with new data if it is different", func() {
			csr := &smh_security.VirtualMeshCertificateSigningRequest{
				Spec: smh_security_types.VirtualMeshCertificateSigningRequestSpec{
					CsrData: []byte("csr-data"),
				},
				Status: smh_security_types.VirtualMeshCertificateSigningRequestStatus{
					Response: &smh_security_types.VirtualMeshCertificateSigningRequestStatus_Response{
						CaCertificate:   []byte("ca-cert"),
						RootCertificate: []byte("root-cert"),
					},
				},
			}
			certData := &cert_secrets.IntermediateCAData{}
			certClient.EXPECT().
				EnsureSecretKey(ctx, csr).
				Return(certData, nil)

			certData.RootCert = []byte("new-root-cert")
			certData.CaCert = []byte("new-ca-cert")
			certData.CertChain = certgen.AppendRootCerts(certData.CaCert, certData.RootCert)

			secretClient.EXPECT().
				GetSecret(ctx, client.ObjectKey{Name: csr_generator.IstioCaSecretName, Namespace: "istio-system"}).
				Return(certData.BuildSecret(csr_generator.IstioCaSecretName, "istio-system"), nil)

			matchCert := &cert_secrets.IntermediateCAData{
				RootCAData: cert_secrets.RootCAData{
					RootCert: csr.Status.GetResponse().GetRootCertificate(),
				},
				CertChain: certgen.AppendRootCerts(
					csr.Status.GetResponse().GetCaCertificate(), csr.Status.GetResponse().GetRootCertificate(),
				),
				CaCert: csr.Status.GetResponse().GetCaCertificate(),
			}
			secretClient.EXPECT().
				UpdateSecret(ctx, matchCert.BuildSecret(csr_generator.IstioCaSecretName, "istio-system")).
				Return(nil)

			status := istioCsrGenerator.GenerateIstioCSR(ctx, csr)
			Expect(status).To(Equal(&smh_security_types.VirtualMeshCertificateSigningRequestStatus{
				ComputedStatus: &smh_core_types.Status{
					State: smh_core_types.Status_ACCEPTED,
				},
				Response: csr.Status.GetResponse(),
			}))
		})

	})

})
