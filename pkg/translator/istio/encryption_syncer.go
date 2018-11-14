package istio

import (
	"context"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"

	"github.com/hashicorp/consul/api"
	v12 "github.com/solo-io/supergloo/pkg/api/external/gloo/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/supergloo/pkg/api/v1"
)

type IstioSyncer struct {
	Kube kubernetes.Interface
}

func Sync(_ context.Context, snap *v1.TranslatorSnapshot) error {
	for _, mesh := range snap.Meshes.List() {
		switch mesh.TargetMesh.MeshType {
		case v1.MeshType_ISTIO:
			encryption := mesh.Encryption
			if encryption == nil {
				continue
			}
			encryptionSecret := encryption.Secret
			if encryptionSecret == nil {
				continue
			}
			secretList := snap.Secrets.List()
			secretInMeshConfig, err := secretList.Find(encryptionSecret.Namespace, encryptionSecret.Name)
			if err != nil {
				return errors.Errorf("Error finding secret referenced in mesh config (%s:%s): %v",
					encryptionSecret.Namespace, encryptionSecret.Name, err)
			}
			tlsSecretFromMeshConfig := secretInMeshConfig.GetTls()
			if tlsSecretFromMeshConfig == nil {
				return errors.Errorf("missing tls secret")
			}

			// this is where custom root certs will live
			istioCacerts, _ := secretList.Find("istio-system", "cacerts")
			istioCacerts.GetOpaque()

			syncSecret(tlsSecret)
		}
	}
	return nil
}

func (s *IstioSyncer) restartCitadel(ctx context.Context) error {
	selector := make(map[string]string)
	selector["istio"] = "citadel"
	return s.restartPods(ctx, "istio-system", selector)
}

func (s *IstioSyncer) restartPods(ctx context.Context, namespace string, selector map[string]string) error {
	if s.Kube == nil {
		return errors.Errorf("kubernetes suppport is currently disabled. see SuperGloo documentation" +
			" for utilizing pod restarts")
	}
	if err := s.Kube.CoreV1().Pods(namespace).DeleteCollection(nil, metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(selector).String(),
	}); err != nil {
		return errors.Wrapf(err, "restarting pods with selector %v", selector)
	}
	return nil
}

func validateTlsSecret(secret *v12.TlsSecret) error {
	if secret.RootCa == "" {
		return errors.Errorf("Root cert is missing.")
	}
	if secret.PrivateKey == "" {
		return errors.Errorf("Private key is missing.")
	}
	// TODO: This should be supported
	if secret.CertChain != "" {
		return errors.Errorf("Updating the root with a cert chain is not supported")
	}
	return nil
}

func (s *IstioSyncer) shouldUpdateCurrentCert(secret *v12.TlsSecret) (bool, error) {
	s.Kube.CoreV1().Secrets("istio-system").Get()
	var queryOpts api.QueryOptions
	currentConfig, _, err := client.Connect().CAGetConfig(&queryOpts)
	if err != nil {
		return false, errors.Errorf("Error getting current root certificate: %v", err)
	}
	currentRoot := currentConfig.Config["RootCert"]
	if currentRoot == secret.RootCa {
		// Root certificate already set
		return false, nil
	}
	return true, nil
}

func syncSecret(tlsSecretFromMeshConfig *v12.TlsSecret, currentCacerts *v12.Secret) error {
	// TODO: This should be configured using the mesh location from the CRD
	// TODO: This requires port forwarding, ingress, or running inside the cluster
	client, err := api.NewClient(api.DefaultConfig())
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
