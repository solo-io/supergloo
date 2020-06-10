package mesh_networking

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/common/container-runtime"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/compute-target/k8s"
	"github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/wire"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func Run(ctx context.Context) {
	ctx = container_runtime.CreateRootContext(ctx, "mesh-networking")
	logger := contextutils.LoggerFrom(ctx)

	// build all the objects needed for multicluster operations
	meshNetworkingContext, err := wire.InitializeMeshNetworking(
		contextutils.WithLogger(ctx, "access_control_enforcer"),
	)
	if err != nil {
		logger.Fatalw("error initializing mesh networking clients", zap.Error(err))
	}

	// block until we die; RIP
	err = mc_manager.SetupAndStartLocalManager(
		meshNetworkingContext.MultiClusterDeps,
		[]mc_manager.AsyncManagerStartOptionsFunc{
			mc_manager.AddAllV1Alpha1ToScheme,
			mc_manager.AddAllIstioToScheme,
			mc_manager.AddAllLinkerdToScheme,
			startComponents(meshNetworkingContext),
		},
		[]mc_manager.NamedAsyncManagerHandler{{
			Name:                "mesh-networking-multicluster-controller",
			AsyncManagerHandler: meshNetworkingContext.MeshNetworkingClusterHandler,
		}},
	)

	if err != nil {
		logger.Fatalw("the local manager instance failed to start up or died with an error", zap.Error(err))
	}
}

// Controller-runtime Watches require the manager to be started first, otherwise it will block indefinitely
// Thus we initialize all components (and their associated watches) as an AsyncManagerStartOptionsFunc.
func startComponents(meshNetworkingContext wire.MeshNetworkingContext) func(context.Context, manager.Manager) error {
	return func(ctx context.Context, m manager.Manager) error {
		logger := contextutils.LoggerFrom(ctx)
		var err error
		if err = meshNetworkingContext.MeshNetworkingSnapshotContext.StartListening(
			contextutils.WithLogger(ctx, "mesh_networking_snapshot_listener"),
		); err != nil {
			logger.Fatalw("error initializing mesh networking snapshot listener", zap.Error(err))
		}
		// start the TrafficPolicyTranslator
		err = meshNetworkingContext.TrafficPolicyTranslator.Start(
			contextutils.WithLogger(ctx, "traffic_policy_translator"),
		)
		if err != nil {
			logger.Fatalw("error initializing TrafficPolicyTranslator", zap.Error(err))
		}

		err = meshNetworkingContext.AccessControlPolicyTranslator.Start(
			contextutils.WithLogger(ctx, "access_control_policy_translator"),
		)
		if err != nil {
			logger.Fatalw("error initializing AccessControlPolicyTranslator", zap.Error(err))
		}

		err = meshNetworkingContext.GlobalAccessPolicyEnforcer.Start(
			contextutils.WithLogger(ctx, "global_access_control_policy_enforcer"),
		)
		if err != nil {
			logger.Fatalw("error initializing GlobalAccessControlPolicyEnforcer", zap.Error(err))
		}

		err = meshNetworkingContext.FederationResolver.Start(
			contextutils.WithLogger(ctx, "federation_resolver"),
		)
		if err != nil {
			logger.Fatalw("error initializing FederationResolver", zap.Error(err))
		}
		return nil
	}
}
