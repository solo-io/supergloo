package surveyutils_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/gloo/pkg/cliutil/testutil"
	"github.com/solo-io/supergloo/cli/pkg/options"

	. "github.com/solo-io/supergloo/cli/pkg/surveyutils"
)

var _ = Describe("TlsSecret", func() {
	It("populates tls secret input", func() {

		testutil.ExpectInteractive(func(c *testutil.Console) {
			c.ExpectString("path to root-cert file")
			c.SendLine("/some/path/root-cert.pem")
			c.ExpectString("path to ca-cert file")
			c.SendLine("/some/path/ca-cert.pem")
			c.ExpectString("path to ca-key file")
			c.SendLine("/some/path/ca-key.pem")
			c.ExpectString("path to cert-chain file")
			c.SendLine("/some/path/cert-chain.pem")
			c.ExpectEOF()
		}, func() {
			var secret options.CreateTlsSecret
			err := SurveyTlsSecret(&secret)
			Expect(err).NotTo(HaveOccurred())

			Expect(secret.RootCaFilename).To(Equal("/some/path/root-cert.pem"))
			Expect(secret.CaCertFilename).To(Equal("/some/path/ca-cert.pem"))
			Expect(secret.PrivateKeyFilename).To(Equal("/some/path/ca-key.pem"))
			Expect(secret.CertChainFilename).To(Equal("/some/path/cert-chain.pem"))
		})
	})
})
