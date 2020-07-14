package csr_generator

import (
	"context"

	k8s_core "github.com/solo-io/external-apis/pkg/api/k8s/core/v1"
	smh_security "github.com/solo-io/service-mesh-hub/pkg/api/security.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/common/csr/certgen"
	cert_secrets "github.com/solo-io/service-mesh-hub/pkg/common/csr/certgen/secrets"
	k8s_errs "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	PrivateKeyNameSuffix = "-private-key"
	PrivateKeySizeBytes  = 4096
)

type certClient struct {
	secretClient        k8s_core.SecretClient
	signer              certgen.Signer
	privateKeyGenerator PrivateKeyGenerator
}

func NewCertClient(
	secretClient k8s_core.SecretClient,
	signer certgen.Signer,
	privateKeyGenerator PrivateKeyGenerator,
) CertClient {
	return &certClient{
		secretClient:        secretClient,
		signer:              signer,
		privateKeyGenerator: privateKeyGenerator,
	}
}

// Persist the intermediate cert's private key as a secret of type cert_secrets.IntermediateCertSecretType
func (c *certClient) EnsureSecretKey(
	ctx context.Context,
	obj *smh_security.VirtualMeshCertificateSigningRequest,
) (*cert_secrets.IntermediateCAData, error) {
	secret, err := c.secretClient.GetSecret(ctx, client.ObjectKey{Name: buildSecretName(obj), Namespace: obj.GetNamespace()})
	if err != nil {
		if !k8s_errs.IsNotFound(err) {
			return nil, err
		}
		privateKey, err := c.privateKeyGenerator.GenerateRSA(PrivateKeySizeBytes)
		if err != nil {
			return nil, err
		}
		certData := &cert_secrets.IntermediateCAData{
			CaPrivateKey: privateKey,
		}
		newSecret := certData.BuildSecret(buildSecretName(obj), obj.GetNamespace())
		if err = c.secretClient.CreateSecret(ctx, newSecret); err != nil {
			return nil, err
		}
		return certData, nil

	}
	return cert_secrets.IntermediateCADataFromSecret(secret)
}

// suffix the name of the CSR with "-private-key" to avoid confusion, since we're reusing the
// cert_secrets.IntermediateCertSecretType secret type
func buildSecretName(obj *smh_security.VirtualMeshCertificateSigningRequest) string {
	return obj.GetName() + PrivateKeyNameSuffix
}
