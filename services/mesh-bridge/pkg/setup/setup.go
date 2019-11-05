package setup

import (
	"context"
	"os"
	"time"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/stats"
	"github.com/solo-io/mesh-projects/pkg/version"
	"go.uber.org/zap"
)

func Main(customCtx context.Context, errHandler func(error)) error {
	if os.Getenv("START_STATS_SERVER") != "" {
		stats.StartStatsServer()
	}

	writeNamespace := os.Getenv("POD_NAMESPACE")
	if writeNamespace == "" {
		writeNamespace = "sm-marketplace"
	}

	rootCtx := createRootContext(customCtx)

	if errHandler == nil {
		errHandler = func(err error) {
			if err == nil {
				return
			}
			contextutils.LoggerFrom(rootCtx).Errorf("error: %v", err)
		}
	}

	if err := runMeshBridgeEventLoop(rootCtx, writeNamespace, errHandler); err != nil {
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
	rootCtx = contextutils.WithLogger(rootCtx, "mesh-bridge")
	loggingContext := []interface{}{"version", version.Version}
	rootCtx = contextutils.WithLoggerValues(rootCtx, loggingContext...)
	return rootCtx
}

func runMeshBridgeEventLoop(ctx context.Context, writeNamespace string, errHandler func(error)) error {
	loops, err := MustInitializeMeshBridge(ctx)
	if err != nil {
		contextutils.LoggerFrom(ctx).Fatalw("Failed to initialize mesh bridge event loops", zap.Error(err))
	}

	errs, err := loops.MultiClusterRunFunc()
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

	time.Sleep(time.Second * 5)
	// run mesh bridge loop
	contextutils.LoggerFrom(ctx).Infow("Starting event loop",
		zap.Any("watchOpts", loops.WatchOpts),
		zap.Strings("watchNamespaces", loops.WatchNamespaces))
	errs, err = loops.OperatorLoop.Run(loops.WatchNamespaces, loops.WatchOpts)
	if err != nil {
		contextutils.LoggerFrom(ctx).Fatalw("Failed to start event loop", zap.Error(err))
	}
	for err := range errs {
		contextutils.LoggerFrom(ctx).Errorw("Error during event loop", zap.Error(err))
	}
	return nil
}
