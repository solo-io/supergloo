package istio

import (
	"context"

	"github.com/solo-io/supergloo/pkg/secret"

	"k8s.io/client-go/kubernetes"

	istiov1 "github.com/solo-io/supergloo/pkg/api/external/istio/encryption/v1"

	"github.com/solo-io/supergloo/pkg/api/v1"
)

type EncryptionSyncer struct {
	IstioNamespace string
	Kube           kubernetes.Interface
	SecretClient   istiov1.IstioCacertsSecretClient
}

func (s *EncryptionSyncer) Sync(ctx context.Context, snap *v1.TranslatorSnapshot) error {
	for _, mesh := range snap.Meshes.List() {
		if err := s.syncMesh(ctx, mesh, snap); err != nil {
			return err
		}
	}
	return nil
}

func (s *EncryptionSyncer) syncMesh(ctx context.Context, mesh *v1.Mesh, snap *v1.TranslatorSnapshot) error {
	if mesh.GetIstio() == nil {
		return nil
	}
	secretList := snap.Istiocerts.List()
	secretSyncer := secret.SecretSyncer{
		Kube:         s.Kube,
		SecretClient: s.SecretClient,
		SecretList:   secretList,
		Preinstall:   false,
	}
	return secretSyncer.SyncSecret(ctx, mesh.GetIstio().InstallationNamespace, mesh.Encryption)
}
