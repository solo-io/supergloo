package surveyutils_test

import (
	"context"
	"time"

	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/gloo/pkg/cliutil/testutil"
	"github.com/solo-io/supergloo/cli/pkg/options"
	. "github.com/solo-io/supergloo/cli/pkg/surveyutils"
)

var _ = Describe("Faultinjection", func() {
	BeforeEach(func() {
		clients.UseMemoryClients()
	})
	It("abort", func() {
		testutil.ExpectInteractive(func(c *testutil.Console) {
			c.ExpectString("select type of fault injection rule")
			c.PressDown()
			c.SendLine("")
			c.ExpectString("select type of abort rule")
			c.SendLine("")
			c.ExpectString("percentage of requests to inject (0-100)")
			c.SendLine("50")
			c.ExpectString("enter status code to abort request with (valid http status code)")
			c.SendLine("404")
			c.ExpectEOF()
		}, func() {
			in := &options.CreateRoutingRule{}
			err := SurveyFaultInjectionSpec(context.TODO(), in)
			Expect(err).NotTo(HaveOccurred())
			Expect(in.RoutingRuleSpec.FaultInjection).To(Equal(options.FaultInjection{
				Percent: 50,
				Abort: options.FaultInjectionAbort{
					Http: v1.FaultInjection_Abort_HttpStatus{HttpStatus: 404},
				},
				Delay: options.FaultInjectionDelay{Fixed: 0},
			}))
		})
	})

	It("delay", func() {
		testutil.ExpectInteractive(func(c *testutil.Console) {
			c.ExpectString("select type of fault injection rule")
			c.SendLine("")
			c.ExpectString("select type of delay rule")
			c.SendLine("")
			c.ExpectString("percentage of requests to inject (0-100)")
			c.SendLine("50")
			c.ExpectString("enter fixed delay duration")
			c.SendLine("1s")
			c.ExpectEOF()
		}, func() {
			in := &options.CreateRoutingRule{}
			err := SurveyFaultInjectionSpec(context.TODO(), in)
			Expect(err).NotTo(HaveOccurred())
			Expect(in.RoutingRuleSpec.FaultInjection).To(Equal(options.FaultInjection{
				Percent: 50,
				Abort: options.FaultInjectionAbort{
					Http: v1.FaultInjection_Abort_HttpStatus{HttpStatus: 0},
				},
				Delay: options.FaultInjectionDelay{Fixed: time.Second * 1},
			}))
		})
	})

})
