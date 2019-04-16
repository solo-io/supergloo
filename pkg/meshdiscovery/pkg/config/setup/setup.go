package setup

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/pkg/clientset"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/pkg/config"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/pkg/config/istio"
	"github.com/solo-io/supergloo/pkg/registration"
)

type SuperglooCongigLoop struct {
	Clientset  *clientset.Clientset
	ErrHandler func(error)
}

func NewDiscoveryConfigLoop(clientset *clientset.Clientset, errHandler func(error)) *SuperglooCongigLoop {
	return &SuperglooCongigLoop{Clientset: clientset, ErrHandler: errHandler}
}

func (s *SuperglooCongigLoop) Run(ctx context.Context, enabled registration.EnabledConfigLoops) error {
	ctx = contextutils.WithLogger(ctx, "advanced-mesh-discovery")

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
func createConfigSyncers(ctx context.Context, cs *clientset.Clientset, enabled registration.EnabledConfigLoops) ([]config.AdvancedMeshDiscovery, error) {
	var syncers []config.AdvancedMeshDiscovery

	if enabled.Istio {
		istioPlugin, err := istio.NewIstioAdvancedDiscovery(ctx, cs)
		if err != nil {
			return nil, err
		}
		syncers = append(syncers, istioPlugin)
	}

	return syncers, nil
}

// start the istio config event loop
func runConfigEventLoop(ctx context.Context, plugins []config.AdvancedMeshDiscovery) error {

	for _, plugin := range plugins {
		plugin := plugin
		combinedErrs, err := plugin.Run(ctx)s
		if err != nil {
			return err
		}
		go func() {
			for {
				select {
				case err := <-combinedErrs:
					plugin.HandleError(ctx, err)
				case <-ctx.Done():
				}
			}
		}()
	}

	return nil
}
