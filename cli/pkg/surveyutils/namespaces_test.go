package surveyutils_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/gloo/pkg/cliutil/testutil"
	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"
	. "github.com/solo-io/supergloo/cli/pkg/surveyutils"
)

var _ = Describe("SelectUpstreams", func() {
	BeforeEach(func() {
		clients.UseMemoryClients()
	})
	It("selects upstreams from the list until skip is selected", func() {
		testutil.ExpectInteractive(func(c *testutil.Console) {
			c.ExpectString("add a namespace (choose <done> to finish): ")
			c.PressDown()
			c.SendLine("")
			c.ExpectString("add a namespace (choose <done> to finish): ")
			c.PressDown()
			c.PressDown()
			c.SendLine("")
			c.ExpectString("add a namespace (choose <done> to finish): ")
			c.SendLine("")
			c.ExpectEOF()
		}, func() {
			nss, err := SurveyNamespaces()
			Expect(err).NotTo(HaveOccurred())
			Expect(nss).To(Equal([]string{"default", "supergloo-system"}))
		})
	})
})
