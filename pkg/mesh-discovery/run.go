package mesh_discovery

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	mc_manager "github.com/solo-io/service-mesh-hub/pkg/common/compute-target/k8s"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/common/container-runtime"
	md_multicluster "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/compute-target"
	k8s_tenancy "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/discovery/cluster-tenancy/k8s"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/discovery/mesh/k8s"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/starter"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/wire"
	"go.uber.org/zap"
)

func Run(rootCtx context.Context) {
	ctx := container_runtime.CreateRootContext(rootCtx, "mesh-discovery")
	logger := contextutils.LoggerFrom(ctx)

	start := starter.NewStarter()
	start.Start(ctx, starter.Options{})

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
		localManager,
		[]k8s.MeshScanner{
			discoveryContext.MeshDiscovery.IstioMeshScanner,
			discoveryContext.MeshDiscovery.ConsulConnectMeshScanner,
			discoveryContext.MeshDiscovery.LinkerdMeshScanner,
		},
		[]k8s_tenancy.ClusterTenancyScannerFactory{
			discoveryContext.ClusterTenancy.AppMeshClusterTenancyScannerFactory,
		},
		discoveryContext,
	)

	if err != nil {
		logger.Fatalw("error initializing discovery cluster handler", zap.Error(err))
	}

	// block until we die; RIP
	err = mc_manager.SetupAndStartLocalManager(
		discoveryContext.MultiClusterDeps,

		// need to be sure to register the v1alpha1 CRDs with the controller runtime
		[]mc_manager.AsyncManagerStartOptionsFunc{
			mc_manager.AddAllV1Alpha1ToScheme,
		},

		[]mc_manager.NamedAsyncManagerHandler{{
			Name:                "discovery-controller",
			AsyncManagerHandler: deploymentHandler,
		}},
	)

	if err != nil {
		logger.Fatalw("the local manager instance failed to start up or died with an error", zap.Error(err))
	}
}
