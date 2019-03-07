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
			Expect(i.Status.Reason).To(Equal(""))
			return i.Status.State, nil
		}, time.Minute*2).Should(Equal(core.Status_Accepted))

		eventuallyGetServiceWithError("istio-system", "istio-pilot").ShouldNot(HaveOccurred())

		eventuallyGetMeshWithError("supergloo-system", "my-istio").ShouldNot(HaveOccurred())

		err = utils.Supergloo("uninstall --name=my-istio")
		Expect(err).NotTo(HaveOccurred())

		eventuallyGetServiceWithError("istio-system", "istio-pilot").Should(HaveOccurred())

		eventuallyGetMeshWithError("supergloo-system", "my-istio").Should(HaveOccurred())

	})
})

func eventuallyGetServiceWithError(namespace, svcName string) GomegaAsyncAssertion {
	return Eventually(func() error {
		_, err := kube.CoreV1().Services(namespace).Get(svcName, metav1.GetOptions{})
		return err
	}, time.Minute*2)
}

func eventuallyGetMeshWithError(namespace, svcName string) GomegaAsyncAssertion {
	meshClient := helpers.MustMeshClient()
	return Eventually(func() error {
		_, err := meshClient.Read(namespace, svcName, clients.ReadOpts{})
		return err
	}, time.Minute*2)
}
