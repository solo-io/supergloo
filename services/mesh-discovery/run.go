package mesh_discovery

import (
	"context"

	zephyr_core "github.com/solo-io/mesh-projects/pkg/clients/zephyr/discovery"
	"github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/multicluster/controllers"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/mesh-projects/pkg/common/concurrency"
	"github.com/solo-io/mesh-projects/services/common/multicluster"
	mc_manager "github.com/solo-io/mesh-projects/services/common/multicluster/manager"
	"github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/mesh"
	md_multicluster "github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/multicluster"
	"github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/wire"
	"go.uber.org/zap"
)

func Run(ctx context.Context) {
	logger := contextutils.LoggerFrom(ctx)

	// build all the objects needed for multicluster operations
	discoveryContext, err := wire.InitializeDiscovery(ctx)
	if err != nil {
		logger.Fatalw("error initializing discovery clients", zap.Error(err))
	}

	// this is our main entrypoint for mesh-discovery
	// when it detects a new cluster get registered, it builds a deployment controller
	// with the controller factory, and attaches the given mesh finders to it
	deploymentHandler, err := md_multicluster.NewDiscoveryClusterHandler(
		ctx,
		discoveryContext.ImageParser,
		discoveryContext.MultiClusterDeps.LocalManager,
		controllers.NewDeploymentControllerFactory(),
		zephyr_core.NewMeshClient(discoveryContext.MultiClusterDeps.LocalManager),
		[]mesh.MeshScanner{
			discoveryContext.IstioMeshScanner,
			discoveryContext.ConsulConnectMeshScanner,
			discoveryContext.LinkerdMeshScanner,
		},
		controllers.NewPodControllerFactory(),
		zephyr_core.NewMeshWorkloadClient(discoveryContext.MultiClusterDeps.LocalManager.Manager().GetClient()),
		concurrency.NewThreadSafeMap,
	)

	if err != nil {
		logger.Fatalw("error initializing discovery cluster handler", zap.Error(err))
	}

	// block until we die; RIP
	err = multicluster.SetupAndStartLocalManager(
		discoveryContext.MultiClusterDeps,

		// need to be sure to register the v1alpha1 CRDs with the controller runtime
		[]mc_manager.AsyncManagerStartOptionsFunc{multicluster.AddSchemeV1Alpha1},

		[]multicluster.NamedAsyncManagerHandler{{
			Name:                "discovery-controller",
			AsyncManagerHandler: deploymentHandler,
		}},
	)

	if err != nil {
		logger.Fatalw("the local manager instance failed to start up or died with an error", zap.Error(err))
	}
}
