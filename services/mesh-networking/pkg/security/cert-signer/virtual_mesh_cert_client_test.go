package cert_signer_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	. "github.com/solo-io/go-utils/testutils"
	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	v1alpha1_types "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	mock_kubernetes_core "github.com/solo-io/mesh-projects/pkg/clients/kubernetes/core/mocks"
	mock_zephyr_networking "github.com/solo-io/mesh-projects/pkg/clients/zephyr/networking/mocks"
	"github.com/solo-io/mesh-projects/pkg/env"
	mock_certgen "github.com/solo-io/mesh-projects/pkg/security/certgen/mocks"
	cert_secrets "github.com/solo-io/mesh-projects/pkg/security/secrets"
	cert_signer "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/security/cert-signer"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("virtual mesh cert client", func() {
	var (
		ctrl                  *gomock.Controller
		ctx                   context.Context
		secretClient          *mock_kubernetes_core.MockSecretsClient
		virtualMeshClient     *mock_zephyr_networking.MockVirtualMeshClient
		virtualMeshCertClient cert_signer.VirtualMeshCertClient
		rootCertGenerator     *mock_certgen.MockRootCertGenerator
		testErr               = eris.New("hello")
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		secretClient = mock_kubernetes_core.NewMockSecretsClient(ctrl)
		virtualMeshClient = mock_zephyr_networking.NewMockVirtualMeshClient(ctrl)
		rootCertGenerator = mock_certgen.NewMockRootCertGenerator(ctrl)
		virtualMeshCertClient = cert_signer.NewVirtualMeshCertClient(secretClient, virtualMeshClient, rootCertGenerator)
	})

	It("will fail if virtualMesh cannot be found", func() {
		meshRef := &core_types.ResourceRef{
			Name:      "name",
			Namespace: "namespace",
		}
		virtualMeshClient.EXPECT().Get(ctx, meshRef.Name, meshRef.Namespace).Return(nil, testErr)
		_, err := virtualMeshCertClient.GetRootCaBundle(ctx, meshRef)
		Expect(err).To(HaveOccurred())
		Expect(err).To(HaveInErrorChain(testErr))
	})

	It("will use user provided trust bundle in vm if set", func() {
		meshRef := &core_types.ResourceRef{
			Name:      "name",
			Namespace: "namespace",
		}
		vm := &v1alpha1.VirtualMesh{
			ObjectMeta: metav1.ObjectMeta{
				Name:      meshRef.Name,
				Namespace: meshRef.Namespace,
			},
			Spec: v1alpha1_types.VirtualMeshSpec{
				CertificateAuthority: &v1alpha1_types.VirtualMeshSpec_CertificateAuthority{
					Type: &v1alpha1_types.VirtualMeshSpec_CertificateAuthority_Provided_{
						Provided: &v1alpha1_types.VirtualMeshSpec_CertificateAuthority_Provided{
							Certificate: &core_types.ResourceRef{
								Name:      "tb_name",
								Namespace: "tb_namespace",
							},
						},
					},
				},
			},
		}
		virtualMeshClient.EXPECT().Get(ctx, meshRef.Name, meshRef.Namespace).Return(vm, nil)
		secretClient.
			EXPECT().
			Get(ctx,
				vm.Spec.GetCertificateAuthority().GetProvided().GetCertificate().GetName(),
				vm.Spec.GetCertificateAuthority().GetProvided().GetCertificate().GetNamespace(),
			).
			Return(nil, testErr)
		_, err := virtualMeshCertClient.GetRootCaBundle(ctx, meshRef)
		Expect(err).To(HaveOccurred())
		Expect(err).To(HaveInErrorChain(testErr))
	})

	It("will return proper CA data", func() {
		meshRef := &core_types.ResourceRef{
			Name:      "name",
			Namespace: "namespace",
		}
		vm := &v1alpha1.VirtualMesh{
			ObjectMeta: metav1.ObjectMeta{
				Name:      meshRef.Name,
				Namespace: meshRef.Namespace,
			},
			Spec: v1alpha1_types.VirtualMeshSpec{
				CertificateAuthority: &v1alpha1_types.VirtualMeshSpec_CertificateAuthority{
					Type: &v1alpha1_types.VirtualMeshSpec_CertificateAuthority_Provided_{
						Provided: &v1alpha1_types.VirtualMeshSpec_CertificateAuthority_Provided{
							Certificate: &core_types.ResourceRef{
								Name:      "tb_name",
								Namespace: "tb_namespace",
							},
						},
					},
				},
			},
		}
		matchData := &cert_secrets.RootCAData{
			PrivateKey: []byte("private_key"),
			RootCert:   []byte("root_cert"),
		}
		matchSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      vm.Spec.GetCertificateAuthority().GetProvided().GetCertificate().GetName(),
				Namespace: vm.Spec.GetCertificateAuthority().GetProvided().GetCertificate().GetNamespace(),
			},
			Data: map[string][]byte{
				cert_secrets.RootCertID:       matchData.RootCert,
				cert_secrets.RootPrivateKeyID: matchData.PrivateKey,
			},
			Type: cert_secrets.RootCertSecretType,
		}
		virtualMeshClient.EXPECT().Get(ctx, meshRef.Name, meshRef.Namespace).Return(vm, nil)
		secretClient.
			EXPECT().
			Get(ctx,
				vm.Spec.GetCertificateAuthority().GetProvided().GetCertificate().GetName(),
				vm.Spec.GetCertificateAuthority().GetProvided().GetCertificate().GetNamespace(),
			).
			Return(matchSecret, nil)
		data, err := virtualMeshCertClient.GetRootCaBundle(ctx, meshRef)
		Expect(err).NotTo(HaveOccurred())
		Expect(data).To(Equal(matchData))
	})

	It("will create auto-generated root cert if CertificateAuthority is not user provided", func() {
		meshRef := &core_types.ResourceRef{
			Name:      "name",
			Namespace: "namespace",
		}
		vm := &v1alpha1.VirtualMesh{
			ObjectMeta: metav1.ObjectMeta{
				Name:      meshRef.Name,
				Namespace: meshRef.Namespace,
			},
			Spec: v1alpha1_types.VirtualMeshSpec{
				CertificateAuthority: &v1alpha1_types.VirtualMeshSpec_CertificateAuthority{
					Type: &v1alpha1_types.VirtualMeshSpec_CertificateAuthority_Builtin_{
						Builtin: &v1alpha1_types.VirtualMeshSpec_CertificateAuthority_Builtin{},
					},
				},
			},
		}
		virtualMeshClient.EXPECT().Get(ctx, meshRef.Name, meshRef.Namespace).Return(vm, nil)
		secretClient.
			EXPECT().
			Get(ctx, cert_signer.DefaultRootCaName(vm), env.GetWriteNamespace()).
			Return(nil, errors.NewNotFound(corev1.Resource("secret"), "non-extant-secret"))
		expectedRootCaData := &cert_secrets.RootCAData{}
		rootCertGenerator.
			EXPECT().
			GenRootCertAndKey(vm.Spec.GetCertificateAuthority().GetBuiltin()).
			Return(expectedRootCaData, nil)
		secretClient.
			EXPECT().
			Create(ctx, expectedRootCaData.BuildSecret(cert_signer.DefaultRootCaName(vm), env.GetWriteNamespace())).
			Return(nil)
		rootCaData, err := virtualMeshCertClient.GetRootCaBundle(ctx, meshRef)
		Expect(err).ToNot(HaveOccurred())
		Expect(rootCaData).To(Equal(expectedRootCaData))
	})

	It("will get auto-generated root cert if CertificateAuthority is not user provided and already exists", func() {
		meshRef := &core_types.ResourceRef{
			Name:      "name",
			Namespace: "namespace",
		}
		vm := &v1alpha1.VirtualMesh{
			ObjectMeta: metav1.ObjectMeta{
				Name:      meshRef.Name,
				Namespace: meshRef.Namespace,
			},
			Spec: v1alpha1_types.VirtualMeshSpec{
				CertificateAuthority: &v1alpha1_types.VirtualMeshSpec_CertificateAuthority{
					Type: &v1alpha1_types.VirtualMeshSpec_CertificateAuthority_Builtin_{
						Builtin: &v1alpha1_types.VirtualMeshSpec_CertificateAuthority_Builtin{},
					},
				},
			},
		}
		virtualMeshClient.EXPECT().Get(ctx, meshRef.Name, meshRef.Namespace).Return(vm, nil)
		expectedRootCaData := &cert_secrets.RootCAData{}
		secretClient.
			EXPECT().
			Get(ctx, cert_signer.DefaultRootCaName(vm), env.GetWriteNamespace()).
			Return(expectedRootCaData.BuildSecret(cert_signer.DefaultRootCaName(vm), env.GetWriteNamespace()), nil)
		rootCaData, err := virtualMeshCertClient.GetRootCaBundle(ctx, meshRef)
		Expect(err).ToNot(HaveOccurred())
		Expect(rootCaData).To(Equal(expectedRootCaData))
	})

	It("will default to using builtin cert", func() {
		meshRef := &core_types.ResourceRef{
			Name:      "name",
			Namespace: "namespace",
		}
		vm := &v1alpha1.VirtualMesh{
			ObjectMeta: metav1.ObjectMeta{
				Name:      meshRef.Name,
				Namespace: meshRef.Namespace,
			},
			Spec: v1alpha1_types.VirtualMeshSpec{},
		}
		virtualMeshClient.EXPECT().Get(ctx, meshRef.Name, meshRef.Namespace).Return(vm, nil)
		expectedRootCaData := &cert_secrets.RootCAData{}
		secretClient.
			EXPECT().
			Get(ctx, cert_signer.DefaultRootCaName(vm), env.GetWriteNamespace()).
			Return(expectedRootCaData.BuildSecret(cert_signer.DefaultRootCaName(vm), env.GetWriteNamespace()), nil)
		rootCaData, err := virtualMeshCertClient.GetRootCaBundle(ctx, meshRef)
		Expect(err).ToNot(HaveOccurred())
		Expect(rootCaData).To(Equal(expectedRootCaData))
	})
})
