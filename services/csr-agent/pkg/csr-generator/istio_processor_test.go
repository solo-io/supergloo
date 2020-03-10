package csr_generator_test

import (
	"context"
	"strings"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	security_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1"
	security_types "github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1/types"
	mock_kubernetes_core "github.com/solo-io/mesh-projects/pkg/clients/kubernetes/core/mocks"
	mock_security_config "github.com/solo-io/mesh-projects/pkg/clients/zephyr/security/mocks"
	"github.com/solo-io/mesh-projects/pkg/security/certgen"
	mock_certgen "github.com/solo-io/mesh-projects/pkg/security/certgen/mocks"
	cert_secrets "github.com/solo-io/mesh-projects/pkg/security/secrets"
	csr_generator "github.com/solo-io/mesh-projects/services/csr-agent/pkg/csr-generator"
	mock_csr_agent_controller "github.com/solo-io/mesh-projects/services/csr-agent/pkg/csr-generator/mocks"
	pki_util "istio.io/istio/security/pkg/pki/util"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ = Describe("csr processor", func() {
	var (
		ctrl              *gomock.Controller
		ctx               context.Context
		csrClient         *mock_security_config.MockMeshGroupCSRClient
		secretClient      *mock_kubernetes_core.MockSecretsClient
		certClient        *mock_csr_agent_controller.MockCertClient
		signer            *mock_certgen.MockSigner
		istioCsrGenerator csr_generator.IstioCSRGenerator

		testErr = eris.New("hello")
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		csrClient = mock_security_config.NewMockMeshGroupCSRClient(ctrl)
		secretClient = mock_kubernetes_core.NewMockSecretsClient(ctrl)
		certClient = mock_csr_agent_controller.NewMockCertClient(ctrl)
		signer = mock_certgen.NewMockSigner(ctrl)
		istioCsrGenerator = csr_generator.NewIstioCSRGenerator(csrClient, secretClient, certClient, signer)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("will return an error if secret key cannot be ensured", func() {
		csr := &security_v1alpha1.MeshGroupCertificateSigningRequest{}
		certClient.EXPECT().
			EnsureSecretKey(ctx, csr).
			Return(nil, testErr)

		status := istioCsrGenerator.GenerateIstioCSR(ctx, csr)
		Expect(status).To(Equal(&security_types.MeshGroupCertificateSigningRequestStatus{
			ComputedStatus: &core_types.ComputedStatus{
				Status:  core_types.ComputedStatus_INVALID,
				Message: csr_generator.FailedToRetrievePrivateKeyError(testErr).Error(),
			},
		}))
	})

	Context("no csr", func() {

		It("will return an error if csr cannot be generated", func() {
			csr := &security_v1alpha1.MeshGroupCertificateSigningRequest{
				Spec: security_types.MeshGroupCertificateSigningRequestSpec{
					CertConfig: &security_types.CertConfig{
						Hosts: []string{"host1", "host2"},
						Org:   "Istio",
					},
				},
			}
			certData := &cert_secrets.RootCaData{
				CaPrivateKey: []byte("cs-key"),
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
			Expect(status).To(Equal(&security_types.MeshGroupCertificateSigningRequestStatus{
				ComputedStatus: &core_types.ComputedStatus{
					Status:  core_types.ComputedStatus_INVALID,
					Message: csr_generator.FailedToGenerateCSRError(testErr).Error(),
				},
			}))
		})

		It("will return an error mgcsr cannot be updated with csr bytes", func() {
			csr := &security_v1alpha1.MeshGroupCertificateSigningRequest{
				Spec: security_types.MeshGroupCertificateSigningRequestSpec{
					CertConfig: &security_types.CertConfig{
						Hosts: []string{"host1", "host2"},
						Org:   "Istio",
					},
				},
			}
			certData := &cert_secrets.RootCaData{
				CaPrivateKey: []byte("cs-key"),
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
				Update(ctx, matchCsr).
				Return(testErr)

			status := istioCsrGenerator.GenerateIstioCSR(ctx, csr)
			Expect(status).To(Equal(&security_types.MeshGroupCertificateSigningRequestStatus{
				ComputedStatus: &core_types.ComputedStatus{
					Status:  core_types.ComputedStatus_INVALID,
					Message: csr_generator.FailesToAddCsrToResource(testErr).Error(),
				},
			}))
		})

		It("will return an error mgcsr cannot be updated with csr bytes", func() {
			csr := &security_v1alpha1.MeshGroupCertificateSigningRequest{
				Spec: security_types.MeshGroupCertificateSigningRequestSpec{
					CertConfig: &security_types.CertConfig{
						Hosts: []string{"host1", "host2"},
						Org:   "Istio",
					},
				},
			}
			certData := &cert_secrets.RootCaData{
				CaPrivateKey: []byte("cs-key"),
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
				Update(ctx, matchCsr).
				Return(nil)

			status := istioCsrGenerator.GenerateIstioCSR(ctx, csr)
			Expect(status).To(Equal(&security_types.MeshGroupCertificateSigningRequestStatus{
				ComputedStatus: &core_types.ComputedStatus{
					Status: core_types.ComputedStatus_ACCEPTED,
				},
			}))
		})

	})

	Context("with cert data", func() {

		It("will fail if istio CA cannot be updated", func() {
			csr := &security_v1alpha1.MeshGroupCertificateSigningRequest{
				Spec: security_types.MeshGroupCertificateSigningRequestSpec{
					CsrData: []byte("csr-data"),
				},
				Status: security_types.MeshGroupCertificateSigningRequestStatus{
					Response: &security_types.MeshGroupCertificateSigningResponse{
						CaCertificate:   []byte("ca-cert"),
						RootCertificate: []byte("root-cert"),
					},
				},
			}
			certData := &cert_secrets.RootCaData{}
			certClient.EXPECT().
				EnsureSecretKey(ctx, csr).
				Return(certData, nil)

			secretClient.EXPECT().
				Get(ctx, csr_generator.IstioCaSecretName, "istio-system").
				Return(nil, testErr)

			matchErr := csr_generator.FailedToUpdateCaError(testErr)
			status := istioCsrGenerator.GenerateIstioCSR(ctx, csr)
			Expect(status).To(Equal(&security_types.MeshGroupCertificateSigningRequestStatus{
				ComputedStatus: &core_types.ComputedStatus{
					Status:  core_types.ComputedStatus_INVALID,
					Message: matchErr.Error(),
				},
				Response: csr.Status.GetResponse(),
			}))
		})

		It("will create a new resource if the secret doesn't exist", func() {
			csr := &security_v1alpha1.MeshGroupCertificateSigningRequest{
				Spec: security_types.MeshGroupCertificateSigningRequestSpec{
					CsrData: []byte("csr-data"),
				},
				Status: security_types.MeshGroupCertificateSigningRequestStatus{
					Response: &security_types.MeshGroupCertificateSigningResponse{
						CaCertificate:   []byte("ca-cert"),
						RootCertificate: []byte("root-cert"),
					},
				},
			}
			certData := &cert_secrets.RootCaData{}
			certClient.EXPECT().
				EnsureSecretKey(ctx, csr).
				Return(certData, nil)

			secretClient.EXPECT().
				Get(ctx, csr_generator.IstioCaSecretName, "istio-system").
				Return(nil, errors.NewNotFound(schema.GroupResource{}, "name"))

			certData.RootCert = csr.Status.GetResponse().GetRootCertificate()
			certData.CaCert = csr.Status.GetResponse().GetCaCertificate()
			certData.CertChain = certgen.AppendRootCerts(certData.CaCert, certData.RootCert)
			secretClient.EXPECT().
				Create(ctx, certData.BuildSecret(csr_generator.IstioCaSecretName, "istio-system")).
				Return(testErr)

			matchErr := csr_generator.FailedToUpdateCaError(testErr)
			status := istioCsrGenerator.GenerateIstioCSR(ctx, csr)
			Expect(status).To(Equal(&security_types.MeshGroupCertificateSigningRequestStatus{
				ComputedStatus: &core_types.ComputedStatus{
					Status:  core_types.ComputedStatus_INVALID,
					Message: matchErr.Error(),
				},
				Response: csr.Status.GetResponse(),
			}))
		})

		It("won't try to update istio CA if it already exists, and the content is equal", func() {
			csr := &security_v1alpha1.MeshGroupCertificateSigningRequest{
				Spec: security_types.MeshGroupCertificateSigningRequestSpec{
					CsrData: []byte("csr-data"),
				},
				Status: security_types.MeshGroupCertificateSigningRequestStatus{
					Response: &security_types.MeshGroupCertificateSigningResponse{
						CaCertificate:   []byte("ca-cert"),
						RootCertificate: []byte("root-cert"),
					},
				},
			}
			certData := &cert_secrets.RootCaData{}
			certClient.EXPECT().
				EnsureSecretKey(ctx, csr).
				Return(certData, nil)

			certData.RootCert = csr.Status.GetResponse().GetRootCertificate()
			certData.CaCert = csr.Status.GetResponse().GetCaCertificate()
			certData.CertChain = certgen.AppendRootCerts(certData.CaCert, certData.RootCert)
			secretClient.EXPECT().
				Get(ctx, csr_generator.IstioCaSecretName, "istio-system").
				Return(certData.BuildSecret(csr_generator.IstioCaSecretName, "istio-system"), nil)

			status := istioCsrGenerator.GenerateIstioCSR(ctx, csr)
			Expect(status).To(Equal(&security_types.MeshGroupCertificateSigningRequestStatus{
				ComputedStatus: &core_types.ComputedStatus{
					Status: core_types.ComputedStatus_ACCEPTED,
				},
				Response: csr.Status.GetResponse(),
			}))
		})

		It("will update cacerts with new data if it is different", func() {
			csr := &security_v1alpha1.MeshGroupCertificateSigningRequest{
				Spec: security_types.MeshGroupCertificateSigningRequestSpec{
					CsrData: []byte("csr-data"),
				},
				Status: security_types.MeshGroupCertificateSigningRequestStatus{
					Response: &security_types.MeshGroupCertificateSigningResponse{
						CaCertificate:   []byte("ca-cert"),
						RootCertificate: []byte("root-cert"),
					},
				},
			}
			certData := &cert_secrets.RootCaData{}
			certClient.EXPECT().
				EnsureSecretKey(ctx, csr).
				Return(certData, nil)

			certData.RootCert = []byte("new-root-cert")
			certData.CaCert = []byte("new-ca-cert")
			certData.CertChain = certgen.AppendRootCerts(certData.CaCert, certData.RootCert)

			secretClient.EXPECT().
				Get(ctx, csr_generator.IstioCaSecretName, "istio-system").
				Return(certData.BuildSecret(csr_generator.IstioCaSecretName, "istio-system"), nil)

			matchCert := &cert_secrets.RootCaData{
				CertAndKeyData: cert_secrets.CertAndKeyData{
					CertChain: certgen.AppendRootCerts(
						csr.Status.GetResponse().GetCaCertificate(), csr.Status.GetResponse().GetRootCertificate(),
					),
					RootCert: csr.Status.GetResponse().GetRootCertificate(),
				},
				CaCert: csr.Status.GetResponse().GetCaCertificate(),
			}
			secretClient.EXPECT().
				Update(ctx, matchCert.BuildSecret(csr_generator.IstioCaSecretName, "istio-system")).
				Return(nil)

			status := istioCsrGenerator.GenerateIstioCSR(ctx, csr)
			Expect(status).To(Equal(&security_types.MeshGroupCertificateSigningRequestStatus{
				ComputedStatus: &core_types.ComputedStatus{
					Status: core_types.ComputedStatus_ACCEPTED,
				},
				Response: csr.Status.GetResponse(),
			}))
		})

	})

})
