package mesh_discovery

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/bootstrap"
	"github.com/solo-io/service-mesh-hub/services/common/multicluster"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/multicluster/manager"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/k8s/mesh"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/k8s/mesh-workload/appmesh"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/k8s/mesh-workload/istio"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/k8s/mesh-workload/linkerd"
	md_multicluster "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/mesh-platform"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/wire"
	"go.uber.org/zap"
)

func Run(rootCtx context.Context) {
	ctx := bootstrap.CreateRootContext(rootCtx, "mesh-discovery")

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
		[]mesh.MeshScanner{
			discoveryContext.MeshDiscovery.IstioMeshScanner,
			discoveryContext.MeshDiscovery.ConsulConnectMeshScanner,
			discoveryContext.MeshDiscovery.LinkerdMeshScanner,
		},
		md_multicluster.MeshWorkloadScannerFactoryImplementations{
			types.MeshType_ISTIO:   istio.NewIstioMeshWorkloadScanner,
			types.MeshType_LINKERD: linkerd.NewLinkerdMeshWorkloadScanner,
			types.MeshType_APPMESH: appmesh.NewAppMeshWorkloadScanner,
		},
		discoveryContext,
	)

	if err != nil {
		logger.Fatalw("error initializing discovery cluster handler", zap.Error(err))
	}

	// block until we die; RIP
	err = multicluster.SetupAndStartLocalManager(
		discoveryContext.MultiClusterDeps,

		// need to be sure to register the v1alpha1 CRDs with the controller runtime
		[]mc_manager.AsyncManagerStartOptionsFunc{
			multicluster.AddAllV1Alpha1ToScheme,
		},

		[]multicluster.NamedAsyncManagerHandler{{
			Name:                "discovery-controller",
			AsyncManagerHandler: deploymentHandler,
		}},
	)

	if err != nil {
		logger.Fatalw("the local manager instance failed to start up or died with an error", zap.Error(err))
	}
}
