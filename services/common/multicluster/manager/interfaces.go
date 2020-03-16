package mc_manager

import (
	"context"

	"github.com/avast/retry-go"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

//go:generate mockgen -source interfaces.go -destination ./mocks/interfaces.go

// these functions are intended to be used as callbacks for a resource watcher, where the
// resources represent KubeConfigs
type KubeConfigHandler interface {
	ClusterAdded(cfg *rest.Config, clusterName string) error
	ClusterRemoved(cluster string) error
}

// These functions are intended to be used as an extension to the KubeConfigHandler.
// Only one manager needs to be created per cluster, so these callbacks will be
// called when a manager has been created for a given cluster
type AsyncManagerHandler interface {
	ClusterAdded(ctx context.Context, mgr AsyncManager, clusterName string) error
	ClusterRemoved(cluster string) error
}

// these functions are intended to be used as callbacks for a resource watcher, where the
// resources are async managers for a given kubeconfig/cluster
type AsyncManagerInformer interface {
	AddHandler(informer AsyncManagerHandler, name string) error
	RemoveHandler(name string) error
}

/*
	This interface is meant to represent an asynchronous wrapper on top of a controller-runtime manager.
	It comes with opinionated ways to start and stop managers in go routines, as well as check their error
	status
*/
type AsyncManager interface {
	// returns the manager associated with the async manager
	Manager() manager.Manager
	// returns the context of the async manager
	Context() context.Context
	// returns the err which has occurred
	Error() error
	// returns the channel which is closed when an err occurs
	GotError() <-chan struct{}
	// start the async manager, does not block, will signal the `GotError()` channel if an error occurs
	Start(opts ...AsyncManagerStartOptionsFunc) error
	// stops the async manager, does not block
	Stop()
}

type AsyncManagerFactory interface {
	New(parentCtx context.Context, cfg *rest.Config,
		opts AsyncManagerOptions) (AsyncManager, error)
}

// Simple map get interface to expose the map of dynamic clients to the local controller
type DynamicClientGetter interface {
	GetClientForCluster(clusterName string, opts ...retry.Option) (client.Client, bool)
}
