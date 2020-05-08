package mesh_networking

import (
	"context"
	"time"

	"github.com/solo-io/go-utils/contextutils"
	v1alpha12 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	v1 "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/apps/v1"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/bootstrap"
	"github.com/solo-io/service-mesh-hub/pkg/reconciliation"
	"github.com/solo-io/service-mesh-hub/pkg/selector"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/compute-target/k8s"
	traffic_policy_validation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/validation"
	"github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/wire"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func Run(ctx context.Context) {
	ctx = bootstrap.CreateRootContext(ctx, "mesh-networking")
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
		//err = meshNetworkingContext.TrafficPolicyTranslator.Start(
		//	contextutils.WithLogger(ctx, "traffic_policy_translator"),
		//)
		//if err != nil {
		//	logger.Fatalw("error initializing TrafficPolicyTranslator", zap.Error(err))
		//}
		validationReconciler := reconciliation.NewPeriodicReconciler()
		go validationReconciler.Start(ctx, time.Second, "tp validation",
			traffic_policy_validation.NewValidationLoop(
				v1alpha1.NewTrafficPolicyClient(m.GetClient()),
				v1alpha12.NewMeshServiceClient(m.GetClient()),
				traffic_policy_validation.NewValidator(selector.NewResourceSelector(
					v1alpha12.NewMeshServiceClient(m.GetClient()),
					v1alpha12.NewMeshWorkloadClient(m.GetClient()),
					func(client client.Client) v1.DeploymentClient {
						return v1.NewDeploymentClient(client)
					},
					nil,
				)),
			).RunOnce)

		err = meshNetworkingContext.AccessControlPolicyTranslator.Start(
			contextutils.WithLogger(ctx, "access_control_policy_translator"),
		)
		if err != nil {
			logger.Fatalw("error intitializing AccessControlPolicyTranslator", zap.Error(err))
		}

		err = meshNetworkingContext.GlobalAccessPolicyEnforcer.Start(
			contextutils.WithLogger(ctx, "global_access_control_policy_enforcer"),
		)
		if err != nil {
			logger.Fatalw("error intitializing GlobalAccessControlPolicyEnforcer", zap.Error(err))
		}

		err = meshNetworkingContext.FederationResolver.Start(
			contextutils.WithLogger(ctx, "federation_resolver"),
		)
		if err != nil {
			logger.Fatalw("error intitializing FederationResolver", zap.Error(err))
		}
		return nil
	}
}
