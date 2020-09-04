package utils_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/service-mesh-hub/pkg/certificates/agent/utils"
	. "github.com/solo-io/service-mesh-hub/pkg/certificates/issuer/utils"
	"istio.io/istio/security/pkg/pki/util"
)

var _ = Describe("CertGen workflow", func() {
	assertCsrWorks := func(signingRoot, signingKey []byte) {
		privateKey, err := utils.GeneratePrivateKey()
		Expect(err).NotTo(HaveOccurred())

		hosts := []string{"spiffe://custom-domain/ns/istio-system/sa/istio-pilot-service-account"}
		csr, err := utils.GenerateCertificateSigningRequest(
			hosts,
			"service-mesh-hub",
			privateKey,
		)
		Expect(err).NotTo(HaveOccurred())

		inetermediaryCert, err := GenCertForCSR(
			hosts,
			csr,
			signingRoot,
			signingKey,
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(inetermediaryCert)).To(ContainSubstring("-----BEGIN CERTIFICATE-----"))
	}

	It("generates a certificate using generated self signed cert, private key, and certificate signing request", func() {

		options := util.CertOptions{
			Org:          "org",
			IsCA:         true,
			IsSelfSigned: true,
			TTL:          time.Hour * 24 * 365,
			RSAKeySize:   4096,
			PKCS8Key:     false, // currently only supporting PKCS1
		}
		signingRoot, signingKey, err := util.GenCertKeyFromOptions(options)
		Expect(err).NotTo(HaveOccurred())

		assertCsrWorks(signingRoot, signingKey)
	})
})
