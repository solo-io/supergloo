package istio

import (
	"context"
	"time"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/pkg/clientset"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/pkg/config/common"
	"go.uber.org/zap"
)

type IstioAdvancedDiscoveryPlugin struct {
	rootCtx context.Context
	cs      *clientset.Clientset
}

func NewIstioAdvancedDiscoveryPlugin(ctx context.Context, cs *clientset.Clientset) *IstioAdvancedDiscoveryPlugin {
	ctx = contextutils.WithLogger(ctx, "istio-advanced-discovery-config-event-loop")
	return &IstioAdvancedDiscoveryPlugin{rootCtx: ctx, cs: cs}
}

func (iasd *IstioAdvancedDiscoveryPlugin) Run(ctx context.Context) chan<- *common.EnabledConfigLoops {
	diffChan := make(chan *common.EnabledConfigLoops)
	logger := contextutils.LoggerFrom(ctx)
	go func() {
		go func() {
			// create a new context for each loop, cancel it before each loop
			var cancel context.CancelFunc = func() {}
			previousDiff := &common.EnabledConfigLoops{}
			// use closure to allow cancel function to be updated as context changes
			defer func() { cancel() }()
			for {
				select {
				case diff, ok := <-diffChan:
					if !ok {
						return
					}
					subLogger := logger.With(zap.Any("previous", previousDiff), zap.Any("current", diff))
					subLogger.Infof("checking if diff has changed")
					if diff.Istio() == previousDiff.Istio() {
						subLogger.Infof("states are consistent")
						continue
					}
					// cancel any open watches from previous diff
					cancel()

					ctx, canc := context.WithCancel(ctx)
					cancel = canc

					if diff.Istio() {
						subLogger.Infof("istio state has changed to true")
						if err := iasd.startEventLoop(ctx); err != nil {
							iasd.handleError(ctx, err)
							return
						}
					}

					// set previous diff to current
					previousDiff = diff
				case <-ctx.Done():
					return
				}
			}
		}()
	}()

	return diffChan
}

func (iasd *IstioAdvancedDiscoveryPlugin) startEventLoop(ctx context.Context) error {

	emitter := v1.NewIstioDiscoveryEmitter(
		iasd.cs.Input.Pod,
		iasd.cs.Discovery.Mesh,
		iasd.cs.Input.Install,
	)

	syncer := newIstioConfigSyncer(ctx, iasd.cs)
	watchOpts := clients.WatchOpts{
		Ctx:         ctx,
		RefreshRate: time.Minute * 1,
	}
	el := v1.NewIstioDiscoveryEventLoop(emitter, syncer)
	eventLoopErrs, err := el.Run(nil, watchOpts)
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case err := <-eventLoopErrs:
				iasd.handleError(ctx, err)
			case <-ctx.Done():
			}
		}
	}()
	return nil
}

func (iasd *IstioAdvancedDiscoveryPlugin) handleError(ctx context.Context, err error) {
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorf("config error: %v", err)
	}
}
