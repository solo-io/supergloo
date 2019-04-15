package istio

import (
	"context"
	"time"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/pkg/clientset"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/pkg/config/common"
)

type IstioAdvancedDiscoveryPlugin struct {
	ctx context.Context
	cs  *clientset.Clientset

	el v1.IstioDiscoveryEventLoop
}

func NewIstioAdvancedDiscoveryPlugin(ctx context.Context, cs *clientset.Clientset) (*IstioAdvancedDiscoveryPlugin, error) {
	ctx = contextutils.WithLogger(ctx, "istio-advanced-discovery-config-event-loop")

	istioClients, err := clientset.IstioFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "initializing istio clients")
	}

	emitter := v1.NewIstioDiscoveryEmitter(
		cs.Input.Pod,
		cs.Discovery.Mesh,
		cs.Input.Install,
		istioClients.MeshPolicies,
	)

	syncer := newIstioConfigSyncer(ctx, cs, istioClients)

	el := v1.NewIstioDiscoveryEventLoop(emitter, syncer)

	return &IstioAdvancedDiscoveryPlugin{ctx: ctx, cs: cs, el: el}, nil
}

func (iasd *IstioAdvancedDiscoveryPlugin) Run() (<-chan error, error) {
	watchOpts := clients.WatchOpts{
		Ctx:         iasd.ctx,
		RefreshRate: time.Second * 1,
	}
	return iasd.el.Run(nil, watchOpts)
}

func (iasd *IstioAdvancedDiscoveryPlugin) HandleError(err error) {
	if err != nil {
		contextutils.LoggerFrom(iasd.ctx).Errorf("config error: %v", err)
	}
}

func (iasd *IstioAdvancedDiscoveryPlugin) Enabled(enabled *common.EnabledConfigLoops) bool {
	return enabled.Istio()
}
