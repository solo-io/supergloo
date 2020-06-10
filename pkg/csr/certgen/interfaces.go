package certgen

import (
	"time"

	networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/types"
	cert_secrets "github.com/solo-io/service-mesh-hub/pkg/csr/certgen/secrets"
	pki_util "istio.io/istio/security/pkg/pki/util"
)

//go:generate mockgen -destination ./mocks/mocks.go -source interfaces.go

// Certificate generation methods
type RootCertGenerator interface {
	// Generate a new private key and use it to create a self-signed root certificate
	GenRootCertAndKey(builtinCA *networking_types.VirtualMeshSpec_CertificateAuthority_Builtin) (*cert_secrets.RootCAData, error)
}

// Signer is a higher level abstraction around complex certificate workflows
type Signer interface {
	/*
		GenCerFromCSR generates a pem encoded certificate given the parameters

		csrPem: The pem encoded csr bytes
		signingCertPem: the pem encoded certificate which should sign the new certificate
		signingKey: The pem encoded private key which should sign the new cert
		subjectIds: The subjects which the new cert should apply to
		ttl: time to live, the duration of the certificate
		isCa: Whether or not the new cert is a Certificate Authority
	*/
	GenCertFromEncodedCSR(
		csrPem, signingCertPem, signingKey []byte,
		subjectIDs []string,
		ttl time.Duration,
		isCA bool,
	) (cert []byte, err error)
	/*
		GenCSRWithKey generates a pem encoded csr with the key provided.
		The key can be provided via:
			1. options.SignerKey: This must be a crypto.Signer
			2. options.SignerKeyPem: Pem encoded private key which will be unmarshalled into a crypto.Signer().
			Currently this must be a PKCS1 encoded key
		The hosts should be the Spiffe hosts identities
		The org should be the organization the cert is being created for.
	*/
	GenCSRWithKey(options pki_util.CertOptions) (csr []byte, err error)
	// GenCSR functions the same as GenCSRWithKey except that it creates a new key and returns it along with the CSR
	GenCSR(options pki_util.CertOptions) (csr, privKey []byte, err error)
}
