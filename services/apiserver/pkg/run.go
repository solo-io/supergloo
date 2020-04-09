package apiserver

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/service-mesh-hub/services/apiserver/pkg/wire"
	"github.com/solo-io/service-mesh-hub/services/common/multicluster"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/multicluster/manager"
	"github.com/solo-io/service-mesh-hub/services/internal/config"
	"go.uber.org/zap"
)

func Run(rootCtx context.Context) {
	ctx := config.CreateRootContext(rootCtx, "apiserver")

	logger := contextutils.LoggerFrom(ctx)

	apiserverContext, err := wire.InitializeApiServer(ctx)
	if err != nil {
		logger.Fatalw("error initializing api server context", zap.Error(err))
	}
	localManager := apiserverContext.MultiClusterDeps.LocalManager


	// block until we die; RIP
	if err := multicluster.SetupAndStartLocalManager(
		apiserverContext.MultiClusterDeps,
		// need to be sure to register the v1alpha1 CRDs with the controller runtime
		[]mc_manager.AsyncManagerStartOptionsFunc{multicluster.AddAllV1Alpha1ToScheme},
		[]multicluster.NamedAsyncManagerHandler{},
	); err != nil {
		logger.Fatalw("the local manager instance failed to start up or died with an error", zap.Error(err))
	}
}
