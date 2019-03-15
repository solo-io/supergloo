package create_test

import (
	"fmt"
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/supergloo/cli/pkg/helpers"
	"github.com/solo-io/supergloo/cli/test/utils"
	"github.com/solo-io/supergloo/test/inputs"
)

var _ = Describe("SecretTls", func() {
	var (
		rootCert      *os.File
		caCert        *os.File
		certChain     *os.File
		caKey         *os.File
		secretContent = inputs.InputTlsSecret("", "")
	)
	BeforeEach(func() {
		secretContent.CaCert = "CaCert"
		secretContent.CaKey = "CaKey"
		secretContent.RootCert = "RootCert"
		secretContent.CertChain = "CertChain"
		helpers.UseMemoryClients()
		var err error
		rootCert, err = ioutil.TempFile("", "rootCert")
		Expect(err).NotTo(HaveOccurred())
		caCert, err = ioutil.TempFile("", "caCert")
		Expect(err).NotTo(HaveOccurred())
		certChain, err = ioutil.TempFile("", "certChain")
		Expect(err).NotTo(HaveOccurred())
		caKey, err = ioutil.TempFile("", "caKey")
		Expect(err).NotTo(HaveOccurred())

		_, err = rootCert.Write([]byte(secretContent.RootCert))
		Expect(err).NotTo(HaveOccurred())
		_, err = caCert.Write([]byte(secretContent.CaCert))
		Expect(err).NotTo(HaveOccurred())
		_, err = certChain.Write([]byte(secretContent.CertChain))
		Expect(err).NotTo(HaveOccurred())
		_, err = caKey.Write([]byte(secretContent.CaKey))
		Expect(err).NotTo(HaveOccurred())
	})

	It("creates the tls secrets using create secret tls", func() {
		err := utils.Supergloo(fmt.Sprintf("create secret tls --name poppy "+
			"--cacert %v --cakey %v --rootcert %v --certchain %v",
			caCert.Name(),
			caKey.Name(),
			rootCert.Name(),
			certChain.Name()))
		Expect(err).NotTo(HaveOccurred())

		sec, err := helpers.MustTlsSecretClient().Read("supergloo-system", "poppy", clients.ReadOpts{})
		Expect(err).NotTo(HaveOccurred())
		Expect(sec.CaCert).To(Equal(secretContent.CaCert))
		Expect(sec.CertChain).To(Equal(secretContent.CertChain))
		Expect(sec.CaKey).To(Equal(secretContent.CaKey))
		Expect(sec.RootCert).To(Equal(secretContent.RootCert))
	})
})
