package setup

import (
	"context"
	"os"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients/wrapper"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/stats"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/clientset"
)

// customCtx and customErrHandler are expected to be passed by tests
func Main(customCtx context.Context, errHandler func(error)) error {
	if os.Getenv("START_STATS_SERVER") != "" {
		stats.StartStatsServer()
	}

	rootCtx := createRootContext(customCtx)

	clientSet, err := clientset.ClientsetFromContext(rootCtx)
	if err != nil {
		return err
	}

	if err := runDiscoveryEventLoop(rootCtx, clientSet, errHandler); err != nil {
		return err
	}

	<-rootCtx.Done()
	return nil
}

func createRootContext(customCtx context.Context) context.Context {
	rootCtx := customCtx
	if rootCtx == nil {
		rootCtx = context.Background()
	}
	rootCtx = contextutils.WithLogger(rootCtx, "meshdiscovery")
	return rootCtx
}

func runDiscoveryEventLoop(ctx context.Context, cs *clientset.Clientset, errHandler func(error)) error {

	emitter := v1.NewDiscoverySimpleEmitter(wrapper.AggregatedWatchFromClients(
		wrapper.ClientWatchOpts{BaseClient: cs.Input.Upstream.BaseClient()},
		wrapper.ClientWatchOpts{BaseClient: cs.Input.Pod.BaseClient()},
		wrapper.ClientWatchOpts{BaseClient: cs.Input.ConfigMap.BaseClient()},
	))
	eventLoop := v1.NewDiscoverySimpleEventLoop(emitter)

	errs, err := eventLoop.Run(ctx)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case err := <-errs:
				errHandler(err)
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}
