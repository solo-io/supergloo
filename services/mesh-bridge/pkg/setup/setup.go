package setup

import (
	"context"
	"os"
	"time"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/stats"
	"github.com/solo-io/mesh-projects/pkg/api/external/istio/networking/v1alpha3"
	"github.com/solo-io/mesh-projects/services/internal/config"
	"github.com/solo-io/mesh-projects/services/internal/kube"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func Main(customCtx context.Context, errHandler func(error)) error {
	stats.ConditionallyStartStatsServer()

	writeNamespace := os.Getenv("POD_NAMESPACE")
	if writeNamespace == "" {
		writeNamespace = "sm-marketplace"
	}

	if errHandler == nil {
		errHandler = func(err error) {
			if err == nil {
				return
			}
			contextutils.LoggerFrom(customCtx).Errorf("error: %v", err)
		}
	}

	if err := runMeshBridgeEventLoop(customCtx, writeNamespace, errHandler); err != nil {
		return err
	}

	<-customCtx.Done()
	return nil

}

func runMeshBridgeEventLoop(ctx context.Context, writeNamespace string, errHandler func(error)) error {
	logger := contextutils.LoggerFrom(ctx)
	logger.Infow("waiting for crds to be registered before starting up",
		zap.Any("service entry", v1alpha3.ServiceEntryCrd))
	// Do not start running until service entry crd has been registered
	ticker := time.Tick(2 * time.Second)
	cfg := kube.MustGetKubeConfig(ctx)
	eg := &errgroup.Group{}
	eg.Go(func() error {
		for {
			select {
			case <-ticker:
				if config.CrdsExist(cfg, v1alpha3.ServiceEntryCrd) {
					return nil
				}
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	})

	if err := eg.Wait(); err != nil {
		return err
	}

	loops, err := MustInitializeMeshBridge(ctx)
	if err != nil {
		logger.Fatalw("Failed to initialize mesh bridge event loops", zap.Error(err))
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
	logger.Infow("Starting event loop",
		zap.Any("watchOpts", loops.WatchOpts),
		zap.Strings("watchNamespaces", loops.WatchNamespaces))
	errs, err = loops.OperatorLoop.Run(loops.WatchNamespaces, loops.WatchOpts)
	if err != nil {
		logger.Fatalw("Failed to start event loop", zap.Error(err))
	}
	for err := range errs {
		logger.Errorw("Error during event loop", zap.Error(err))
	}
	return nil
}
