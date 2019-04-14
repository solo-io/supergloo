package config

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/pkg/clientset"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/pkg/config/common"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/pkg/config/istio"
)

func RunAdvancedDiscoverySyncers(ctx context.Context, cs *clientset.Clientset, enabled common.EnabledConfigLoops) error {
	ctx = contextutils.WithLogger(ctx, "mesh-discovery-config-event-loop")
	// logger := contextutils.LoggerFrom(ctx)

	configSyncers, err := seutpAdvancedDiscoveryPlugins(ctx, cs, enabled)
	if err != nil {
		return err
	}

	if err := runConfigEventLoop(ctx, configSyncers); err != nil {
		return err
	}

	return nil
}

func seutpAdvancedDiscoveryPlugins(ctx context.Context, cs *clientset.Clientset, enabled common.EnabledConfigLoops) (common.AdvancedDiscoverySycnerList, error) {
	plugins := common.AdvancedDiscoverySycnerList{}
	if enabled.Istio {
		istioPlugin, err := istio.NewIstioAdvancedDiscoveryPlugin(ctx, cs)
		if err != nil {
			return nil, err
		}
		plugins = append(plugins, istioPlugin)
	}
	return plugins, nil
}

// start the istio config event loop
func runConfigEventLoop(ctx context.Context, pluginList common.AdvancedDiscoverySycnerList) error {

	for _, plugin := range pluginList {
		eventLoopErrs, err := plugin.Run()
		if err != nil {
			return err
		}
		go func() {
			for {
				select {
				case err := <-eventLoopErrs:
					plugin.HandleError(err)
				case <-ctx.Done():
				}
			}
		}()
	}
	return nil
}
