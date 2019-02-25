package e2e_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/cli/pkg/helpers"
	"github.com/solo-io/supergloo/cli/test/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("E2e", func() {
	It("installs and uninstalls istio", func() {
		err := utils.Supergloo("install istio --name=my-istio")
		Expect(err).NotTo(HaveOccurred())

		installClient := helpers.MustInstallClient()

		Eventually(func() (core.Status_State, error) {
			i, err := installClient.Read("supergloo-system", "my-istio", clients.ReadOpts{})
			if err != nil {
				return 0, err
			}
			Expect(i.Status.State).NotTo(Equal(core.Status_Rejected))
			return i.Status.State, nil
		}, time.Minute*2).Should(Equal(core.Status_Accepted))

		Eventually(func() error {
			_, err := kube.CoreV1().Services("istio-system").Get("istio-pilot", metav1.GetOptions{})
			return err
		}).ShouldNot(HaveOccurred())

		meshClient := helpers.MustMeshClient()
		Eventually(func() error {
			_, err := meshClient.Read("supergloo-system", "my-istio", clients.ReadOpts{})
			return err
		}).ShouldNot(HaveOccurred())

		err = utils.Supergloo("uninstall --name=my-istio")
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() error {
			_, err := kube.CoreV1().Services("istio-system").Get("istio-pilot", metav1.GetOptions{})
			return err
		}, time.Minute*2).Should(HaveOccurred())

		Eventually(func() error {
			_, err := meshClient.Read("supergloo-system", "my-istio", clients.ReadOpts{})
			return err
		}).Should(HaveOccurred())

	})
})
