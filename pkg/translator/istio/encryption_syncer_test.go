package istio_test

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/test/helpers"
	"github.com/solo-io/solo-kit/test/tests/typed"
	. "github.com/solo-io/supergloo/pkg/api/external/istio/encryption/v1"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"k8s.io/client-go/kubernetes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
)

var _ = Describe("EncryptionSyncer", func() {
	tester := &typed.KubeConfigMapRcTester{}
	var (
		namespace string
		kube      kubernetes.Interface
	)
	BeforeEach(func() {
		namespace = helpers.RandString(6)
		fact := tester.Setup(namespace)
		kube = fact.(*factory.KubeConfigMapClientFactory).Clientset
	})
	AfterEach(func() {
		tester.Teardown(namespace)
	})
	FIt("works", func() {
		secretClient, err := NewIstioCacertsSecretClient(&factory.KubeSecretClientFactory{
			Clientset: kube,
		})
		Expect(err).NotTo(HaveOccurred())

		istioSecret := IstioCacertsSecret{
			Metadata: core.Metadata{
				Name:      "mysecret",
				Namespace: namespace,
			},
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
