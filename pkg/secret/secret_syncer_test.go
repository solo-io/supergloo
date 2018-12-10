package secret_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/memory"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/mock/pkg/kube"
	istiov1 "github.com/solo-io/supergloo/pkg/api/external/istio/encryption/v1"
	"github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/secret"
)

var T *testing.T

func TestSecret(t *testing.T) {
	RegisterFailHandler(Fail)
	T = t
	RunSpecs(t, "Shared Suite")
}

var _ = Describe("SecretSyncer", func() {
	secretNamespace := "foo"

	getIstioSecret := func(namespace string, name string) *istiov1.IstioCacertsSecret {
		return &istiov1.IstioCacertsSecret{
			Metadata: core.Metadata{
				Namespace: namespace,
				Name:      name,
			},
		}
	}

	getTestSecrets := func() istiov1.IstioCacertsSecretList {
		var list istiov1.IstioCacertsSecretList
		missingRoot := getIstioSecret(secretNamespace, "missing_root")
		list = append(list, missingRoot)
		missingKey := getIstioSecret(secretNamespace, "missing_key")
		missingKey.RootCert = "root"
		list = append(list, missingKey)
		valid := getIstioSecret(secretNamespace, "valid")
		valid.RootCert = "root"
		valid.CaKey = "key"
		list = append(list, valid)
		return list
	}

	getRef := func(namespace string, name string) *core.ResourceRef {
		return &core.ResourceRef{
			Namespace: namespace,
			Name:      name,
		}
	}

	var mockPodClient *mock_kube.MockPodClient
	var mockSecretClient *mock_kube.MockSecretClient
	var syncer secret.SecretSyncer

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
		syncer = secret.SecretSyncer{
			SecretClient:      mockSecretClient,
			PodClient:         mockPodClient,
			Preinstall:        false,
			IstioSecretClient: client,
			IstioSecretList:   getTestSecrets(),
		}
	})

	It("handles nil encryption", func() {
		Expect(syncer.SyncSecret(context.TODO(), secretNamespace, nil)).To(BeNil())
	})

	It("handles mtls disabled", func() {
		encryption := &v1.Encryption{
			TlsEnabled: false,
		}
		Expect(syncer.SyncSecret(context.TODO(), secretNamespace, encryption)).To(BeNil())
	})

	It("handles mtls enabled nil secret", func() {
		encryption := &v1.Encryption{
			TlsEnabled: true,
		}
		Expect(syncer.SyncSecret(context.TODO(), secretNamespace, encryption)).To(BeNil())
	})

	It("errors mtls enabled with missing secret", func() {
		encryption := &v1.Encryption{
			TlsEnabled: true,
			Secret:     getRef(secretNamespace, "missing"),
		}
		err := syncer.SyncSecret(context.TODO(), secretNamespace, encryption)
		Expect(err).NotTo(BeNil())
		Expect(err.Error()).Should(ContainSubstring("Error finding secret referenced in mesh config"))
	})

	It("errors mtls enabled with invalid secret: missing root", func() {
		encryption := &v1.Encryption{
			TlsEnabled: true,
			Secret:     getRef(secretNamespace, "missing_root"),
		}
		err := syncer.SyncSecret(context.TODO(), secretNamespace, encryption)
		Expect(err).NotTo(BeNil())
		Expect(err.Error()).Should(ContainSubstring("Root cert is missing."))
	})

	It("errors mtls enabled with invalid secret: missing key", func() {
		encryption := &v1.Encryption{
			TlsEnabled: true,
			Secret:     getRef(secretNamespace, "missing_key"),
		}
		err := syncer.SyncSecret(context.TODO(), secretNamespace, encryption)
		Expect(err).NotTo(BeNil())
		Expect(err.Error()).Should(ContainSubstring("Private key is missing."))
	})

	It("creates secret when no existing and restarts citadel and deletes default", func() {
		encryption := &v1.Encryption{
			TlsEnabled: true,
			Secret:     getRef(secretNamespace, "valid"),
		}

		deleteCall := mockSecretClient.EXPECT().Delete(secretNamespace, "istio.default").Return(nil).Times(1)
		selector := make(map[string]string)
		selector["istio"] = "citadel"
		mockPodClient.EXPECT().RestartPods(secretNamespace, selector).Return(nil).Times(1).After(deleteCall)
		Expect(syncer.SyncSecret(context.TODO(), secretNamespace, encryption)).To(BeNil())
	})

	It("creates secret when no existing and doesn't restart citadel or delete default during install", func() {
		encryption := &v1.Encryption{
			TlsEnabled: true,
			Secret:     getRef(secretNamespace, "valid"),
		}
		syncer.Preinstall = true

		mockSecretClient.EXPECT().Delete(secretNamespace, "istio.default").Return(nil).Times(0)
		selector := make(map[string]string)
		selector["istio"] = "citadel"
		mockPodClient.EXPECT().RestartPods(secretNamespace, selector).Return(nil).Times(0)
		Expect(syncer.SyncSecret(context.TODO(), secretNamespace, encryption)).To(BeNil())
	})
})
