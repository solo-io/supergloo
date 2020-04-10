package apiserver

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/service-mesh-hub/services/apiserver/pkg/wire"
	"github.com/solo-io/service-mesh-hub/services/common/multicluster"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/multicluster/manager"
	"github.com/solo-io/service-mesh-hub/services/internal/config"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func Run(rootCtx context.Context) {
	ctx := config.CreateRootContext(rootCtx, "apiserver")

	logger := contextutils.LoggerFrom(ctx)

	apiserverContext, err := wire.InitializeApiServer(ctx)
	if err != nil {
		logger.Fatalw("error initializing api server context", zap.Error(err))
	}

	var eg *errgroup.Group
	eg, ctx = errgroup.WithContext(ctx)

	eg.Go(func() error {
		return apiserverContext.Server.Run()
	})

	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		select {
		case <-ctx.Done():
			apiserverContext.Server.Stop()
		case <-sigs:
			apiserverContext.Server.Stop()
		}
	}()

	eg.Go(func() error {
		return multicluster.SetupAndStartLocalManager(
			apiserverContext.MultiClusterDeps,
			// need to be sure to register the v1alpha1 CRDs with the controller runtime
			[]mc_manager.AsyncManagerStartOptionsFunc{multicluster.AddAllV1Alpha1ToScheme},
			[]multicluster.NamedAsyncManagerHandler{},
		)
	})

	// block until we die; RIP
	if err := eg.Wait(); err != nil {
		logger.Fatalw("The app has crashed", zap.Error(err))
	}
}
