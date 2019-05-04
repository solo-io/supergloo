package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"io/ioutil"

	"github.com/solo-io/go-utils/log"
	"k8s.io/client-go/util/cert"
)

// This function generates a self-signed TLS certificate
func main() {

	// Generate the CA certificate that will be used to sign the webhook server certificate
	caPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("Failed to create CA private key: %v", err)
	}
	caCert, err := cert.NewSelfSignedCACert(cert.Config{CommonName: "supergloo-webhook-cert-ca"}, caPrivateKey)
	if err != nil {
		log.Fatalf("Failed to create CA cert: %v", err)
	}
	caCertPEM := cert.EncodeCertPEM(caCert)

	// Generate webhook server certificate
	serverCertPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("Failed to create server cert private key: %v", err)
	}
	serverCertPrivateKeyPEM, err := cert.MarshalPrivateKeyToPEM(serverCertPrivateKey)
	if err != nil {
		log.Fatalf("Failed to create server cert private key: %v", err)
	}
	signedServerCert, err := cert.NewSignedCert(cert.Config{
		CommonName:   "appmesh-sidecar-injector.supergloo-system.svc",
		Organization: []string{"solo.io"},
		AltNames: cert.AltNames{
			DNSNames: []string{
				"appmesh-sidecar-injector",
				"appmesh-sidecar-injector.supergloo-system",
				"appmesh-sidecar-injector.supergloo-system.svc",
				"appmesh-sidecar-injector.supergloo-system.svc.cluster.local",
			},
		},
		Usages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}, serverCertPrivateKey, caCert, caPrivateKey)
	if err != nil {
		log.Fatalf("Failed to create server cert: %v", err)
	}
	signedServerCertPEM := cert.EncodeCertPEM(signedServerCert)

	// Save all the certificate files
	caCertFile := "ca.crt"
	serverCertFile := "cert.pem"
	serverPrivateKey := "key.pem"
	if err := ioutil.WriteFile(caCertFile, caCertPEM, 0644); err != nil {
		log.Fatalf("Failed to write CA cert file: %v", err)
	}
	if err := ioutil.WriteFile(serverCertFile, signedServerCertPEM, 0600); err != nil {
		log.Fatalf("Failed to write server cert file: %v", err)
	}
	if err := ioutil.WriteFile(serverPrivateKey, serverCertPrivateKeyPEM, 0644); err != nil {
		log.Fatalf("Failed to write server key file: %v", err)
	}
}
