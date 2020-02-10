package mesh_group

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/mesh-projects/pkg/env"
	"github.com/solo-io/mesh-projects/services/common/multicluster"
	mc_manager "github.com/solo-io/mesh-projects/services/common/multicluster/manager"
	mc_predicate "github.com/solo-io/mesh-projects/services/common/multicluster/predicate"
	"github.com/solo-io/mesh-projects/services/mesh-group/pkg/controller"
	mg_multicluster "github.com/solo-io/mesh-projects/services/mesh-group/pkg/multicluster"
	"github.com/solo-io/mesh-projects/services/mesh-group/pkg/wire"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/util/sets"
)

func Run(ctx context.Context) {
	logger := contextutils.LoggerFrom(ctx)

	// build all the objects needed for multicluster operations
	meshDiscoveryContext, err := wire.InitializeMeshGroup(ctx)
	if err != nil {
		logger.Fatalw("error initializing mesh group clients", zap.Error(err))
	}

	// this is our main entrypoint for mesh-grouping
	addClusterHandler, err := mg_multicluster.NewMeshGroupClusterHandler(
		ctx,
		meshDiscoveryContext.MultiClusterDeps.LocalManager,
	)
	if err != nil {
		logger.Fatalw("error initializing mesh group cluster handler", zap.Error(err))
	}

	// block until we die; RIP
	err = multicluster.SetupAndStartLocalManager(
		meshDiscoveryContext.MultiClusterDeps,
		[]mc_manager.AsyncManagerStartOptionsFunc{
			// start the mesh group event handler watching only the default write namespace, make this configurable?
			controller.NewMeshGroupControllerStarter(
				meshDiscoveryContext.MeshGroupEventHandler,
				mc_predicate.WhitelistedNamespacePredicateProvider(sets.NewString(env.DefaultWriteNamespace)),
				controller.IgnoreStatusUpdatePredicate(ctx),
			),
		},
		[]multicluster.NamedAsyncManagerHandler{{
			Name:                "mesh-group-controller",
			AsyncManagerHandler: addClusterHandler,
		}},
	)

	if err != nil {
		logger.Fatalw("the local manager instance failed to start up or died with an error", zap.Error(err))
	}
}
