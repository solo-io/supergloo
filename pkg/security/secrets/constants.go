package cert_secrets

import (
	corev1 "k8s.io/api/core/v1"
)

const (
	IntermediateCertSecretType corev1.SecretType = "solo.io/ca-intermediate"

	RootCertSecretType corev1.SecretType = "solo.io/ca-root"

	// CaCertID is the CA certificate chain file.
	CaCertID = "ca-cert.pem"
	// CaPrivateKeyID is the private key file of CA.
	CaPrivateKeyID = "ca-key.pem"
	// CertChainID is the ID/name for the certificate chain file.
	CertChainID = "cert-chain.pem"
	// RootPrivateKeyID is the ID/name for the private key file.
	// Unfortunately has to be `key.pem`, not `root-key.pem` to match istio :(
	RootPrivateKeyID = "key.pem"
	// RootCertID is the ID/name for the CA root certificate file.
	RootCertID = "root-cert.pem"
)
