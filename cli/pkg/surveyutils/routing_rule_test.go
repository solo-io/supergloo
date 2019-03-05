package surveyutils

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
)

// TODO: unexclude this test when c.ExpectEOF() is fixed
// relevant issue: https://github.com/solo-io/gloo/issues/387
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

var _ = Describe("RoutingRule", func() {
	BeforeEach(func() {
		helpers.UseMemoryClients()
		for _, us := range inputs.BookInfoUpstreams("hi") {
			_, err := helpers.MustUpstreamClient().Write(us, clients.WriteOpts{})
			Expect(err).NotTo(HaveOccurred())
		}
	})
	Context("surveyMatcher", func() {
		It("fills in the matcher from user input", func() {
			testutil.ExpectInteractive(func(c *testutil.Console) {
				c.ExpectString("add a request matcher for this rule?")
				c.SendLine("y")
				c.ExpectString("Choose a path match type:")
				c.PressDown()
				c.SendLine("")
				c.ExpectString("What path regex should we match?")
				c.SendLine("/foo")
				c.ExpectString("enter a key-value pair in the format KEY=VAL. leave empty to finish: ")
				c.SendLine("")
				c.ExpectString("HTTP Method to match for this route (empty to skip)?")
				c.SendLine("GET")
				c.ExpectString("HTTP Method to match for this route (empty to skip)?")
				c.SendLine("")
				c.ExpectString("add a request matcher for this rule?")
				c.SendLine("N")
				c.ExpectEOF()
			}, func() {
				in := &options.RequestMatchersValue{}
				err := surveyMatcher(in)
				Expect(err).NotTo(HaveOccurred())
				Expect(*in).To(Equal(options.RequestMatchersValue{
					{PathPrefix: "", PathExact: "", PathRegex: "/foo", Methods: []string{"GET"}, HeaderMatcher: nil},
				}))
			})
		})
	})
})
