package istio

import (
	"context"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"

	"k8s.io/client-go/kubernetes"

	gloov1 "github.com/solo-io/supergloo/pkg/api/external/gloo/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/translator/kube"
)

type EncryptionSyncer struct {
	Kube         kubernetes.Interface
	SecretClient gloov1.SecretClient
	ctx          context.Context
}

const (
	defaultIstioNamespace            = "istio-system"
	customRootCertificateSecretName  = "cacerts"
	defaultRootCertificateSecretName = "istio.default"
	istioLabelKey                    = "istio"
	citadelLabelValue                = "citadel"
)

func (s *EncryptionSyncer) Sync(ctx context.Context, snap *v1.TranslatorSnapshot) error {
	s.ctx = ctx
	for _, mesh := range snap.Meshes.List() {
		if err := s.syncMesh(mesh, snap); err != nil {
			return err
		}
	}
	return nil
}

func (s *EncryptionSyncer) syncMesh(mesh *v1.Mesh, snap *v1.TranslatorSnapshot) error {
	_, ok := mesh.MeshType.(*v1.Mesh_Istio)
	if !ok {
		// not our mesh, we don't care
		return nil
	}
	encryption := mesh.Encryption
	if encryption == nil {
		return nil
	}
	encryptionSecret := encryption.Secret
	if encryptionSecret == nil {
		return nil
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

	// this is where custom root certs will live once configured, if not found istioCacerts will be nil
	istioCacerts, _ := secretList.Find(defaultIstioNamespace, customRootCertificateSecretName)

	return s.syncSecret(tlsSecretFromMeshConfig, istioCacerts)
}

func (s *EncryptionSyncer) syncSecret(tlsSecretFromMeshConfig *gloov1.TlsSecret, currentCacerts *gloov1.Secret) error {
	if err := validateTlsSecret(tlsSecretFromMeshConfig); err != nil {
		return err
	}

	cacertsSecret := convertToCacerts(tlsSecretFromMeshConfig)
	if !cacertsSecret.Equal(currentCacerts) {
		if err := s.writeCacerts(cacertsSecret); err != nil {
			return err
		}
		// now we need to ensure istio changes to use this cert:
		// make sure istio.default is deleted, and restart citadel
		if err := s.deleteIstioDefaultSecret(); err != nil {
			return err
		}
		return s.restartCitadel()
	} else {
		return nil
	}
}

func (s *EncryptionSyncer) writeCacerts(secret *gloov1.Secret) error {
	_, err := s.SecretClient.Write(secret, clients.WriteOpts{})
	return err
}

func validateTlsSecret(secret *gloov1.TlsSecret) error {
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

func convertToCacerts(tlsSecretFromMeshConfig *gloov1.TlsSecret) *gloov1.Secret {
	cacerts := gloov1.IstioCacertsSecret{
		CaKey:     tlsSecretFromMeshConfig.PrivateKey,
		CaCert:    tlsSecretFromMeshConfig.RootCa,
		CertChain: tlsSecretFromMeshConfig.CertChain,
		RootCert:  tlsSecretFromMeshConfig.RootCa,
	}
	cacertsWrapper := gloov1.Secret_Cacerts{
		Cacerts: &cacerts,
	}
	secret := gloov1.Secret{
		Kind: &cacertsWrapper,
		Metadata: core.Metadata{
			Namespace: defaultIstioNamespace,
			Name:      customRootCertificateSecretName,
		},
	}
	return &secret
}

func (s *EncryptionSyncer) deleteIstioDefaultSecret() error {
	// Using Kube API directly cause we don't expect this secret to be tagged and it should be mostly a one-time op
	return s.Kube.CoreV1().Secrets(defaultIstioNamespace).Delete(defaultRootCertificateSecretName, &metav1.DeleteOptions{})
}

func (s *EncryptionSyncer) restartCitadel() error {
	selector := make(map[string]string)
	selector[istioLabelKey] = citadelLabelValue
	return kube.RestartPods(s.Kube, defaultIstioNamespace, selector)
}
