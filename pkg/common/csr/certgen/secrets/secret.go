package cert_secrets

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// The root CA from the perspective of the MeshGroup
// A user supplied root cert may be itself derived from another CA, but
// that is irrelevant for the MeshGroup.
type RootCAData struct {
	PrivateKey []byte
	RootCert   []byte
}

func (c *RootCAData) BuildSecret(name, namespace string) *corev1.Secret {
	return &corev1.Secret{
		Data: map[string][]byte{
			RootPrivateKeyID: c.PrivateKey,
			RootCertID:       c.RootCert,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Type: RootCertSecretType,
	}
}

// The intermediate CA derived from the root CA of the MeshGroup
type IntermediateCAData struct {
	RootCAData
	CertChain    []byte
	CaCert       []byte
	CaPrivateKey []byte
}

var _ CertSecretBuilder = &IntermediateCAData{}

func (r *IntermediateCAData) BuildSecret(name, namespace string) *corev1.Secret {
	return &corev1.Secret{
		Data: map[string][]byte{
			CertChainID:      r.CertChain,
			RootPrivateKeyID: r.PrivateKey,
			RootCertID:       r.RootCert,
			CaCertID:         r.CaCert,
			CaPrivateKeyID:   r.CaPrivateKey,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Type: IntermediateCertSecretType,
	}
}

func IntermediateCADataFromSecret(secret *corev1.Secret) (*IntermediateCAData, error) {
	caKey, ok := secret.Data[CaPrivateKeyID]
	if !ok {
		return nil, NoCaKeyFoundError(secret.ObjectMeta)
	}
	caCert, ok := secret.Data[CaCertID]
	if !ok {
		return nil, NoCaCertFoundError(secret.ObjectMeta)
	}
	certChain, ok := secret.Data[CertChainID]
	if !ok {
		return nil, NoCertChainFoundError(secret.ObjectMeta)
	}
	rootCAData, err := RootCADataFromSecret(secret)
	if err != nil {
		return nil, err
	}
	return &IntermediateCAData{
		RootCAData:   *rootCAData,
		CertChain:    certChain,
		CaCert:       caCert,
		CaPrivateKey: caKey,
	}, nil
}

func RootCADataFromSecret(secret *corev1.Secret) (*RootCAData, error) {
	rootCert, ok := secret.Data[RootCertID]
	if !ok {
		return nil, NoRootCertFoundError(secret.ObjectMeta)
	}
	privateKey, ok := secret.Data[RootPrivateKeyID]
	if !ok {
		return nil, NoPrivateKeyFoundError(secret.ObjectMeta)
	}
	return &RootCAData{
		PrivateKey: privateKey,
		RootCert:   rootCert,
	}, nil
}
