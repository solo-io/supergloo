package surveyutils_test

import (
	"context"

	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/gloo/pkg/cliutil/testutil"
	skclients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/test/inputs"

	. "github.com/solo-io/supergloo/cli/pkg/surveyutils"
)

var _ = Describe("SelectUpstreams", func() {
	BeforeEach(func() {
		clients.UseMemoryClients()
		for _, us := range inputs.BookInfoUpstreams("hi") {
			_, err := clients.MustUpstreamClient().Write(us, skclients.WriteOpts{})
			Expect(err).NotTo(HaveOccurred())
		}
	})
	It("selects upstreams from the list until skip is selected", func() {
		testutil.ExpectInteractive(func(c *testutil.Console) {
			c.ExpectString("add an upstream (choose <done> to finish): ")
			c.PressDown()
			c.SendLine("")
			c.ExpectString("add an upstream (choose <done> to finish): ")
			c.PressDown()
			c.PressDown()
			c.SendLine("")
			c.ExpectString("add an upstream (choose <done> to finish): ")
			c.SendLine("")
			c.ExpectEOF()
		}, func() {
			ups, err := SurveyUpstreams(context.TODO())
			Expect(err).NotTo(HaveOccurred())
			Expect(ups).To(Equal([]core.ResourceRef{
				{
					Name:      "default-details-9080",
					Namespace: "hi",
				},
				{
					Name:      "default-details-v1-9080",
					Namespace: "hi",
				},
			}))
		})
	})
})
