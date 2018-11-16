package istio_test

import (
	"context"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/test/helpers"
	"github.com/solo-io/solo-kit/test/tests/typed"
	"github.com/solo-io/supergloo/pkg/api/external/istio/encryption/v1"
	supergloov1 "github.com/solo-io/supergloo/pkg/api/v1"
	. "github.com/solo-io/supergloo/pkg/translator/istio"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"

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
	It("works", func() {
		istioSecret := v1.IstioCacertsSecret{
			Metadata: core.Metadata{
				Name:      "mysecret",
				Namespace: namespace,
			},
			RootCert:  "foo",
			CertChain: "",
			CaCert:    "foo",
			CaKey:     "bar",
		}

		settingsClientFactory := &factory.KubeSecretClientFactory{
			Clientset: kube,
		}
		secretClient, err := v1.NewIstioCacertsSecretClient(settingsClientFactory)
		Expect(err).NotTo(HaveOccurred())

		ref := istioSecret.Metadata.Ref()
		err = (&EncryptionSyncer{
			IstioNamespace: namespace,
			Kube:           kube,
			SecretClient:   secretClient,
		}).Sync(context.TODO(), &supergloov1.TranslatorSnapshot{
			Meshes: map[string]supergloov1.MeshList{
				"ignored-at-this-point": {{
					TargetMesh: &supergloov1.TargetMesh{
						MeshType: supergloov1.MeshType_ISTIO,
					},
					Encryption: &supergloov1.Encryption{
						TlsEnabled: true,
						Secret:     &ref,
					},
				}},
			},
			Istiocerts: v1.IstiocertsByNamespace{"whatever": {&istioSecret}},
		})
		Expect(err).NotTo(HaveOccurred())

		secret, err := kube.CoreV1().Secrets(namespace).Get("cacerts", v12.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(secret.Data).To(Equal(map[string][]uint8{
			"ca-key.pem": {98, 97, 114, 10},
			"ca-cert.pem": {102, 111, 111, 10},
			"root-cert.pem": {102, 111, 111, 10},
		}))

		// TODO: actually load from Kube and inspect
	})
})
