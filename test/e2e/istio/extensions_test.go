package istio_test

import (
	"fmt"
	"github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/service-mesh-hub/pkg/api/settings.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/common/defaults"
	"github.com/solo-io/service-mesh-hub/test/extensions"
	"github.com/solo-io/service-mesh-hub/test/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = FDescribe("Istio Networking Extensions", func() {
	var (
		err          error
		manifest     utils.Manifest
		smhNamespace = defaults.GetPodNamespace()
	)

	AfterEach(func() {
		manifest.Cleanup(smhNamespace)
	})

	It("enables communication across clusters using global dns names", func() {
		manifest, err = utils.NewManifest("extension-settings.yaml")
		Expect(err).NotTo(HaveOccurred())

		By("with federation enabled, TrafficShifts can be used for subsets across meshes ", func() {
			// run extensions server
			go func() {
				defer GinkgoRecover()
				err := extensions.RunExtensionsServer()
				Expect(err).NotTo(HaveOccurred())
			}()
			// run echo server
			go func() {
				defer GinkgoRecover()
				err := extensions.RunEchoSerer()
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
						Address:                    fmt.Sprintf("%v:%v", extensions.DockerHostAddress, extensions.ExtensionsServerPort),
						Insecure:                   true,
						ReconnectOnNetworkFailures: true,
					}},
				},
			})
			Expect(err).NotTo(HaveOccurred())
			err = manifest.KubeApply(smhNamespace)
			Expect(err).NotTo(HaveOccurred())

			// restart the networking pod to pick up the settings
			err = testutils.Kubectl("delete", "pod", "-l", "app=networking", "-n", smhNamespace)
			Expect(err).NotTo(HaveOccurred())

			// check we can eventually hit the echo server via the gateway
			Eventually(curlGateway(extensions.EchoServerHostname, "/", "echo-this-back-to-me"), "30s", "1s").Should(ContainSubstring("echo-this-back-to-me"))
		})
	})
})
