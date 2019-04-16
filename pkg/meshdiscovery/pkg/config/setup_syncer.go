package config

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/pkg/clientset"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/pkg/config/common"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/pkg/config/istio"
)

func NewAdvancedDiscoverySyncers(ctx context.Context, cs *clientset.Clientset) chan<- *common.EnabledConfigLoops {
	ctx = contextutils.WithLogger(ctx, "mesh-discovery-config-event-loop")

	configSyncers := seutpAdvancedDiscoveryPlugins(ctx, cs)
	diffChan := make(chan *common.EnabledConfigLoops)

	runConfigEventLoop(ctx, configSyncers, diffChan)

	return diffChan
}

func seutpAdvancedDiscoveryPlugins(ctx context.Context, cs *clientset.Clientset) common.AdvancedDiscoverySycnerList {
	plugins := common.AdvancedDiscoverySycnerList{
		istio.NewIstioAdvancedDiscoveryPlugin(ctx, cs),
	}
	return plugins
}

// start the istio config event loop
func runConfigEventLoop(ctx context.Context, pluginList common.AdvancedDiscoverySycnerList, diffChan chan *common.EnabledConfigLoops) {
	pluginDiffs := make([]chan<- *common.EnabledConfigLoops, len(pluginList))
	for _, plugin := range pluginList {
		pluginDiffs = append(pluginDiffs, plugin.Run(ctx))
	}

	go func() {
		for {
			select {
			case diff, ok := <-diffChan:
				if !ok {
					return
				}
				for _, pluginDiffChan := range pluginDiffs {
					pluginDiffChan <- diff
				}
			case <-ctx.Done():
			}
		}
	}()
}
