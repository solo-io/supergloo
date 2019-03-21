package uninstall_test

import (
	"fmt"

	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"

	"github.com/solo-io/supergloo/test/inputs"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	skclients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/supergloo/cli/test/utils"
)

var _ = Describe("Uninstall", func() {

	BeforeEach(func() {
		clients.UseMemoryClients()
	})

	It("should enable an existing + disabled install", func() {
		name := "input"
		namespace := "ns"
		inst := inputs.IstioInstall(name, namespace, "any", "1.0.5", false)
		ic := clients.MustInstallClient()
		_, err := ic.Write(inst, skclients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())

		err = utils.Supergloo("uninstall istio " +
			fmt.Sprintf("--name=%v ", name) +
			fmt.Sprintf("--namespace=%v ", namespace))
		Expect(err).NotTo(HaveOccurred())

		updatedInstall, err := ic.Read(namespace, name, skclients.ReadOpts{})
		Expect(err).NotTo(HaveOccurred())
		Expect(updatedInstall.Disabled).To(BeTrue())

	})
})
