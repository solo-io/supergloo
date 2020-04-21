package mc_wire

import (
	"context"

	"github.com/google/wire"
	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/service-mesh-hub/services/common/multicluster"
	manager2 "github.com/solo-io/service-mesh-hub/services/common/multicluster/manager"
	mc_watcher "github.com/solo-io/service-mesh-hub/services/common/multicluster/watcher"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/rest-api/aws"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// package up all the dependencies needed to build an instance of `MultiClusterDependencies`,
// which should be considered to be the only output type you need to consume from this provider set
var MulticlusterProviderSet = wire.NewSet(
	ClusterProviderSet,
	AsyncManagerFactoryProvider,
	AsyncManagerControllerProvider,
	LocalManagerStarterProvider,
	MulticlusterDependenciesProvider,
)

// Used for operators which do not have any multi cluster dependencies, such as the csr-agent
var ClusterProviderSet = wire.NewSet(
	LocalKubeConfigProvider,
	LocalManagerProvider,
	DynamicClientProvider,
)

func LocalKubeConfigProvider() (*rest.Config, error) {
	return kubeutils.GetConfig("", "")
}

func LocalManagerStarterProvider(
	controller *manager2.AsyncManagerController,
	awsCredsHandler aws.AwsCredsHandler,
) manager2.AsyncManagerStartOptionsFunc {
	return mc_watcher.StartLocalManager(controller, awsCredsHandler)
}

func LocalManagerProvider(ctx context.Context, cfg *rest.Config) (manager2.AsyncManager, error) {
	mgr, err := manager.New(cfg, manager.Options{})
	if err != nil {
		return nil, err
	}

	return manager2.NewAsyncManager(ctx, mgr), nil
}

func DynamicClientProvider(mgr manager2.AsyncManager) client.Client {
	return mgr.Manager().GetClient()
}

func AsyncManagerFactoryProvider() manager2.AsyncManagerFactory {
	return manager2.NewAsyncManagerFactory()
}

func AsyncManagerControllerProvider(ctx context.Context, localManager manager2.AsyncManager) *manager2.AsyncManagerController {
	return manager2.NewAsyncManagerControllerFromLocal(ctx, localManager.Manager(), manager2.NewAsyncManagerFactory())
}

func DynamicClientGetterProvider(controller *manager2.AsyncManagerController) manager2.DynamicClientGetter {
	return controller
}

func MulticlusterDependenciesProvider(
	ctx context.Context,
	localManager manager2.AsyncManager,
	asyncManagerController *manager2.AsyncManagerController,
	localManagerStarter manager2.AsyncManagerStartOptionsFunc,
) multicluster.MultiClusterDependencies {
	return multicluster.MultiClusterDependencies{
		LocalManager:         localManager,
		AsyncManagerInformer: asyncManagerController.AsyncManagerInformer(),
		KubeConfigHandler:    asyncManagerController.KubeConfigHandler(),
		LocalManagerStarter:  localManagerStarter,
		Context:              ctx,
	}
}
