package utils

import (
	"encoding/pem"
	"time"

	pkiutil "istio.io/istio/security/pkg/pki/util"
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
	certificate = "CERTIFICATE"
)

func GenCertForCSR(
	hosts []string, csrPem, signingCert, privateKey []byte,
) ([]byte, error) {
	// TODO(ilackarms): allow configuring this TTL in the virtual mesh
	ttl := time.Until(time.Now().AddDate(1, 0, 0))

	// The following three function calls allow the input byte arrays to be PEM encoded, so that the caller does not
	// need to pre decode the data.
	cert, err := pkiutil.ParsePemEncodedCertificate(signingCert)
	if err != nil {
		return nil, err
	}
	csr, err := pkiutil.ParsePemEncodedCSR(csrPem)
	if err != nil {
		return nil, err
	}
	key, err := pkiutil.ParsePemEncodedKey(privateKey)
	if err != nil {
		return nil, err
	}

	newCertBytes, err := pkiutil.GenCertFromCSR(
		csr,
		cert,
		csr.PublicKey,
		key,
		hosts,
		ttl,
		true,
	)
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
