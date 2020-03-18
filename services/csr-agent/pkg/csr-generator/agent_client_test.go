package csr_generator_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	. "github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1"
	mock_kubernetes_core "github.com/solo-io/mesh-projects/pkg/clients/kubernetes/core/mocks"
	mock_certgen "github.com/solo-io/mesh-projects/pkg/security/certgen/mocks"
	cert_secrets "github.com/solo-io/mesh-projects/pkg/security/secrets"
	csr_generator "github.com/solo-io/mesh-projects/services/csr-agent/pkg/csr-generator"
	mock_csr_agent_controller "github.com/solo-io/mesh-projects/services/csr-agent/pkg/csr-generator/mocks"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ = Describe("agent client", func() {
	var (
		ctrl                    *gomock.Controller
		ctx                     context.Context
		secretClient            *mock_kubernetes_core.MockSecretsClient
		signer                  *mock_certgen.MockSigner
		certClient              csr_generator.CertClient
		mockPrivateKeyGenerator *mock_csr_agent_controller.MockPrivateKeyGenerator
		testErr                 = eris.New("hello")
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		secretClient = mock_kubernetes_core.NewMockSecretsClient(ctrl)
		signer = mock_certgen.NewMockSigner(ctrl)
		mockPrivateKeyGenerator = mock_csr_agent_controller.NewMockPrivateKeyGenerator(ctrl)
		certClient = csr_generator.NewCertClient(secretClient, signer, mockPrivateKeyGenerator)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("will return the errror if secret client does not return is not found", func() {
		csr := &v1alpha1.VirtualMeshCertificateSigningRequest{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "name",
				Namespace: "namespace",
			},
		}
		secretClient.EXPECT().
			Get(ctx, csr.GetName()+csr_generator.PrivateKeyNameSuffix, csr.GetNamespace()).
			Return(nil, testErr)

		_, err := certClient.EnsureSecretKey(ctx, csr)
		Expect(err).To(HaveInErrorChain(testErr))
	})

	It("will attempt to marshal into cert secret if secret is found", func() {
		csr := &v1alpha1.VirtualMeshCertificateSigningRequest{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "name",
				Namespace: "namespace",
			},
		}
		secret := &v1.Secret{}
		secretClient.EXPECT().
			Get(ctx, csr.GetName()+csr_generator.PrivateKeyNameSuffix, csr.GetNamespace()).
			Return(secret, nil)

		_, err := certClient.EnsureSecretKey(ctx, csr)
		Expect(err).To(HaveInErrorChain(cert_secrets.NoCaKeyFoundError(secret.ObjectMeta)))
	})

	It("will attempt to marshal into cert secret, will return data if successful", func() {
		csr := &v1alpha1.VirtualMeshCertificateSigningRequest{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "name",
				Namespace: "namespace",
			},
		}
		secret := &v1.Secret{
			Data: map[string][]byte{
				cert_secrets.CaCertID:         []byte("cacert"),
				cert_secrets.CaPrivateKeyID:   []byte("cakey"),
				cert_secrets.CertChainID:      []byte("certchain"),
				cert_secrets.RootPrivateKeyID: []byte("key"),
				cert_secrets.RootCertID:       []byte("cert"),
			},
		}
		secretClient.EXPECT().
			Get(ctx, csr.GetName()+csr_generator.PrivateKeyNameSuffix, csr.GetNamespace()).
			Return(secret, nil)

		caData, err := certClient.EnsureSecretKey(ctx, csr)
		Expect(err).NotTo(HaveOccurred())
		matchData, err := cert_secrets.IntermediateCADataFromSecret(secret)
		Expect(err).NotTo(HaveOccurred())
		Expect(caData).To(Equal(matchData))
	})

	It("will attempt to create new secret if old one cannot be found", func() {
		csr := &v1alpha1.VirtualMeshCertificateSigningRequest{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "name",
				Namespace: "namespace",
			},
		}

		secretClient.EXPECT().
			Get(ctx, csr.GetName()+csr_generator.PrivateKeyNameSuffix, csr.GetNamespace()).
			Return(nil, errors.NewNotFound(schema.GroupResource{}, "name"))

		key := []byte{'a', 'b', 'c'}
		mockPrivateKeyGenerator.
			EXPECT().
			GenerateRSA(csr_generator.PrivateKeySizeBytes).
			Return(key, nil)

		intermediateCAData := &cert_secrets.IntermediateCAData{
			CaPrivateKey: key,
		}
		expectedSecret := intermediateCAData.BuildSecret(csr.GetName()+csr_generator.PrivateKeyNameSuffix, csr.GetNamespace())
		secretClient.
			EXPECT().
			Create(ctx, expectedSecret).
			Return(nil)
		_, err := certClient.EnsureSecretKey(ctx, csr)
		Expect(err).ToNot(HaveOccurred())
	})

})
