package surveyutils_test

import (
	"context"

	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/gloo/pkg/cliutil/testutil"
	skclients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/supergloo/cli/pkg/options"
	"github.com/solo-io/supergloo/test/inputs"

	. "github.com/solo-io/supergloo/cli/pkg/surveyutils"
)

var _ = Describe("Uninstall", func() {
	BeforeEach(func() {
		clients.UseMemoryClients()
	})

	It("should create the expected install ", func() {

		name1 := "input1"
		name2 := "input2"
		namespace := "ns"
		inst1 := inputs.IstioInstall(name1, namespace, "any", "1.0.5", false)
		inst2 := inputs.IstioInstall(name2, namespace, "any", "1.0.5", false)
		ic := clients.MustInstallClient()
		_, err := ic.Write(inst1, skclients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())
		_, err = ic.Write(inst2, skclients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())

		testutil.ExpectInteractive(func(c *testutil.Console) {
			c.ExpectString("which install to uninstall? ")
			c.PressDown()
			c.SendLine("")
			c.ExpectEOF()
		}, func() {
			in := options.Options{Ctx: context.TODO()}
			err := SurveyUninstall(&in)
			Expect(err).NotTo(HaveOccurred())
			Expect(in.Uninstall.Metadata.Name).To(Equal(name2))
		})
	})
})
