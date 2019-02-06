package istio

import (
	"context"

	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/secret"
)

type EncryptionSyncer struct {
	SecretSyncer secret.SecretSyncer
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
	return s.SecretSyncer.SyncSecret(ctx, mesh.GetIstio().InstallationNamespace, mesh.Encryption, secretList, false)
}
