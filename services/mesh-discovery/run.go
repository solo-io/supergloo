package mesh_discovery

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/container-runtime"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/compute-target/k8s"
	md_multicluster "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/compute-target"
	k8s_tenancy "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/cluster-tenancy/k8s"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh-workload/k8s/istio"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh-workload/k8s/linkerd"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh/k8s"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/wire"
	"go.uber.org/zap"
)

func Run(rootCtx context.Context) {
	ctx := container_runtime.CreateRootContext(rootCtx, "mesh-discovery")

	logger := contextutils.LoggerFrom(ctx)

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
		md_multicluster.MeshWorkloadScannerFactoryImplementations{
			types.MeshType_ISTIO1_5: istio.NewIstioMeshWorkloadScanner,
			types.MeshType_ISTIO1_6: istio.NewIstioMeshWorkloadScanner,
			types.MeshType_LINKERD:  linkerd.NewLinkerdMeshWorkloadScanner,
			types.MeshType_APPMESH:  discoveryContext.MeshDiscovery.AppMeshWorkloadScannerFactory,
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
