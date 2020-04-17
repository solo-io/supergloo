package cert_signer_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	. "github.com/solo-io/go-utils/testutils"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	zephyr_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	mock_certgen "github.com/solo-io/service-mesh-hub/pkg/security/certgen/mocks"
	cert_secrets "github.com/solo-io/service-mesh-hub/pkg/security/secrets"
	cert_signer "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/security/cert-signer"
	mock_kubernetes_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/kubernetes/core/v1"
	mock_zephyr_networking "github.com/solo-io/service-mesh-hub/test/mocks/clients/networking.zephyr.solo.io/v1alpha1"
	k8s_core_types "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("virtual mesh cert client", func() {
	var (
		ctrl                  *gomock.Controller
		ctx                   context.Context
		secretClient          *mock_kubernetes_core.MockSecretClient
		virtualMeshClient     *mock_zephyr_networking.MockVirtualMeshClient
		virtualMeshCertClient cert_signer.VirtualMeshCertClient
		rootCertGenerator     *mock_certgen.MockRootCertGenerator
		testErr               = eris.New("hello")
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		secretClient = mock_kubernetes_core.NewMockSecretClient(ctrl)
		virtualMeshClient = mock_zephyr_networking.NewMockVirtualMeshClient(ctrl)
		rootCertGenerator = mock_certgen.NewMockRootCertGenerator(ctrl)
		virtualMeshCertClient = cert_signer.NewVirtualMeshCertClient(secretClient, virtualMeshClient, rootCertGenerator)
	})

	It("will fail if virtualMesh cannot be found", func() {
		meshRef := &zephyr_core_types.ResourceRef{
			Name:      "name",
			Namespace: "namespace",
		}
		virtualMeshClient.EXPECT().GetVirtualMesh(ctx, client.ObjectKey{Name: meshRef.Name, Namespace: meshRef.Namespace}).Return(nil, testErr)
		_, err := virtualMeshCertClient.GetRootCaBundle(ctx, meshRef)
		Expect(err).To(HaveOccurred())
		Expect(err).To(HaveInErrorChain(testErr))
	})

	It("will use user provided trust bundle in vm if set", func() {
		meshRef := &zephyr_core_types.ResourceRef{
			Name:      "name",
			Namespace: "namespace",
		}
		vm := &zephyr_networking.VirtualMesh{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      meshRef.Name,
				Namespace: meshRef.Namespace,
			},
			Spec: zephyr_networking_types.VirtualMeshSpec{
				CertificateAuthority: &zephyr_networking_types.VirtualMeshSpec_CertificateAuthority{
					Type: &zephyr_networking_types.VirtualMeshSpec_CertificateAuthority_Provided_{
						Provided: &zephyr_networking_types.VirtualMeshSpec_CertificateAuthority_Provided{
							Certificate: &zephyr_core_types.ResourceRef{
								Name:      "tb_name",
								Namespace: "tb_namespace",
							},
						},
					},
				},
			},
		}
		virtualMeshClient.EXPECT().GetVirtualMesh(ctx, client.ObjectKey{Name: meshRef.Name, Namespace: meshRef.Namespace}).Return(vm, nil)
		secretClient.
			EXPECT().
			GetSecret(ctx,
				client.ObjectKey{
					Name:      vm.Spec.GetCertificateAuthority().GetProvided().GetCertificate().GetName(),
					Namespace: vm.Spec.GetCertificateAuthority().GetProvided().GetCertificate().GetNamespace()}).
			Return(nil, testErr)
		_, err := virtualMeshCertClient.GetRootCaBundle(ctx, meshRef)
		Expect(err).To(HaveOccurred())
		Expect(err).To(HaveInErrorChain(testErr))
	})

	It("will return proper CA data", func() {
		meshRef := &zephyr_core_types.ResourceRef{
			Name:      "name",
			Namespace: "namespace",
		}
		vm := &zephyr_networking.VirtualMesh{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      meshRef.Name,
				Namespace: meshRef.Namespace,
			},
			Spec: zephyr_networking_types.VirtualMeshSpec{
				CertificateAuthority: &zephyr_networking_types.VirtualMeshSpec_CertificateAuthority{
					Type: &zephyr_networking_types.VirtualMeshSpec_CertificateAuthority_Provided_{
						Provided: &zephyr_networking_types.VirtualMeshSpec_CertificateAuthority_Provided{
							Certificate: &zephyr_core_types.ResourceRef{
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
		matchSecret := &k8s_core_types.Secret{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      vm.Spec.GetCertificateAuthority().GetProvided().GetCertificate().GetName(),
				Namespace: vm.Spec.GetCertificateAuthority().GetProvided().GetCertificate().GetNamespace(),
			},
			Data: map[string][]byte{
				cert_secrets.RootCertID:       matchData.RootCert,
				cert_secrets.RootPrivateKeyID: matchData.PrivateKey,
			},
			Type: cert_secrets.RootCertSecretType,
		}
		virtualMeshClient.EXPECT().GetVirtualMesh(ctx, client.ObjectKey{Name: meshRef.Name, Namespace: meshRef.Namespace}).Return(vm, nil)
		secretClient.
			EXPECT().
			GetSecret(ctx,
				client.ObjectKey{
					Name:      vm.Spec.GetCertificateAuthority().GetProvided().GetCertificate().GetName(),
					Namespace: vm.Spec.GetCertificateAuthority().GetProvided().GetCertificate().GetNamespace()},
			).
			Return(matchSecret, nil)
		data, err := virtualMeshCertClient.GetRootCaBundle(ctx, meshRef)
		Expect(err).NotTo(HaveOccurred())
		Expect(data).To(Equal(matchData))
	})

	It("will create auto-generated root cert if CertificateAuthority is not user provided", func() {
		meshRef := &zephyr_core_types.ResourceRef{
			Name:      "name",
			Namespace: "namespace",
		}
		vm := &zephyr_networking.VirtualMesh{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      meshRef.Name,
				Namespace: meshRef.Namespace,
			},
			Spec: zephyr_networking_types.VirtualMeshSpec{
				CertificateAuthority: &zephyr_networking_types.VirtualMeshSpec_CertificateAuthority{
					Type: &zephyr_networking_types.VirtualMeshSpec_CertificateAuthority_Builtin_{
						Builtin: &zephyr_networking_types.VirtualMeshSpec_CertificateAuthority_Builtin{},
					},
				},
			},
		}
		virtualMeshClient.EXPECT().GetVirtualMesh(ctx, client.ObjectKey{Name: meshRef.Name, Namespace: meshRef.Namespace}).Return(vm, nil)
		secretClient.
			EXPECT().
			GetSecret(ctx, client.ObjectKey{Name: cert_signer.DefaultRootCaName(vm), Namespace: env.GetWriteNamespace()}).
			Return(nil, errors.NewNotFound(k8s_core_types.Resource("secret"), "non-extant-secret"))
		expectedRootCaData := &cert_secrets.RootCAData{}
		rootCertGenerator.
			EXPECT().
			GenRootCertAndKey(vm.Spec.GetCertificateAuthority().GetBuiltin()).
			Return(expectedRootCaData, nil)
		secretClient.
			EXPECT().
			CreateSecret(ctx, expectedRootCaData.BuildSecret(cert_signer.DefaultRootCaName(vm), env.GetWriteNamespace())).
			Return(nil)
		rootCaData, err := virtualMeshCertClient.GetRootCaBundle(ctx, meshRef)
		Expect(err).ToNot(HaveOccurred())
		Expect(rootCaData).To(Equal(expectedRootCaData))
	})

	It("will get auto-generated root cert if CertificateAuthority is not user provided and already exists", func() {
		meshRef := &zephyr_core_types.ResourceRef{
			Name:      "name",
			Namespace: "namespace",
		}
		vm := &zephyr_networking.VirtualMesh{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      meshRef.Name,
				Namespace: meshRef.Namespace,
			},
			Spec: zephyr_networking_types.VirtualMeshSpec{
				CertificateAuthority: &zephyr_networking_types.VirtualMeshSpec_CertificateAuthority{
					Type: &zephyr_networking_types.VirtualMeshSpec_CertificateAuthority_Builtin_{
						Builtin: &zephyr_networking_types.VirtualMeshSpec_CertificateAuthority_Builtin{},
					},
				},
			},
		}
		virtualMeshClient.EXPECT().GetVirtualMesh(ctx, client.ObjectKey{Name: meshRef.Name, Namespace: meshRef.Namespace}).Return(vm, nil)
		expectedRootCaData := &cert_secrets.RootCAData{}
		secretClient.
			EXPECT().
			GetSecret(ctx, client.ObjectKey{Name: cert_signer.DefaultRootCaName(vm), Namespace: env.GetWriteNamespace()}).
			Return(expectedRootCaData.BuildSecret(cert_signer.DefaultRootCaName(vm), env.GetWriteNamespace()), nil)
		rootCaData, err := virtualMeshCertClient.GetRootCaBundle(ctx, meshRef)
		Expect(err).ToNot(HaveOccurred())
		Expect(rootCaData).To(Equal(expectedRootCaData))
	})

	It("will default to using builtin cert", func() {
		meshRef := &zephyr_core_types.ResourceRef{
			Name:      "name",
			Namespace: "namespace",
		}
		vm := &zephyr_networking.VirtualMesh{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      meshRef.Name,
				Namespace: meshRef.Namespace,
			},
			Spec: zephyr_networking_types.VirtualMeshSpec{},
		}
		virtualMeshClient.EXPECT().GetVirtualMesh(ctx, client.ObjectKey{Name: meshRef.Name, Namespace: meshRef.Namespace}).Return(vm, nil)
		expectedRootCaData := &cert_secrets.RootCAData{}
		secretClient.
			EXPECT().
			GetSecret(ctx, client.ObjectKey{Name: cert_signer.DefaultRootCaName(vm), Namespace: env.GetWriteNamespace()}).
			Return(expectedRootCaData.BuildSecret(cert_signer.DefaultRootCaName(vm), env.GetWriteNamespace()), nil)
		rootCaData, err := virtualMeshCertClient.GetRootCaBundle(ctx, meshRef)
		Expect(err).ToNot(HaveOccurred())
		Expect(rootCaData).To(Equal(expectedRootCaData))
	})
})
