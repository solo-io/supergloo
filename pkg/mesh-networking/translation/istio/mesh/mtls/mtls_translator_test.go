package mtls_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"

	. "github.com/onsi/gomega"
	v1sets "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/sets"
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	mock_istio "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/output/istio/mocks"
	mock_local "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/output/local/mocks"
	networkingv1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	mock_reporting "github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting/mocks"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/mesh/mtls"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("MtlsTranslator", func() {
	var (
		ctrl *gomock.Controller
		ctx  context.Context

		mockIstioBuilder *mock_istio.MockBuilder
		mockLocalBuilder *mock_local.MockBuilder
		mockReporter     *mock_reporting.MockReporter
	)

	BeforeEach(func() {
		ctrl, ctx = gomock.WithContext(context.Background(), GinkgoT())
		mockIstioBuilder = mock_istio.NewMockBuilder(ctrl)
		mockLocalBuilder = mock_local.NewMockBuilder(ctrl)
		mockReporter = mock_reporting.NewMockReporter(ctrl)
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
		certSettings := &networkingv1.VirtualMeshSpec_RootCertificateAuthority_SelfSignedCert{
			TtlDays:         364,
			RsaKeySizeBytes: 4097,
			OrgName:         "my-org",
		}
		mesh := &discoveryv1.Mesh{
			Spec: discoveryv1.MeshSpec{
				Type: &discoveryv1.MeshSpec_Istio_{
					Istio: &discoveryv1.MeshSpec_Istio{},
				},
			},
		}
		vm := &discoveryv1.MeshStatus_AppliedVirtualMesh{
			Spec: &networkingv1.VirtualMeshSpec{
				MtlsConfig: &networkingv1.VirtualMeshSpec_MTLSConfig{
					TrustModel: &networkingv1.VirtualMeshSpec_MTLSConfig_Shared{
						Shared: &networkingv1.VirtualMeshSpec_MTLSConfig_SharedTrust{
							RootCertificateAuthority: &networkingv1.VirtualMeshSpec_RootCertificateAuthority{
								CaSource: &networkingv1.VirtualMeshSpec_RootCertificateAuthority_Generated{
									Generated: certSettings,
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
			Expect(sets.Key(secret)).To(Equal(vm.GetRef().GetName() + "." + vm.GetRef().GetNamespace()))
			Expect(mtls.IsSigningCert(secret)).To(BeTrue())
		})

		translator := mtls.NewTranslator(ctx, v1sets.NewSecretSet(), nil)

		translator.Translate(mesh, vm, mockIstioBuilder, mockLocalBuilder, mockReporter)
	})

	It("provided root CA", func() {

		generatedSecret := &corev1.Secret{}

		mesh := &discoveryv1.Mesh{
			Spec: discoveryv1.MeshSpec{
				Type: &discoveryv1.MeshSpec_Istio_{
					Istio: &discoveryv1.MeshSpec_Istio{},
				},
			},
		}
		vm := &discoveryv1.MeshStatus_AppliedVirtualMesh{
			Spec: &networkingv1.VirtualMeshSpec{
				MtlsConfig: &networkingv1.VirtualMeshSpec_MTLSConfig{
					TrustModel: &networkingv1.VirtualMeshSpec_MTLSConfig_Shared{
						Shared: &networkingv1.VirtualMeshSpec_MTLSConfig_SharedTrust{
							RootCertificateAuthority: &networkingv1.VirtualMeshSpec_RootCertificateAuthority{
								CaSource: &networkingv1.VirtualMeshSpec_RootCertificateAuthority_Secret{
									Secret: &v1.ObjectRef{
										Name:      generatedSecret.GetName(),
										Namespace: generatedSecret.GetNamespace(),
									},
								},
							},
						},
					},
				},
			},
		}

		translator := mtls.NewTranslator(ctx, v1sets.NewSecretSet(generatedSecret), nil)

		translator.Translate(mesh, vm, mockIstioBuilder, mockLocalBuilder, mockReporter)
	})

})
