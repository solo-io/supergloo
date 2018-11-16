package istio_test

import (
	"path/filepath"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"k8s.io/client-go/kubernetes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"os"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	gloov1 "github.com/solo-io/supergloo/pkg/api/external/gloo/v1"
	"k8s.io/client-go/tools/clientcmd"
)

var _ = Describe("EncryptionSyncer", func() {
	It("works", func() {
		kubeconfigPath := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		Expect(err).NotTo(HaveOccurred())
		kubeClient, err := kubernetes.NewForConfig(cfg)
		Expect(err).NotTo(HaveOccurred())
		secretClient, err := gloov1.NewSecretClient(&factory.KubeSecretClientFactory{
			Clientset: kubeClient,
		})
		Expect(err).NotTo(HaveOccurred())

		istioSecret := gloov1.IstioCacertsSecret{
			RootCert:  "foo",
			CertChain: "",
			CaCert:    "foo",
			CaKey:     "bar",
		}
		istioSecretWrapper := gloov1.Secret_Cacerts{
			Cacerts: &istioSecret,
		}
		secret := gloov1.Secret{
			Kind: &istioSecretWrapper,
			Metadata: core.Metadata{
				Namespace: "istio-system",
				Name:      "test-tls-secret",
			},
		}

		_, err = secretClient.Write(&secret, clients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())

		// TODO: actually load from Kube and inspect
	})
})
