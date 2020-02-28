package cert_secrets

import (
	corev1 "k8s.io/api/core/v1"
)

const (
	RootCertSecretType corev1.SecretType = "solo.io/ca-root"

	CertAndKeySecretType corev1.SecretType = "solo.io/cert-and-key"

	// CaCertID is the CA certificate chain file.
	CaCertID = "ca-cert.pem"
	// CaPrivateKeyID is the private key file of CA.
	CaPrivateKeyID = "ca-key.pem"
	// CertChainID is the ID/name for the certificate chain file.
	CertChainID = "cert-chain.pem"
	// PrivateKeyID is the ID/name for the private key file.
	PrivateKeyID = "key.pem"
	// RootCertID is the ID/name for the CA root certificate file.
	RootCertID = "root-cert.pem"
)
