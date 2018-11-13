package consul

import (
	"context"

	"github.com/solo-io/supergloo/pkg/api/v1"
)

type EncryptionSyncer struct {
}

func (s *EncryptionSyncer) Sync(ctx context.Context, snap *v1.TranslatorSnapshot) error {
	return nil
}
