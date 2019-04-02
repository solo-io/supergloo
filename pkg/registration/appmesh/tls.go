package appmesh

import (
	"crypto/rand"
	"crypto/rsa"

	"github.com/solo-io/go-utils/errors"
	"k8s.io/client-go/util/cert"
)

type certificates struct {
	// PEM-encoded CA certificate that has been used to sign the server certificate
	caCertificate []byte
	// PEM-encoded server certificate
	serverCertificate []byte
	// PEM-encoded private key that has been used to sign the server certificate
	serverCertKey []byte
}

// This function generates a self-signed TLS certificate
func generateSelfSignedCertificate(config cert.Config) (*certificates, error) {

	// Generate the CA certificate that will be used to sign the webhook server certificate
	caPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create CA private key")
	}
	caCert, err := cert.NewSelfSignedCACert(cert.Config{CommonName: "supergloo-webhook-cert-ca"}, caPrivateKey)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create CA certificate")
	}

	// Generate webhook server certificate
	serverCertPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create server cert private key")
	}
	signedServerCert, err := cert.NewSignedCert(config, serverCertPrivateKey, caCert, caPrivateKey)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create server cert")
	}

	serverCertPrivateKeyPEM, err := cert.MarshalPrivateKeyToPEM(serverCertPrivateKey)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to convert server cert private key to PEM")
	}

	return &certificates{
		caCertificate:     cert.EncodeCertPEM(caCert),
		serverCertificate: cert.EncodeCertPEM(signedServerCert),
		serverCertKey:     serverCertPrivateKeyPEM,
	}, nil
}
