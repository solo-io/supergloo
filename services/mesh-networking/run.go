package mesh_networking

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/mesh-projects/pkg/env"
	"github.com/solo-io/mesh-projects/services/common/multicluster"
	mc_manager "github.com/solo-io/mesh-projects/services/common/multicluster/manager"
	mc_predicate "github.com/solo-io/mesh-projects/services/common/multicluster/predicate"
	"github.com/solo-io/mesh-projects/services/internal/config"
	group_controller "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/groups/controller"
	"github.com/solo-io/mesh-projects/services/mesh-networking/pkg/wire"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/util/sets"
)

func Run(ctx context.Context) {
	ctx = config.CreateRootContext(ctx, "mesh-networking")
	logger := contextutils.LoggerFrom(ctx)

	// build all the objects needed for multicluster operations
	meshNetworkingContext, err := wire.InitializeMeshNetworking(ctx)
	if err != nil {
		logger.Fatalw("error initializing mesh networking clients", zap.Error(err))
	}

	// block until we die; RIP
	err = multicluster.SetupAndStartLocalManager(
		meshNetworkingContext.MultiClusterDeps,
		[]mc_manager.AsyncManagerStartOptionsFunc{
			// start the mesh group event handler watching only the default write namespace, make this configurable?
			group_controller.NewMeshGroupControllerStarter(
				meshNetworkingContext.MeshGroupEventHandler,
				mc_predicate.WhitelistedNamespacePredicateProvider(sets.NewString(env.DefaultWriteNamespace)),
				group_controller.IgnoreStatusUpdatePredicate(ctx),
			),
		},
		[]multicluster.NamedAsyncManagerHandler{{
			Name:                "mesh-networking-multicluster-controller",
			AsyncManagerHandler: meshNetworkingContext.MeshNetworkingClusterHandler,
		}},
	)

	if err != nil {
		logger.Fatalw("the local manager instance failed to start up or died with an error", zap.Error(err))
	}
}
