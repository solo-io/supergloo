package mc_manager

import (
	"context"

	"github.com/rotisserie/eris"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

//go:generate mockgen -source controller.go -destination ./mocks/controller.go

var (
	AsyncManagerFactoryError = func(err error, cluster string) error {
		return eris.Wrapf(err, "failed to create new async manager for %s", cluster)
	}
	AsyncManagerStartError = func(err error, cluster string) error {
		return eris.Wrapf(err, "failed to start new async manager for %s", cluster)
	}
	InformerDeleteFailedError = func(err error, handler, cluster string) error {
		return eris.Wrapf(err, "delete cluster manager callback failed for"+
			" handler: %s on cluster: %s", handler, cluster)
	}
	InformerAddFailedError = func(err error, handler, cluster string) error {
		return eris.Wrapf(err, "add cluster manager callback failed for"+
			" handler: %s on cluster: %s", handler, cluster)
	}
	NoManagerForClusterError = func(cluster string) error {
		return eris.Errorf("could not find manager for cluster %s", cluster)
	}
)

// these functions are intended to be used as callbacks for a resource watcher, where the
// resources represent KubeConfigs
type KubeConfigHandler interface {
	ClusterAdded(cfg *rest.Config, name string) error
	ClusterRemoved(name string) error
}

// These functions are intended to be used as an extension to the KubeConfigHandler.
// Only one manager needs to be created per cluster, so these callbacks will be
// called when a manager has been created for a given cluster
type AsyncManagerHandler interface {
	ClusterAdded(mgr AsyncManager, name string) error
	ClusterRemoved(name string) error
}

// these functions are intended to be used as callbacks for a resource watcher, where the
// resources are async managers for a given kubeconfig/cluster
type AsyncManagerInformer interface {
	AddHandler(informer AsyncManagerHandler, name string) error
	RemoveHandler(name string) error
}

/*
	The AsyncManagerController is meant as utility struct to async receive kube configs, and then create an AsyncManager
	for each one. It also allows consumers to register as handlers, which means they will receive the AsyncManager
	as they are created.
*/
type AsyncManagerController struct {
	ctx     context.Context
	factory AsyncManagerFactory

	managers *AsyncManagerMap
	handlers *AsyncManagerHandlerMap
}

/*
	Create a new AsyncManagerController using the local manager. This will always set the "" entry to the manager
	so it is important  that the input manager is always the local manager.

	The empty string "" is the string ID representation of the local cluster
*/
func NewAsyncManagerControllerFromLocal(ctx context.Context, mgr manager.Manager,
	factory AsyncManagerFactory) *AsyncManagerController {

	ctxMgr := NewAsyncManager(ctx, mgr)

	managers := NewAsyncManagerMap()
	managers.SetManager("", ctxMgr)

	receivers := NewAsyncManagerHandler()
	mcMgr := &AsyncManagerController{
		ctx:      ctx,
		handlers: receivers,
		managers: managers,
		factory:  factory,
	}
	return mcMgr
}

func (m *AsyncManagerController) AsyncManagerInformer() AsyncManagerInformer {
	return m
}
func (m *AsyncManagerController) KubeConfigHandler() KubeConfigHandler {
	return m
}

// default constructor for AsyncManagerController, mostly used for testing
func NewAsyncManagerController(ctx context.Context, handlers *AsyncManagerHandlerMap, managers *AsyncManagerMap,
	factory AsyncManagerFactory) *AsyncManagerController {
	mcMgr := &AsyncManagerController{
		ctx:      ctx,
		handlers: handlers,
		managers: managers,
		factory:  factory,
	}
	return mcMgr
}

func (m *AsyncManagerController) AddHandler(informer AsyncManagerHandler, name string) error {
	return m.handlers.SetHandler(name, informer)
}

func (m *AsyncManagerController) RemoveHandler(name string) error {
	if _, ok := m.handlers.GetHandler(name); !ok {
		return InformerNotRegisteredError
	}
	m.handlers.RemoveHandler(name)
	return nil
}

func (m *AsyncManagerController) ClusterAdded(cfg *rest.Config, name string) error {
	mgr, err := m.factory.New(m.ctx, cfg, AsyncManagerOptions{})
	if err != nil {
		return AsyncManagerFactoryError(err, name)
	}
	if err := mgr.Start(); err != nil {
		return AsyncManagerStartError(err, name)
	}
	for handlerName, handler := range m.handlers.ListHandlersByName() {
		if err := handler.ClusterAdded(mgr, name); err != nil {
			return InformerAddFailedError(err, handlerName, name)
		}
	}
	return m.managers.SetManager(name, mgr)
}

func (m *AsyncManagerController) ClusterRemoved(name string) error {
	mgr, ok := m.managers.GetManager(name)
	if !ok {
		return NoManagerForClusterError(name)
	}
	mgr.Stop()
	for handlerName, handler := range m.handlers.ListHandlersByName() {
		if err := handler.ClusterRemoved(name); err != nil {
			return InformerDeleteFailedError(err, handlerName, name)
		}
	}
	m.managers.RemoveManager(name)
	return nil
}
