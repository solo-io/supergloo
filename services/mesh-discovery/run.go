package mesh_discovery

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	zephyr_core "github.com/solo-io/mesh-projects/pkg/clients/zephyr/discovery"
	"github.com/solo-io/mesh-projects/services/common/multicluster"
	mc_manager "github.com/solo-io/mesh-projects/services/common/multicluster/manager"
	"github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/discovery/mesh"
	mesh_workload "github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/discovery/mesh-workload"
	md_multicluster "github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/multicluster"
	"github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/wire"
	"go.uber.org/zap"
)

func Run(ctx context.Context) {
	logger := contextutils.LoggerFrom(ctx)

	// >:(
	// the default global zap logger, which controller-runtime uses, is a no-op logger
	// https://github.com/uber-go/zap/blob/5dab9368974ab1352e4245f9d33e5bce4c23a034/global.go#L41
	zap.ReplaceGlobals(logger.Desugar())

	// build all the objects needed for multicluster operations
	discoveryContext, err := wire.InitializeDiscovery(ctx)
	if err != nil {
		logger.Fatalw("error initializing discovery clients", zap.Error(err))
	}

	localManager := discoveryContext.MultiClusterDeps.LocalManager

	// this is our main entrypoint for mesh-discovery
	// when it detects a new cluster get registered, it builds a deployment controller
	// with the controller factory, and attaches the given mesh finders to it
	deploymentHandler, err := md_multicluster.NewDiscoveryClusterHandler(
		ctx,
		localManager,
		zephyr_core.NewMeshClient(localManager),
		[]mesh.MeshScanner{
			discoveryContext.MeshDiscovery.IstioMeshScanner,
			discoveryContext.MeshDiscovery.ConsulConnectMeshScanner,
			discoveryContext.MeshDiscovery.LinkerdMeshScanner,
		},
		[]mesh_workload.MeshWorkloadScannerFactory{
			mesh_workload.NewIstioMeshWorkloadScanner,
		},
		zephyr_core.NewMeshWorkloadClient(discoveryContext.MultiClusterDeps.LocalManager.Manager().GetClient()),
		discoveryContext,
	)

	if err != nil {
		logger.Fatalw("error initializing discovery cluster handler", zap.Error(err))
	}

	// block until we die; RIP
	err = multicluster.SetupAndStartLocalManager(
		discoveryContext.MultiClusterDeps,

		// need to be sure to register the v1alpha1 CRDs with the controller runtime
		[]mc_manager.AsyncManagerStartOptionsFunc{multicluster.AddAllV1Alpha1ToScheme},

		[]multicluster.NamedAsyncManagerHandler{{
			Name:                "discovery-controller",
			AsyncManagerHandler: deploymentHandler,
		}},
	)

	if err != nil {
		logger.Fatalw("the local manager instance failed to start up or died with an error", zap.Error(err))
	}
}
