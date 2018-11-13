package consul

import (
	"context"

	"github.com/hashicorp/consul/api"
	v12 "github.com/solo-io/supergloo/pkg/api/external/gloo/v1"

	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/supergloo/pkg/api/v1"
)

type EncryptionSyncer struct {
}

func (s *EncryptionSyncer) Sync(ctx context.Context, snap *v1.TranslatorSnapshot) error {
	for _, mesh := range snap.Meshes.List() {
		switch mesh.TargetMesh.MeshType {
		case v1.MeshType_CONSUL:
			encryption := mesh.Encryption
			if encryption == nil {
				return nil
			}
			encryptionSecret := encryption.Secret
			if encryptionSecret == nil {
				return nil
			}
			secret, err := snap.Secrets.List().Find(encryptionSecret.Namespace, encryptionSecret.Name)
			if err != nil {
				return err
			}
			tlsSecret := secret.GetTls()
			if tlsSecret == nil {
				errors.Errorf("missing tls secret")
				return nil
			}

			s.sync(ctx, tlsSecret)
		}
	}
	return nil
}

func (s *EncryptionSyncer) sync(ctx context.Context, secret *v12.TlsSecret) error {
	// TODO: This should be configured using the mesh location from the CRD
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		errors.Errorf("error creating consul client %v", err)
		return err
	}

	if secret.RootCa == "" {
		errors.Errorf("Root cert is missing.")
		return nil
	}
	if secret.PrivateKey == "" {
		errors.Errorf("Private key is missing.")
		return nil
	}
	// TODO: This should be supported
	if secret.CertChain != "" {
		errors.Errorf("Updating the root with a cert chain is not supported")
		return nil
	}

	innerConfig := make(map[string]interface{})
	innerConfig["LeafCertTTL"] = "72h"
	innerConfig["PrivateKey"] = secret.PrivateKey
	innerConfig["RootCert"] = secret.RootCa
	innerConfig["RotationPeriod"] = "2160h"

	conf := &api.CAConfig{
		Provider: "consul",
		Config:   innerConfig,
	}

	// TODO: Even if this succeeds, Consul will still get into a bad state if this is an RSA cert
	// Need to verify the cert was generated with EC
	var writeOpts api.WriteOptions
	_, err = client.Connect().CASetConfig(conf, &writeOpts)
	if err != nil {
		errors.Errorf("Error updating consul root certificate %v.")
		return err
	}
	return nil
}
