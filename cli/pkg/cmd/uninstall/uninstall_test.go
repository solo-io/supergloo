package uninstall_test

import (
	"fmt"

	"github.com/solo-io/supergloo/test/inputs"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/supergloo/cli/pkg/helpers"
	"github.com/solo-io/supergloo/cli/test/utils"
)

var _ = Describe("Uninstall", func() {

	BeforeEach(func() {
		helpers.UseMemoryClients()
	})

	It("should enable an existing + disabled install", func() {
		name := "input"
		namespace := "ns"
		inst := inputs.IstioInstall(name, namespace, "any", "1.0.5", false)
		ic := helpers.MustInstallClient()
		_, err := ic.Write(inst, clients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())

		err = utils.Supergloo("uninstall istio " +
			fmt.Sprintf("--name=%v ", name) +
			fmt.Sprintf("--namespace=%v ", namespace))
		Expect(err).NotTo(HaveOccurred())

		updatedInstall, err := ic.Read(namespace, name, clients.ReadOpts{})
		Expect(err).NotTo(HaveOccurred())
		Expect(updatedInstall.Disabled).To(BeTrue())

	})
})
