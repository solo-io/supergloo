package multicluster

import (
	"context"

	mc_manager "github.com/solo-io/service-mesh-hub/services/common/multicluster/manager"
)

const (
	MultiClusterLabel      = "solo.io/kubeconfig"
	MultiClusterController = "multi-cluster-controller"
)

// associate a name with the async manager handler
type NamedAsyncManagerHandler struct {
	Name                string
	AsyncManagerHandler mc_manager.AsyncManagerHandler
}

// all the dependencies you should need to start doing multicluster operations
// commonly used as input to `SetupAndStartLocalManager`
type MultiClusterDependencies struct {
	LocalManager         mc_manager.AsyncManager
	AsyncManagerInformer mc_manager.AsyncManagerInformer
	KubeConfigHandler    mc_manager.KubeConfigHandler
	LocalManagerStarter  mc_manager.AsyncManagerStartOptionsFunc
	Context              context.Context
}

// handles attaching all the handlers and running all the startup functions
// ensures that the local manager starter runs last
// blocks the current goroutine until <-ctx.Done() or until the local manager reports an error
func SetupAndStartLocalManager(
	dependencies MultiClusterDependencies,
	startupFuncs []mc_manager.AsyncManagerStartOptionsFunc,
	asyncManagerHandlers []NamedAsyncManagerHandler,
) error {

	localManager := dependencies.LocalManager

	for _, h := range asyncManagerHandlers {
		err := dependencies.AsyncManagerInformer.AddHandler(h.AsyncManagerHandler, h.Name)
		if err != nil {
			return err
		}
	}

	allStartupFuncs := append([]mc_manager.AsyncManagerStartOptionsFunc{}, startupFuncs...)

	// make sure the local manager gets started last
	allStartupFuncs = append(startupFuncs, dependencies.LocalManagerStarter)

	err := localManager.Start(allStartupFuncs...)
	if err != nil {
		return err
	}

	select {
	case <-dependencies.Context.Done():
		return nil
	case <-localManager.GotError():
		return localManager.Error()
	}
}
