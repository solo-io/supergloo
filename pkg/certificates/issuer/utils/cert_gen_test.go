package utils_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "github.com/solo-io/external-apis/pkg/api/k8s/core/v1"
	"github.com/solo-io/service-mesh-hub/pkg/certificates/agent/utils"
	"github.com/solo-io/service-mesh-hub/pkg/certificates/common/secrets"
	"istio.io/istio/security/pkg/pki/util"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	. "github.com/solo-io/service-mesh-hub/pkg/certificates/issuer/utils"
)

var _ = Describe("CertGen workflow", func() {
	assertCsrWorks := func(signingRoot, signingKey []byte) {
		privateKey, err := utils.GeneratePrivateKey()
		Expect(err).NotTo(HaveOccurred())

		csr, err := utils.GenerateCertificateSigningRequest(
			[]string{"spiffe://cluster.local/ns/istio-system/sa/istio-pilot-service-account"},
			"service-mesh-hub",
			privateKey,
		)
		Expect(err).NotTo(HaveOccurred())

		inetermediaryCert, err := GenCertForCSR(
			csr,
			signingRoot,
			signingKey,
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(inetermediaryCert)).To(ContainSubstring("-----BEGIN CERTIFICATE-----"))
	}

	It("generates a certificate using generated self signed cert, private key, and certificate signing request", func() {
		cli, err := client.New(config.GetConfigOrDie(), client.Options{})
		Expect(err).NotTo(HaveOccurred())

		secret, err := v1.NewSecretClient(cli).GetSecret(context.TODO(), client.ObjectKey{
			Name:      "istiod-istio-system-remote-cluster",
			Namespace: "service-mesh-hub",
		})
		Expect(err).NotTo(HaveOccurred())

		rootCaData := secrets.RootCADataFromSecretData(secret.Data)

		assertCsrWorks(rootCaData.RootCert, rootCaData.PrivateKey)

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
