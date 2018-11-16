package istio_test

import (
	"path/filepath"

	. "github.com/solo-io/supergloo/pkg/api/external/istio/encryption/v1"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"k8s.io/client-go/kubernetes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"os"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"k8s.io/client-go/tools/clientcmd"
)

var _ = Describe("EncryptionSyncer", func() {
	FIt("works", func() {
		kubeconfigPath := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		Expect(err).NotTo(HaveOccurred())
		kubeClient, err := kubernetes.NewForConfig(cfg)
		Expect(err).NotTo(HaveOccurred())
		secretClient, err := NewIstioCacertsSecretClient(&factory.KubeSecretClientFactory{
			Clientset: kubeClient,
		})
		Expect(err).NotTo(HaveOccurred())

		istioSecret := IstioCacertsSecret{
			RootCert:  "foo",
			CertChain: "",
			CaCert:    "foo",
			CaKey:     "bar",
		}
		secretClient.Delete(istioSecret.Metadata.Namespace, istioSecret.Metadata.Name, clients.DeleteOpts{})
		_, err = secretClient.Write(&istioSecret, clients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())

		// TODO: actually load from Kube and inspect
	})
})
