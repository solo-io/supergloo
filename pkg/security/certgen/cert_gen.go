package certgen

import (
	"time"

	networking_types "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	cert_secrets "github.com/solo-io/mesh-projects/pkg/security/secrets"
	"istio.io/istio/security/pkg/pki/util"
)

const (
	DefaultRootCertTTLDays     = 365
	DefaultRootCertTTLDuration = DefaultRootCertTTLDays * 24 * time.Hour
	DefaultRootCertRsaKeySize  = 4096
	DefaultOrgName             = "service-mesh-hub"
)

type rootCertGenerator struct{}

func NewRootCertGenerator() RootCertGenerator {
	return &rootCertGenerator{}
}

// Generate a new private key and use it to self-sign a new root certificate
func (c *rootCertGenerator) GenRootCertAndKey(
	builtinCA *networking_types.VirtualMeshSpec_CertificateAuthority_Builtin,
) (*cert_secrets.RootCAData, error) {
	org := DefaultOrgName
	if builtinCA.GetOrgName() != "" {
		org = builtinCA.GetOrgName()
	}
	ttl := DefaultRootCertTTLDuration
	if builtinCA.GetTtlDays() > 0 {
		ttl = time.Duration(builtinCA.GetTtlDays()) * 24 * time.Hour
	}
	rsaKeySize := DefaultRootCertRsaKeySize
	if builtinCA.GetRsaKeySizeBytes() > 0 {
		rsaKeySize = int(builtinCA.GetRsaKeySizeBytes())
	}
	options := util.CertOptions{
		Org:          org,
		IsCA:         true,
		IsSelfSigned: true,
		TTL:          ttl,
		RSAKeySize:   rsaKeySize,
		PKCS8Key:     false, // currently only supporting PKCS1
	}
	cert, key, err := util.GenCertKeyFromOptions(options)
	if err != nil {
		return nil, err
	}
	rootCaData := &cert_secrets.RootCAData{
		PrivateKey: key,
		RootCert:   cert,
	}
	return rootCaData, nil
}
