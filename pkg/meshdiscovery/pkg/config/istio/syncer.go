package istio

import (
	"context"
	"fmt"
	"time"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/pkg/clientset"
	"go.uber.org/zap"
)

type advancedIstioDiscovery struct {
	el v1.IstioDiscoveryEventLoop
}

func NewIstioAdvancedDiscovery(ctx context.Context, cs *clientset.Clientset) (*advancedIstioDiscovery, error) {
	istioClient, err := clientset.IstioFromContext(ctx)
	if err != nil {
		return nil, err
	}

	emitter := v1.NewIstioDiscoveryEmitter(
		cs.Discovery.Mesh,
		cs.Input.Install,
		istioClient.MeshPolicies,
	)
	syncer := newAdvancedIstioDiscoverSyncer()
	el := v1.NewIstioDiscoveryEventLoop(emitter, syncer)

	return &advancedIstioDiscovery{el: el}, nil
}

func (s *advancedIstioDiscovery) Run(ctx context.Context) (<-chan error, error) {
	watchOpts := clients.WatchOpts{
		Ctx:         ctx,
		RefreshRate: time.Minute * 1,
	}
	return s.el.Run(nil, watchOpts)
}

func (s *advancedIstioDiscovery) HandleError(ctx context.Context, err error) {
	if err != nil {
		contextutils.LoggerFrom(ctx).With(zap.Error(err)).Info("advanced istio discovery failure")
	}
}

type advancedIstioDiscoverSyncer struct {
}

func newAdvancedIstioDiscoverSyncer() *advancedIstioDiscoverSyncer {
	return &advancedIstioDiscoverSyncer{}
}

func (s *advancedIstioDiscoverSyncer) Sync(ctx context.Context, snap *v1.IstioDiscoverySnapshot) error {
	ctx = contextutils.WithLogger(ctx, fmt.Sprintf("istio-advanced-discovery-sync-%v", snap.Hash()))
	logger := contextutils.LoggerFrom(ctx)
	fields := []interface{}{
		zap.Int("meshes", len(snap.Meshes.List())),
		zap.Int("installs", len(snap.Installs.List())),
		zap.Int("mesh-policies", len(snap.Meshpolicies)),
	}

	meshPolicies := snap.Meshpolicies
	for _, policy := range meshPolicies {
	}

	logger.Infow("begin sync", fields...)
	defer logger.Infow("end sync", fields...)
	logger.Debugf("full snapshot: %v", snap)

	logger.Infof("sync completed successfully!")
	return nil
}
