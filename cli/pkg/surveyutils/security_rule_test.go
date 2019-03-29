package surveyutils

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/gloo/pkg/cliutil/testutil"
	"github.com/solo-io/supergloo/cli/pkg/options"
)

var _ = Describe("SecurityRule", func() {
	Context("surveyAllowedPaths", func() {
		It("queries the user for a comma-separated list of paths", func() {
			testutil.ExpectInteractive(func(c *testutil.Console) {
				c.ExpectString("enter a comma-separated list of HTTP paths to allow for " +
					"this rule, e.g.: /api,/admin,/auth (leave empty to allow all):")
				c.SendLine("/api,/admin,/auth")
				c.ExpectEOF()
			}, func() {
				in := &options.CreateSecurityRule{}
				err := surveyAllowedPaths(context.TODO(), in)
				Expect(err).NotTo(HaveOccurred())
				Expect(in.AllowedPaths).To(Equal([]string{"/api", "/admin", "/auth"}))
			})
		})
	})

	Context("surveyAllowedMethods", func() {
		It("queries the user for a comma-separated list of methods", func() {
			testutil.ExpectInteractive(func(c *testutil.Console) {
				c.ExpectString("enter a comma-separated list of HTTP methods to allow for " +
					"this rule, e.g.: GET,POST,PATCH (leave empty to allow all):")
				c.SendLine("GET,POST,PATCH")
				c.ExpectEOF()
			}, func() {
				in := &options.CreateSecurityRule{}
				err := surveyAllowedMethods(context.TODO(), in)
				Expect(err).NotTo(HaveOccurred())
				Expect(in.AllowedMethods).To(Equal([]string{"GET", "POST", "PATCH"}))
			})
		})
	})
})
