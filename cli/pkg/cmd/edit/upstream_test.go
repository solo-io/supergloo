package edit_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"
	"github.com/solo-io/supergloo/cli/test/utils"
)

var _ = Describe("edit upstream", func() {
	BeforeEach(func() {
		clients.UseMemoryClients()
	})
	Context("validation", func() {

		It("returns err if name or namespace of mesh aren't present", func() {
			err := utils.Supergloo("edit upstream tls --mesh-namespace one")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("mesh resource name and namespace must be specified"))
		})

		It("returns err if name or namespace of upstream aren't present", func() {
			err := utils.Supergloo("edit upstream tls --namespace one --mesh-name one")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("upstream name and namespace must be specified"))
		})
	})

})
