package cert_secrets

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CertAndKeyData struct {
	CertChain  []byte
	PrivateKey []byte
	RootCert   []byte
}

var _ CertSecretBuilder = &CertAndKeyData{}

func (c *CertAndKeyData) BuildSecret(name, namespace string) *corev1.Secret {
	return &corev1.Secret{
		Data: map[string][]byte{
			CertChainID:  c.CertChain,
			PrivateKeyID: c.PrivateKey,
			RootCertID:   c.RootCert,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Type: CertAndKeySecretType,
	}
}

type RootCaData struct {
	CertAndKeyData
	CaCert       []byte
	CaPrivateKey []byte
}

var _ CertSecretBuilder = &RootCaData{}

func (r *RootCaData) BuildSecret(name, namespace string) *corev1.Secret {
	return &corev1.Secret{
		Data: map[string][]byte{
			CertChainID:    r.CertChain,
			PrivateKeyID:   r.PrivateKey,
			RootCertID:     r.RootCert,
			CaCertID:       r.CaCert,
			CaPrivateKeyID: r.CaPrivateKey,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Type: RootCertSecretType,
	}
}

func RootCaDataFromSecret(secret *corev1.Secret) (*RootCaData, error) {
	caKey, ok := secret.Data[CaPrivateKeyID]
	if !ok {
		return nil, NoCaKeyFoundError(secret.ObjectMeta)
	}
	caCert, ok := secret.Data[CaCertID]
	if !ok {
		return nil, NoCaCertFoundError(secret.ObjectMeta)
	}
	certAndKey, err := CertAndKeyDataFromSecret(secret)
	if err != nil {
		return nil, err
	}
	return &RootCaData{
		CertAndKeyData: *certAndKey,
		CaCert:         caCert,
		CaPrivateKey:   caKey,
	}, nil
}

func CertAndKeyDataFromSecret(secret *corev1.Secret) (*CertAndKeyData, error) {
	rootCert, ok := secret.Data[RootCertID]
	if !ok {
		return nil, NoRootCertFoundError(secret.ObjectMeta)
	}
	privateKey, ok := secret.Data[PrivateKeyID]
	if !ok {
		return nil, NoPrivateKeyFoundError(secret.ObjectMeta)
	}
	certChain, ok := secret.Data[CertChainID]
	if !ok {
		return nil, NoCertChainFoundError(secret.ObjectMeta)
	}
	return &CertAndKeyData{
		CertChain:  certChain,
		PrivateKey: privateKey,
		RootCert:   rootCert,
	}, nil
}
