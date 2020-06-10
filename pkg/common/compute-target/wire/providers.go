package mc_wire

import (
	"context"

	"github.com/google/wire"
	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/kubeconfig"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/multicluster"
	compute_target "github.com/solo-io/service-mesh-hub/pkg/common/compute-target"
	mc_manager "github.com/solo-io/service-mesh-hub/pkg/common/compute-target/k8s"
	mc_watcher "github.com/solo-io/service-mesh-hub/pkg/common/compute-target/secret-event-handler"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	k8s_manager "sigs.k8s.io/controller-runtime/pkg/manager"
)

// package up all the dependencies needed to build an instance of `MultiClusterDependencies`,
// which should be considered to be the only output type you need to consume from this provider set
var MulticlusterProviderSet = wire.NewSet(
	ClusterProviderSet,
	AsyncManagerFactoryProvider,
	KubeClusterCredentialsHandlerProvider,
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
	computeTargetCredentialsHandlers []compute_target.ComputeTargetCredentialsHandler,
) mc_manager.AsyncManagerStartOptionsFunc {
	return mc_watcher.StartLocalManager(computeTargetCredentialsHandlers)
}

func LocalManagerProvider(ctx context.Context, cfg *rest.Config) (mc_manager.AsyncManager, error) {
	mgr, err := k8s_manager.New(cfg, k8s_manager.Options{})
	if err != nil {
		return nil, err
	}

	return mc_manager.NewAsyncManager(ctx, mgr), nil
}

func DynamicClientProvider(mgr mc_manager.AsyncManager) client.Client {
	return mgr.Manager().GetClient()
}

func AsyncManagerFactoryProvider() mc_manager.AsyncManagerFactory {
	return mc_manager.NewAsyncManagerFactory()
}

func KubeClusterCredentialsHandlerProvider(
	kubeConverter kubeconfig.Converter,
) *mc_manager.AsyncManagerController {
	return mc_manager.NewAsyncManagerController(mc_manager.NewAsyncManagerFactory(), kubeConverter)
}

func DynamicClientGetterProvider(controller *mc_manager.AsyncManagerController) multicluster.DynamicClientGetter {
	return controller
}

func MulticlusterDependenciesProvider(
	ctx context.Context,
	localManager mc_manager.AsyncManager,
	asyncManagerController *mc_manager.AsyncManagerController,
	localManagerStarter mc_manager.AsyncManagerStartOptionsFunc,
) mc_manager.MultiClusterDependencies {
	return mc_manager.MultiClusterDependencies{
		LocalManager:         localManager,
		AsyncManagerInformer: asyncManagerController.AsyncManagerInformer(),
		KubeConfigHandler:    asyncManagerController.KubeConfigHandler(),
		LocalManagerStarter:  localManagerStarter,
		Context:              ctx,
	}
}
