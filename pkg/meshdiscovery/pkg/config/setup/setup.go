package setup

import (
	"context"
	"time"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/pkg/clientset"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/pkg/config"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/pkg/config/istio"
	"github.com/solo-io/supergloo/pkg/registration"
	"go.uber.org/zap"
)

type SuperglooCongigLoop struct {
	Clientset  *clientset.Clientset
	ErrHandler func(error)
}

func NewDiscoveryConfigLoop(clientset *clientset.Clientset, errHandler func(error)) *SuperglooCongigLoop {
	return &SuperglooCongigLoop{Clientset: clientset, ErrHandler: errHandler}
}

func (s *SuperglooCongigLoop) Run(ctx context.Context, enabled registration.EnabledConfigLoops) error {
	ctx = contextutils.WithLogger(ctx, "mesh-config-discovery")

	plugins, err := createConfigSyncers(ctx, s.Clientset, enabled)
	if err != nil {
		return err
	}

	if err := runConfigEventLoop(ctx, plugins); err != nil {
		return err
	}

	return nil
}

// Add config syncers here
func createConfigSyncers(ctx context.Context, cs *clientset.Clientset, enabled registration.EnabledConfigLoops) ([]config.EventLoop, error) {
	var syncers []config.EventLoop

	if enabled.Istio {
		istioPlugin, err := istio.NewIstioConfigDiscovery(ctx, cs)
		if err != nil {
			return nil, err
		}
		syncers = append(syncers, istioPlugin)
	}

	return syncers, nil
}

// start the istio config event loop
func runConfigEventLoop(ctx context.Context, plugins []config.EventLoop) error {
	watchOpts := clients.WatchOpts{
		Ctx:         ctx,
		RefreshRate: time.Minute * 1,
	}
	logger := contextutils.LoggerFrom(ctx)
	for _, plugin := range plugins {
		plugin := plugin
		combinedErrs, err := plugin.Run(nil, watchOpts)
		if err != nil {
			return err
		}
		go func() {
			for {
				select {
				case err := <-combinedErrs:
					if err != nil {
						logger.With(zap.Error(err)).Info("discovery config failure")
					}
				case <-ctx.Done():
				}
			}
		}()
	}

	return nil
}
