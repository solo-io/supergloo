package surveyutils_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/gloo/pkg/cliutil/testutil"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/cli/pkg/helpers"
	"github.com/solo-io/supergloo/cli/pkg/options"
	"github.com/solo-io/supergloo/test/inputs"

	. "github.com/solo-io/supergloo/cli/pkg/surveyutils"
)

var _ = Describe("Trafficshifting", func() {
	BeforeEach(func() {
		helpers.UseMemoryClients()
		for _, us := range inputs.BookInfoUpstreams("hi") {
			_, err := helpers.MustUpstreamClient().Write(us, clients.WriteOpts{})
			Expect(err).NotTo(HaveOccurred())
		}
	})
	Context("surveyMatcher", func() {
		It("fills in the traffic shifting config from user input", func() {
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
				c.ExpectString("choose a weight for {default-details-9080 hi}")
				c.SendLine("5")
				c.ExpectString("choose a weight for {default-details-v1-9080 hi}")
				c.SendLine("10")
				c.ExpectEOF()
			}, func() {
				in := &options.CreateRoutingRule{}
				err := SurveyTrafficShiftingSpec(context.TODO(), in)
				Expect(err).NotTo(HaveOccurred())
				Expect(in.RoutingRuleSpec.TrafficShifting).To(Equal(options.TrafficShiftingValue{
					Destinations: &gloov1.MultiDestination{
						Destinations: []*gloov1.WeightedDestination{
							{
								Destination: &gloov1.Destination{
									Upstream: core.ResourceRef{
										Name:      "default-details-9080",
										Namespace: "hi",
									},
								},
								Weight: 5,
							},
							{
								Destination: &gloov1.Destination{
									Upstream: core.ResourceRef{
										Name:      "default-details-v1-9080",
										Namespace: "hi",
									},
								},
								Weight: 10,
							},
						},
					},
				}))
			})
		})
	})
})
