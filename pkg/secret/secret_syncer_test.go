package secret_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/memory"
	"github.com/solo-io/supergloo/mock/pkg/kube"
	istiov1 "github.com/solo-io/supergloo/pkg/api/external/istio/encryption/v1"
	"github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/secret"
	"github.com/solo-io/supergloo/test/util"
)

var T *testing.T

func TestSecret(t *testing.T) {
	RegisterFailHandler(Fail)
	T = t
	RunSpecs(t, "Shared Suite")
}

var _ = Describe("SecretSyncer", func() {

	var mockPodClient *mock_kube.MockPodClient
	var mockSecretClient *mock_kube.MockSecretClient
	var syncer *secret.KubeSecretSyncer

	BeforeEach(func() {
		ctrl := gomock.NewController(T)
		defer ctrl.Finish()

		inMemoryCache := memory.NewInMemoryResourceCache()
		client, err := istiov1.NewIstioCacertsSecretClient(&factory.MemoryResourceClientFactory{
			Cache: inMemoryCache,
		})
		Expect(err).To(BeNil())

		mockPodClient = mock_kube.NewMockPodClient(ctrl)
		mockSecretClient = mock_kube.NewMockSecretClient(ctrl)
		syncer = &secret.KubeSecretSyncer{
			SecretClient:      mockSecretClient,
			PodClient:         mockPodClient,
			IstioSecretClient: client,
		}
	})

	It("handles nil encryption", func() {
		Expect(syncer.SyncSecret(context.TODO(), util.SecretNamespace, nil, util.GetTestSecrets(), false)).To(BeNil())
	})

	It("handles mtls disabled", func() {
		encryption := &v1.Encryption{
			TlsEnabled: false,
		}
		Expect(syncer.SyncSecret(context.TODO(), util.SecretNamespace, encryption, util.GetTestSecrets(), false)).To(BeNil())
	})

	It("handles mtls enabled nil secret", func() {
		encryption := &v1.Encryption{
			TlsEnabled: true,
		}
		Expect(syncer.SyncSecret(context.TODO(), util.SecretNamespace, encryption, util.GetTestSecrets(), false)).To(BeNil())
	})

	It("errors mtls enabled with missing secret", func() {
		encryption := &v1.Encryption{
			TlsEnabled: true,
			Secret:     util.GetRef(util.SecretNamespace, util.SecretNameMissing),
		}
		err := syncer.SyncSecret(context.TODO(), util.SecretNamespace, encryption, util.GetTestSecrets(), false)
		Expect(err).NotTo(BeNil())
		Expect(err.Error()).Should(ContainSubstring("Error finding secret referenced in mesh config"))
	})

	It("errors mtls enabled with invalid secret: missing root", func() {
		encryption := &v1.Encryption{
			TlsEnabled: true,
			Secret:     util.GetRef(util.SecretNamespace, util.SecretNameMissingRoot),
		}
		err := syncer.SyncSecret(context.TODO(), util.SecretNamespace, encryption, util.GetTestSecrets(), false)
		Expect(err).NotTo(BeNil())
		Expect(err.Error()).Should(ContainSubstring("Root cert is missing."))
	})

	It("errors mtls enabled with invalid secret: missing key", func() {
		encryption := &v1.Encryption{
			TlsEnabled: true,
			Secret:     util.GetRef(util.SecretNamespace, util.SecretNameMissingKey),
		}
		err := syncer.SyncSecret(context.TODO(), util.SecretNamespace, encryption, util.GetTestSecrets(), false)
		Expect(err).NotTo(BeNil())
		Expect(err.Error()).Should(ContainSubstring("Private key is missing."))
	})

	It("creates secret when no existing and restarts citadel and deletes default", func() {
		encryption := &v1.Encryption{
			TlsEnabled: true,
			Secret:     util.GetRef(util.SecretNamespace, util.SecretNameValid),
		}

		deleteCall := mockSecretClient.EXPECT().Delete(util.SecretNamespace, secret.DefaultRootCertificateSecretName).Return(nil).Times(1)
		selector := make(map[string]string)
		selector["istio"] = "citadel"
		mockPodClient.EXPECT().RestartPods(util.SecretNamespace, selector).Return(nil).Times(1).After(deleteCall)
		Expect(syncer.SyncSecret(context.TODO(), util.SecretNamespace, encryption, util.GetTestSecrets(), false)).To(BeNil())
	})

	It("creates secret when no existing and doesn't restart citadel or delete default during install", func() {
		encryption := &v1.Encryption{
			TlsEnabled: true,
			Secret:     util.GetRef(util.SecretNamespace, util.SecretNameValid),
		}

		mockSecretClient.EXPECT().Delete(util.SecretNamespace, secret.DefaultRootCertificateSecretName).Return(nil).Times(0)
		selector := make(map[string]string)
		selector["istio"] = "citadel"
		mockPodClient.EXPECT().RestartPods(util.SecretNamespace, selector).Return(nil).Times(0)
		Expect(syncer.SyncSecret(context.TODO(), util.SecretNamespace, encryption, util.GetTestSecrets(), true)).To(BeNil())
	})
})
