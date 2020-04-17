package csr_agent

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	security_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/security.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/bootstrap"
	"github.com/solo-io/service-mesh-hub/services/csr-agent/pkg/wire"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func Run(ctx context.Context) {
	ctx = bootstrap.CreateRootContext(ctx, "csr-agent")
	logger := contextutils.LoggerFrom(ctx)

	// build all the objects needed for multicluster operations
	csrAgentContext, err := wire.InitializeCsrAgent(ctx)
	if err != nil {
		logger.Fatalw("error initializing discovery clients", zap.Error(err))
	}

	if err = csrAgentContext.Manager.Start(startComponents(csrAgentContext)); err != nil {
		logger.Fatalw("the local manager instance failed to start up or died with an error", zap.Error(err))
	}

	select {
	case <-ctx.Done():
		return
	case <-csrAgentContext.Manager.GotError():
		logger.Fatalw("the local manager encountered an error", zap.Error(csrAgentContext.Manager.Error()))
	}
}

// Controller-runtime Watches require the manager to be started first, otherwise it will block indefinitely
// Thus we initialize all components (and their associated watches) as an AsyncManagerStartOptionsFunc.
func startComponents(csrAgentContext wire.CsrAgentContext) func(context.Context, manager.Manager) error {
	return func(ctx context.Context, m manager.Manager) error {
		logger := contextutils.LoggerFrom(ctx)
		// Ensure that VirtualMeshCertificateSigningRequest is registered with the controller-runtime scheme
		err := security_v1alpha1.AddToScheme(m.GetScheme())
		if err != nil {
			logger.Errorf("error adding security to scheme", zap.Error(err))
			return err
		}
		istioCsrHandler := csrAgentContext.VirtualMeshCSRDataSourceFactory(
			contextutils.WithLogger(ctx, "csr_agent_data_source"),
			csrAgentContext.CsrClient,
			csrAgentContext.CsrAgentIstioProcessor,
		)
		if err := csrAgentContext.CsrEventWatcher.AddEventHandler(ctx, istioCsrHandler); err != nil {
			logger.Fatalw("error VirtualMeshCSR handler", zap.Error(err))
		}
		return nil
	}
}
