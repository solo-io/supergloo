package surveyutils_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/gloo/pkg/cliutil/testutil"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/cli/pkg/helpers"
	"github.com/solo-io/supergloo/cli/pkg/options"
	"github.com/solo-io/supergloo/test/inputs"

	. "github.com/solo-io/supergloo/cli/pkg/surveyutils"
)

// TODO: unexclude this test when c.ExpectEOF() is fixed
var _ = XDescribe("RoutingRule", func() {
	BeforeEach(func() {
		helpers.UseMemoryClients()
		for _, us := range inputs.BookInfoUpstreams("hi") {
			_, err := helpers.MustUpstreamClient().Write(us, clients.WriteOpts{})
			Expect(err).NotTo(HaveOccurred())
		}
	})
	Context("upstream selector", func() {
		It("selects upstreams from the list until skip is selected", func() {
			testutil.ExpectInteractive(func(c *testutil.Console) {
				c.ExpectString("create a source selector for this rule? ")
				c.SendLine("y")
				c.ExpectString("what kind of selector would you like to create? ")
				c.PressDown()
				c.SendLine("")
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
				in := &options.CreateRoutingRule{}
				err := SurveyRoutingRule(context.TODO(), in)
				Expect(err).NotTo(HaveOccurred())
				Expect(in.SourceSelector).To(Equal(options.Selector{
					Enabled:      true,
					SelectorType: options.SelectorType_Upstream,
					SelectedUpstreams: []core.ResourceRef{
						{
							Name:      "default-details-9080",
							Namespace: "hi",
						},
						{
							Name:      "default-details-v1-9080",
							Namespace: "hi",
						},
					},
				}))
			})
		})
	})
})
