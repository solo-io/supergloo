package surveyutils_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/gloo/pkg/cliutil/testutil"
	"github.com/solo-io/supergloo/cli/pkg/flagutils"
	"github.com/solo-io/supergloo/cli/pkg/options"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"

	. "github.com/solo-io/supergloo/cli/pkg/surveyutils"
)

var _ = FDescribe("Retries", func() {
	It("sets values for MaxRetries", func() {
		testutil.ExpectInteractive(func(c *testutil.Console) {
			c.ExpectString(flagutils.Description_MaxRetries_Attempts)
			c.SendLine("5")
			c.ExpectString(flagutils.Description_MaxRetries_PerTryTimeout)
			c.SendLine("1m")
			c.ExpectString(flagutils.Description_MaxRetries_RetryOn)
			c.SendLine("5xx")
			c.ExpectEOF()
		}, func() {
			in := &options.MaxRetries{}
			err := SurveyMaxRetries(in)
			Expect(err).NotTo(HaveOccurred())
			Expect(in.Attempts).To(Equal(uint32(5)))
			Expect(in.PerTryTimeout).To(Equal(time.Minute))
			Expect(in.RetryOn).To(Equal("5xx"))
		})
	})
})

var _ = Describe("Retries", func() {
	It("sets values for MaxRetries", func() {
		testutil.ExpectInteractive(func(c *testutil.Console) {
			c.ExpectString(flagutils.Description_RetryBudget_RetryRatio)
			c.SendLine("0.5")
			c.ExpectString(flagutils.Description_RetryBudget_MinRetriesPerSecond)
			c.SendLine("5")
			c.ExpectString(flagutils.Description_RetryBudget_Ttl)
			c.SendLine("1m")
			c.ExpectEOF()
		}, func() {
			in := &v1.RetryBudget{}
			err := SurveyRetryBudget(in)
			Expect(err).NotTo(HaveOccurred())
			Expect(in.RetryRatio).To(Equal(float32(0.5)))
			Expect(in.MinRetriesPerSecond).To(Equal(uint32(5)))
			Expect(in.Ttl).To(Equal(time.Minute))
		})
	})
})
