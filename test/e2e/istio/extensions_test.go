package istio_test

import (
	"fmt"
	"time"

	"github.com/solo-io/gloo-mesh/pkg/api/settings.gloomesh.solo.io/v1alpha2"
	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	"github.com/solo-io/gloo-mesh/test/extensions"
	"github.com/solo-io/gloo-mesh/test/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Istio Networking Extensions", func() {
	var (
		err          error
		manifest     utils.Manifest
		smhNamespace = defaults.GetPodNamespace()
	)

	AfterEach(func() {
		manifest, err = utils.NewManifest("default-settings.yaml")
		Expect(err).NotTo(HaveOccurred())
		// update settings to remove our extensions server
		err = manifest.AppendResources(&v1alpha2.Settings{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Settings",
				APIVersion: v1alpha2.SchemeGroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace: smhNamespace,
				Name:      "settings", // the default/expected name
			},
		})
		Expect(err).NotTo(HaveOccurred())
		err = manifest.KubeApply(smhNamespace)
		Expect(err).NotTo(HaveOccurred())
	})

	It("enables communication across clusters using global dns names", func() {
		manifest, err = utils.NewManifest("extension-settings.yaml")
		Expect(err).NotTo(HaveOccurred())

		By("with extensions enabled, additional configs can be added to SMH outputs", func() {

			helloMsg := "hello from a 3rd party"

			srv := extensions.NewTestExtensionsServer()

			// run extensions server
			go func() {
				defer GinkgoRecover()
				err := srv.Run()
				Expect(err).NotTo(HaveOccurred())
			}()
			// run hello server
			go func() {
				defer GinkgoRecover()
				err := extensions.RunHelloServer(helloMsg)
				Expect(err).NotTo(HaveOccurred())
			}()

			// update settings to connect our extensions server
			err = manifest.AppendResources(&v1alpha2.Settings{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Settings",
					APIVersion: v1alpha2.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: smhNamespace,
					Name:      "settings", // the default/expected name
				},
				Spec: v1alpha2.SettingsSpec{
					NetworkingExtensionServers: []*v1alpha2.NetworkingExtensionsServer{{
						// use the machine's docker host address
						Address:                    fmt.Sprintf("%v:%v", extensions.DockerHostAddress, extensions.ExtensionsServerPort),
						Insecure:                   true,
						ReconnectOnNetworkFailures: true,
					}},
				},
			})
			Expect(err).NotTo(HaveOccurred())
			err = manifest.KubeApply(smhNamespace)
			Expect(err).NotTo(HaveOccurred())

			// ensure the server eventually connects to us
			Eventually(srv.HasConnected, time.Minute*2).Should(BeTrue())

			// check we can eventually hit the echo server via the gateway.
			// This request verifies that Envoy has config provided by Service Entries from our test extensions server.
			Eventually(curlHelloServer, "30s", "1s").Should(ContainSubstring(helloMsg))
		})
	})
})
