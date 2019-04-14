package istio

import (
	"context"
	"fmt"

	"github.com/solo-io/go-utils/contextutils"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"go.uber.org/zap"
)

type istioConfigSyncer struct {
}

func newIstioConfigSyncer() *istioConfigSyncer {
	return &istioConfigSyncer{}
}

func (s *istioConfigSyncer) Sync(ctx context.Context, snap *v1.IstioDiscoverySnapshot) error {
	ctx = contextutils.WithLogger(ctx, fmt.Sprintf("istio-translation-sync-%v", snap.Hash()))
	logger := contextutils.LoggerFrom(ctx)
	fields := []interface{}{
		zap.Int("pods", len(snap.Pods.List())),
		zap.Int("meshes", len(snap.Meshes.List())),
		zap.Int("installs", len(snap.Installs.List())),
		zap.Int("meshpolicies", len(snap.Meshpolicies)),
	}
	logger.Infow("begin sync", fields...)
	defer logger.Infow("end sync", fields...)
	logger.Debugf("full snapshot: %v", snap)

	logger.Infof("sync completed successfully!")
	return nil
}
