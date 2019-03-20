package create_test

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	skclients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
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
		clients.UseMemoryClients()
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
	AfterEach(func() {
		os.Remove(rootCert.Name())
		os.Remove(caCert.Name())
		os.Remove(certChain.Name())
		os.Remove(caKey.Name())
	})

	It("creates the tls secrets using create secret tls", func() {
		err := utils.Supergloo(fmt.Sprintf("create secret tls --name poppy "+
			"--cacert %v --cakey %v --rootcert %v --certchain %v",
			caCert.Name(),
			caKey.Name(),
			rootCert.Name(),
			certChain.Name()))
		Expect(err).NotTo(HaveOccurred())

		sec, err := clients.MustTlsSecretClient().Read("supergloo-system", "poppy", skclients.ReadOpts{})
		Expect(err).NotTo(HaveOccurred())
		Expect(sec.CaCert).To(Equal(secretContent.CaCert))
		Expect(sec.CertChain).To(Equal(secretContent.CertChain))
		Expect(sec.CaKey).To(Equal(secretContent.CaKey))
		Expect(sec.RootCert).To(Equal(secretContent.RootCert))
	})
})
