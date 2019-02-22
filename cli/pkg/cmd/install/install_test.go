package install_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/supergloo/cli/pkg/helpers"
	"github.com/solo-io/supergloo/cli/test/utils"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"

	. "github.com/solo-io/supergloo/cli/pkg/cmd/install"
)

var _ = Describe("Install", func() {

	BeforeEach(func() {
		helpers.UseMemoryClients()
	})

	getInstall := func(name string) *v1.Install {
		in, err := helpers.MustInstallClient().Read("supergloo-system", name, clients.ReadOpts{})
		Expect(err).NotTo(HaveOccurred())
		return in
	}

	Context("static", func() {
		err := utils.Supergloo("")
		Expect(err).NotTo(HaveOccurred())
	})
})
