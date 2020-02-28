package csr_agent_controller

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"

	"github.com/rotisserie/eris"
	securityv1alpha1 "github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1"
	kubernetes_core "github.com/solo-io/mesh-projects/pkg/clients/kubernetes/core"
	"github.com/solo-io/mesh-projects/pkg/security/certgen"
	cert_secrets "github.com/solo-io/mesh-projects/pkg/security/secrets"
	kubeerrs "k8s.io/apimachinery/pkg/api/errors"
)

type certClient struct {
	secretClient kubernetes_core.SecretsClient
	signer       certgen.Signer
}

func NewCertClient(
	secretClient kubernetes_core.SecretsClient,
	signer certgen.Signer,
) CertClient {
	return &certClient{
		secretClient: secretClient,
		signer:       signer,
	}
}

func (c *certClient) EnsureSecretKey(ctx context.Context, obj *securityv1alpha1.MeshGroupCertificateSigningRequest) (*cert_secrets.RootCaData, error) {
	secret, err := c.secretClient.Get(ctx, obj.GetName(), obj.GetNamespace())
	if err != nil {
		if !kubeerrs.IsNotFound(err) {
			return nil, err
		}

		// use large as this is the CA key
		priv, err := rsa.GenerateKey(rand.Reader, 4096)
		if err != nil {
			return nil, eris.Errorf("RSA key generation failed (%v)", err)
		}

		privKey := x509.MarshalPKCS1PrivateKey(priv)
		keyBlock := &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privKey,
		}
		byt := pem.EncodeToMemory(keyBlock)
		certData := &cert_secrets.RootCaData{
			CaPrivateKey: byt,
		}
		newSecret := certData.BuildSecret(obj.GetName(), obj.GetNamespace())
		if err = c.secretClient.Create(ctx, newSecret); err != nil {
			return nil, err
		}
		return certData, nil

	}
	return cert_secrets.RootCaDataFromSecret(secret)
}
