package certgen

import (
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"
	"time"

	"github.com/rotisserie/eris"
	pki_util "istio.io/istio/security/pkg/pki/util"
)

/*
	The reason for these constants stem from the golang pem package
	https://golang.org/pkg/encoding/pem/#Block

	a pem encoded block has the form:

	-----BEGIN Type-----
	Headers
	base64-encoded Bytes
	-----END Type-----

	The constants below are the BEGIN and END strings to instruct the encoder/decoder how to properly format the data
*/
const (
	certificate        = "CERTIFICATE"
	certificateRequest = "CERTIFICATE REQUEST"
)

func YearDuration() time.Duration {
	return time.Until(time.Now().AddDate(1, 0, 0))
}

var InvalidKeyFormattingError = func(err error) error {
	return eris.Wrapf(err, "unable to decode private key, currently only supporting PKCS1 encrypted keys")
}

//go:generate mockgen -destination ./mocks/mock_signer.go -source certgen.go

// Signer is meant as a higher level abstraction around complex certificate workflows
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

func NewSigner() Signer {
	return &signer{}
}

type signer struct{}

func (s *signer) GenCSRWithKey(options pki_util.CertOptions) (csr []byte, err error) {
	// If the signer priv is non-nil use that as the signer key
	priv := options.SignerPriv
	if priv == nil {
		// Attempt to decode the key from the PEM format, currently only one format is supported (PKCS1)
		block, _ := pem.Decode(options.SignerPrivPem)
		priv, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, InvalidKeyFormattingError(err)
		}
	}

	template, err := pki_util.GenCSRTemplate(options)
	if err != nil {
		return nil, fmt.Errorf("CSR template creation failed (%v)", err)
	}

	csr, err = x509.CreateCertificateRequest(rand.Reader, template, priv)
	if err != nil {
		return nil, err
	}

	// Encode the csr to PEM format before returning
	block := &pem.Block{
		Type:  certificateRequest,
		Bytes: csr,
	}
	csrByt := pem.EncodeToMemory(block)
	return csrByt, nil
}

func (s *signer) GenCertFromEncodedCSR(
	csrPem, signingCert, privateKey []byte,
	subjectIDs []string,
	ttl time.Duration,
	isCA bool,
) ([]byte, error) {
	// The following three function calls allow the input byte arrays to be PEM encoded, so that the caller does not
	// need to pre decode the data.
	cert, err := pki_util.ParsePemEncodedCertificate(signingCert)
	if err != nil {
		return nil, err
	}
	csr, err := pki_util.ParsePemEncodedCSR(csrPem)
	if err != nil {
		return nil, err
	}
	key, err := pki_util.ParsePemEncodedKey(privateKey)
	if err != nil {
		return nil, err
	}

	newCertBytes, err := pki_util.GenCertFromCSR(csr, cert, csr.PublicKey, key, subjectIDs, ttl, isCA)
	if err != nil {
		return nil, err
	}
	// This block is the go way to encode the cert into the PEM format before returning it
	block := &pem.Block{
		Type:  certificate,
		Bytes: newCertBytes,
	}
	return pem.EncodeToMemory(block), nil
}

func (s *signer) GenCSR(options pki_util.CertOptions) (csr, privKey []byte, err error) {
	return pki_util.GenCSR(options)
}

/*
	AppendRootCerts appends the root mesh group cert to the generated CaCert, It is yanked from the following Istio
	function:

	https://github.com/istio/istio/blob/5218a80f97cb61ff4a02989b7d9f8c4fda50780f/security/pkg/pki/util/generate_csr.go#L95

	Certificate chains are necessary to verify the authenticity of a certficicate, in this case the authenticity of
	the generated Ca Certificate against the MeshGroup root cert
*/
func AppendRootCerts(caCert, rootCert []byte) []byte {
	var caCertCopy []byte
	if len(caCert) > 0 {
		// Copy the input certificate
		caCertCopy = make([]byte, len(caCert))
		copy(caCertCopy, caCert)
	}
	if len(rootCert) > 0 {
		if len(caCertCopy) > 0 {
			// Append a newline after the last cert
			// Certs are very fooey, this is copy pasted from Istio, plz do not touch
			// Love, eitan
			caCertCopy = []byte(strings.TrimSuffix(string(caCertCopy), "\n") + "\n")
		}
		caCertCopy = append(caCertCopy, rootCert...)
	}
	return caCertCopy
}
