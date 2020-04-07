package csr_agent

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/service-mesh-hub/services/csr-agent/pkg/wire"
	"github.com/solo-io/service-mesh-hub/services/internal/config"
	"go.uber.org/zap"
)

func Run(ctx context.Context) {
	ctx = config.CreateRootContext(ctx, "csr-agent")
	logger := contextutils.LoggerFrom(ctx)

	// build all the objects needed for multicluster operations
	csrAgentContext, err := wire.InitializeCsrAgent(ctx)
	if err != nil {
		logger.Fatalw("error initializing discovery clients", zap.Error(err))
	}

	istioCsrHandler := csrAgentContext.VirtualMeshCSRDataSourceFactory(
		contextutils.WithLogger(ctx, "csr_agent_data_source"),
		csrAgentContext.CsrClient,
		csrAgentContext.CsrAgentIstioProcessor,
	)
	if err = csrAgentContext.CsrController.AddEventHandler(ctx, istioCsrHandler); err != nil {
		logger.Fatalw("error VirtualMeshCSR handler", zap.Error(err))
	}

	if err = csrAgentContext.Manager.Start(); err != nil {
		logger.Fatalw("the local manager instance failed to start up or died with an error", zap.Error(err))
	}

	select {
	case <-ctx.Done():
		return
	case <-csrAgentContext.Manager.GotError():
		logger.Fatalw("the local manager encountered an error", zap.Error(csrAgentContext.Manager.Error()))
	}
}
