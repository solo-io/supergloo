package consul

import (
	"context"
	"fmt"

	"github.com/hashicorp/consul/api"

	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/supergloo/pkg/api/v1"

	istio "github.com/solo-io/supergloo/pkg/api/external/istio/encryption/v1"
)

type ConsulSyncer struct {
	LocalPort int
}

func (c *ConsulSyncer) Sync(_ context.Context, snap *v1.TranslatorSnapshot) error {
	for _, mesh := range snap.Meshes.List() {
		_, ok := mesh.MeshType.(*v1.Mesh_Consul)
		if !ok {
			// not our mesh, we don't care
			continue
		}
		encryption := mesh.Encryption
		if encryption == nil {
			continue
		}
		encryptionSecret := encryption.Secret
		if encryptionSecret == nil {
			continue
		}
		secret, err := snap.Istiocerts.List().Find(encryptionSecret.Namespace, encryptionSecret.Name)
		if err != nil {
			return err
		}

		port := c.LocalPort
		if port <= 0 {
			port = 8500
		}
		if err := syncSecret(secret, port); err != nil {
			return err
		}
	}
	return nil
}

func validateTlsSecret(secret *istio.IstioCacertsSecret) error {
	if secret.CaCert == "" {
		return errors.Errorf("Ca cert is missing.")
	}
	if secret.CaKey == "" {
		return errors.Errorf("Private key is missing.")
	}
	if secret.RootCert != "" && secret.RootCert != secret.CaCert {
		return errors.Errorf("If root cert is provided it must equal ca cert currently")
	}
	// TODO: This should be supported
	if secret.CertChain != "" {
		return errors.Errorf("Updating the root with a cert chain is not supported")
	}
	return nil
}

func getConsulInnerConfigMap(secret *istio.IstioCacertsSecret) map[string]interface{} {
	innerConfig := make(map[string]interface{})
	innerConfig["LeafCertTTL"] = "72h"
	innerConfig["PrivateKey"] = secret.CaCert
	innerConfig["RootCert"] = secret.RootCert
	innerConfig["RotationPeriod"] = "2160h"
	return innerConfig
}

func getConsulConfigMap(secret *istio.IstioCacertsSecret) *api.CAConfig {
	return &api.CAConfig{
		Provider: "consul",
		Config:   getConsulInnerConfigMap(secret),
	}
}

func shouldUpdateCurrentCert(client *api.Client, secret *istio.IstioCacertsSecret) (bool, error) {
	var queryOpts api.QueryOptions
	currentConfig, _, err := client.Connect().CAGetConfig(&queryOpts)
	if err != nil {
		return false, errors.Errorf("Error getting current root certificate: %v", err)
	}
	currentRoot := currentConfig.Config["RootCert"]
	if currentRoot == secret.CaCert {
		// Root certificate already set
		return false, nil
	}
	return true, nil
}

func syncSecret(secret *istio.IstioCacertsSecret, port int) error {
	// TODO: This should be configured using the mesh location from the CRD
	// TODO: This requires port forwarding, ingress, or running inside the cluster
	consulCfg := &api.Config{
		Address: fmt.Sprintf("127.0.0.1:%d", port),
	}
	client, err := api.NewClient(consulCfg)
	if err != nil {
		return errors.Errorf("error creating consul client %v", err)
	}
	if err = validateTlsSecret(secret); err != nil {
		return err
	}
	shouldUpdate, err := shouldUpdateCurrentCert(client, secret)
	if err != nil {
		return err
	}
	if !shouldUpdate {
		return nil
	}

	conf := getConsulConfigMap(secret)

	// TODO: Even if this succeeds, Consul will still get into a bad state if this is an RSA cert
	// Need to verify the cert was generated with EC
	var writeOpts api.WriteOptions
	if _, err = client.Connect().CASetConfig(conf, &writeOpts); err != nil {
		return errors.Errorf("Error updating consul root certificate %v.")
	}
	return nil
}
