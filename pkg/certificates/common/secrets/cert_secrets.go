package secrets

const (
	/*
		TODO(ilackarms): document the expected structure of secrets (required for VirtualMeshes  using a user-provided root CA)
	*/
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
	// TLS certificate file for istio gateway credential
	GatewayCertID = "cert"
	// TLS key file for istio gateway credential
	GatewayKeyID = "key"
	// CA certificate file for  gateway credential
	GatewayCaCertID = "cacert"
)

// The root CA from the perspective of the MeshGroup
// A user supplied root cert may be itself derived from another CA, but
// that is irrelevant for the MeshGroup.
type RootCAData struct {
	PrivateKey []byte
	RootCert   []byte
}

func RootCADataFromSecretData(data map[string][]byte) RootCAData {
	rootCert := data[RootCertID]
	privateKey := data[RootPrivateKeyID]
	return RootCAData{
		PrivateKey: privateKey,
		RootCert:   rootCert,
	}
}

func (c *RootCAData) ToSecretData() map[string][]byte {
	return map[string][]byte{
		RootPrivateKeyID: c.PrivateKey,
		RootCertID:       c.RootCert,
	}
}

// The intermediate CA derived from the root CA of the MeshGroup
type IntermediateCAData struct {
	RootCAData
	CertChain    []byte
	CaCert       []byte
	CaPrivateKey []byte
}

func (d IntermediateCAData) ToSecretData() map[string][]byte {
	return map[string][]byte{
		CertChainID:      d.CertChain,
		RootPrivateKeyID: d.PrivateKey,
		RootCertID:       d.RootCert,
		CaCertID:         d.CaCert,
		CaPrivateKeyID:   d.CaPrivateKey,
	}
}

func IntermediateCADataFromSecretData(data map[string][]byte) IntermediateCAData {
	caKey := data[CaPrivateKeyID]
	caCert := data[CaCertID]
	certChain := data[CertChainID]
	rootCAData := RootCADataFromSecretData(data)
	return IntermediateCAData{
		RootCAData:   rootCAData,
		CertChain:    certChain,
		CaCert:       caCert,
		CaPrivateKey: caKey,
	}
}

// Credential for gateways of type mutual
type GatewayMTLSCredentialData struct {
	CaCert []byte
	Cert   []byte
	Key    []byte
}

func (d GatewayMTLSCredentialData) ToSecretData() map[string][]byte {
	return map[string][]byte{
		GatewayCaCertID: d.CaCert,
		GatewayCertID:   d.Cert,
		GatewayKeyID:    d.Key,
	}
}

func GatewayMTLSCredentialDataFromSecretData(data map[string][]byte) GatewayMTLSCredentialData {
	return GatewayMTLSCredentialData{
		CaCert: data[GatewayCaCertID],
		Cert:   data[GatewayCertID],
		Key:    data[GatewayKeyID],
	}
}
