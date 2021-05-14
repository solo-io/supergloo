package mtls_test

import (
	"context"
	"fmt"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1sets "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/sets"
	certificatesv1 "github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/v1"
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	mock_istio "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/output/istio/mocks"
	mock_local "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/output/local/mocks"
	networkingv1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	mock_reporting "github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting/mocks"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/mesh/mtls"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/metautils"
	skv2corev1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("MtlsTranslator", func() {
	var (
		ctrl *gomock.Controller
		ctx  context.Context

		mockIstioBuilder *mock_istio.MockBuilder
		mockLocalBuilder *mock_local.MockBuilder
		mockReporter     *mock_reporting.MockReporter

		istioMesh         *discoveryv1.Mesh
		childResourceMeta *metav1.ObjectMeta
	)

	BeforeEach(func() {
		ctrl, ctx = gomock.WithContext(context.Background(), GinkgoT())
		mockIstioBuilder = mock_istio.NewMockBuilder(ctrl)
		mockLocalBuilder = mock_local.NewMockBuilder(ctrl)
		mockReporter = mock_reporting.NewMockReporter(ctrl)

		istioMesh = &discoveryv1.Mesh{
			ObjectMeta: metav1.ObjectMeta{
				Name: "my-istio-mesh",
			},
			Spec: discoveryv1.MeshSpec{
				Type: &discoveryv1.MeshSpec_Istio_{
					Istio: &discoveryv1.MeshSpec_Istio{
						TrustDomain:          "cluster.not-local",
						IstiodServiceAccount: "istiod-not-standard",
						Installation: &discoveryv1.MeshSpec_MeshInstallation{
							Namespace: "istio-system-2",
							Cluster:   "cluster-name",
						},
					},
				},
				AgentInfo: &discoveryv1.MeshSpec_AgentInfo{
					AgentNamespace: "gloo-mesh",
				},
			},
		}

		childResourceMeta = &metav1.ObjectMeta{
			Name:        istioMesh.GetName(),
			Namespace:   istioMesh.Spec.GetAgentInfo().GetAgentNamespace(),
			ClusterName: istioMesh.Spec.GetIstio().GetInstallation().GetCluster(),
			Labels:      metautils.TranslatedObjectLabels(),
		}

	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("will skip if non-istio mesh", func() {
		translator := mtls.NewTranslator(ctx, nil, nil)
		mesh := &discoveryv1.Mesh{}
		vm := &discoveryv1.MeshStatus_AppliedVirtualMesh{}
		translator.Translate(mesh, vm, mockIstioBuilder, mockLocalBuilder, mockReporter)
	})

	It("generated root CA", func() {
		certSettings := &certificatesv1.CommonCertOptions{
			OrgName: "my-org",
		}

		vm := &discoveryv1.MeshStatus_AppliedVirtualMesh{
			Ref: &skv2corev1.ObjectRef{
				Name:      "my-vm",
				Namespace: "gloo-mesh",
			},
			Spec: &networkingv1.VirtualMeshSpec{
				MtlsConfig: &networkingv1.VirtualMeshSpec_MTLSConfig{
					TrustModel: &networkingv1.VirtualMeshSpec_MTLSConfig_Shared{
						Shared: &networkingv1.SharedTrust{
							CertificateAuthority: &networkingv1.SharedTrust_RootCertificateAuthority{
								RootCertificateAuthority: &networkingv1.RootCertificateAuthority{
									CaSource: &networkingv1.RootCertificateAuthority_Generated{
										Generated: certSettings,
									},
								},
							},
						},
					},
				},
			},
		}

		// Can't actually check against the exact context
		mockLocalBuilder.EXPECT().AddSecrets(gomock.Any()).Do(func(secret *corev1.Secret) {
			// Make sure the name is the same, maybe decode the cert and check the data?
			Expect(secret.GetName()).To(Equal(vm.GetRef().GetName() + "." + vm.GetRef().GetNamespace()))
			Expect(mtls.IsSigningCert(secret)).To(BeTrue())
			for k, v := range secret.Data {
				fmt.Println(k)
				fmt.Println(string(v))
			}
		})

		mockIstioBuilder.EXPECT().
			AddIssuedCertificates(gomock.Any()).
			Do(func(issuedCert *certificatesv1.IssuedCertificate) {
				cert := &certificatesv1.IssuedCertificate{
					ObjectMeta: *childResourceMeta,
					Spec: certificatesv1.IssuedCertificateSpec{
						Hosts: []string{"spiffe://cluster.not-local/ns/istio-system-2/sa/istiod-not-standard"},
						Org:   "Istio",
						CertificateAuthority: &certificatesv1.IssuedCertificateSpec_GlooMeshCa{
							GlooMeshCa: &certificatesv1.RootCertificateAuthority{
								CertificateAuthority: &certificatesv1.RootCertificateAuthority_SigningCertificateSecret{
									SigningCertificateSecret: &skv2corev1.ObjectRef{
										Name:      vm.GetRef().GetName() + "." + vm.GetRef().GetNamespace(),
										Namespace: "gloo-mesh",
									},
								},
							},
						},
						IssuedCertificateSecret: &skv2corev1.ObjectRef{
							Name:      "cacerts",
							Namespace: istioMesh.Spec.GetIstio().GetInstallation().GetNamespace(),
						},
					},
				}
				metautils.AppendParent(ctx, cert, vm.GetRef(), networkingv1.VirtualMesh{}.GVK())
				Expect(cert).To(Equal(issuedCert))
			})

		mockIstioBuilder.EXPECT().AddPodBounceDirectives(nil)

		translator := mtls.NewTranslator(ctx, v1sets.NewSecretSet(), nil)

		translator.Translate(istioMesh, vm, mockIstioBuilder, mockLocalBuilder, mockReporter)
	})

	It("provided root CA", func() {

		generatedSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-secret",
				Namespace: "my-namespace",
			},
		}

		vm := &discoveryv1.MeshStatus_AppliedVirtualMesh{
			Ref: &skv2corev1.ObjectRef{
				Name:      "my-vm",
				Namespace: "gloo-mesh",
			},
			Spec: &networkingv1.VirtualMeshSpec{
				MtlsConfig: &networkingv1.VirtualMeshSpec_MTLSConfig{
					TrustModel: &networkingv1.VirtualMeshSpec_MTLSConfig_Shared{
						Shared: &networkingv1.SharedTrust{
							CertificateAuthority: &networkingv1.SharedTrust_RootCertificateAuthority{
								RootCertificateAuthority: &networkingv1.RootCertificateAuthority{
									CaSource: &networkingv1.RootCertificateAuthority_Secret{
										Secret: &skv2corev1.ObjectRef{
											Name:      generatedSecret.GetName(),
											Namespace: generatedSecret.GetNamespace(),
										},
									},
								},
							},
						},
					},
				},
			},
		}

		mockIstioBuilder.EXPECT().
			AddIssuedCertificates(gomock.Any()).
			Do(func(issuedCert *certificatesv1.IssuedCertificate) {
				cert := &certificatesv1.IssuedCertificate{
					ObjectMeta: *childResourceMeta,
					Spec: certificatesv1.IssuedCertificateSpec{
						Hosts: []string{"spiffe://cluster.not-local/ns/istio-system-2/sa/istiod-not-standard"},
						Org:   "Istio",
						CertificateAuthority: &certificatesv1.IssuedCertificateSpec_GlooMeshCa{
							GlooMeshCa: &certificatesv1.RootCertificateAuthority{
								CertificateAuthority: &certificatesv1.RootCertificateAuthority_SigningCertificateSecret{
									SigningCertificateSecret: &skv2corev1.ObjectRef{
										Name:      generatedSecret.GetName(),
										Namespace: generatedSecret.GetNamespace(),
									},
								},
							},
						},
						IssuedCertificateSecret: &skv2corev1.ObjectRef{
							Name:      "cacerts",
							Namespace: istioMesh.Spec.GetIstio().GetInstallation().GetNamespace(),
						},
					},
				}
				metautils.AppendParent(ctx, cert, vm.GetRef(), networkingv1.VirtualMesh{}.GVK())
				Expect(cert).To(Equal(issuedCert))
			})

		mockIstioBuilder.EXPECT().AddPodBounceDirectives(nil)

		translator := mtls.NewTranslator(ctx, v1sets.NewSecretSet(generatedSecret), nil)

		translator.Translate(istioMesh, vm, mockIstioBuilder, mockLocalBuilder, mockReporter)
	})

})
