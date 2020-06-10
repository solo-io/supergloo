package mc_manager

import (
	"context"

	"github.com/avast/retry-go"
	"github.com/rotisserie/eris"
	compute_target "github.com/solo-io/service-mesh-hub/pkg/common/compute-target"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/kubeconfig"
	k8s_core_types "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	DefaultClientGetterRetryAttempts = 6
)

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
	NoClientForClusterError = func(cluster string) error {
		return eris.Errorf("could not find dynamic client for cluster %s", cluster)
	}
	ClientNotFoundError = func(clusterName string) error {
		return eris.Errorf("Client not found for cluster with name: %s", clusterName)
	}
	KubeConfigInvalidFormatError = func(err error, name, namespace string) error {
		return eris.Wrapf(err, "invalid kube config in secret %s, %s", name, namespace)
	}
)

/*
	The AsyncManagerController is meant as utility struct to async receive kube configs, and then create an AsyncManager
	for each one. It also allows consumers to register as handlers, which means they will receive the AsyncManager
	as they are created.
*/
type AsyncManagerController struct {
	factory       AsyncManagerFactory
	managers      *AsyncManagerMap
	handlers      *AsyncManagerHandlerMap
	kubeConverter kubeconfig.Converter
}

/*
	Create a new AsyncManagerController using the local manager. This will always set the "" entry to the manager
	so it is important  that the input manager is always the local manager.

	The empty string "" is the string ID representation of the local cluster
*/
func NewAsyncManagerController(
	factory AsyncManagerFactory,
	kubeConverter kubeconfig.Converter,
) *AsyncManagerController {
	managers := NewAsyncManagerMap()
	receivers := NewAsyncManagerHandler()
	mcMgr := &AsyncManagerController{
		handlers:      receivers,
		managers:      managers,
		factory:       factory,
		kubeConverter: kubeConverter,
	}
	return mcMgr
}

func (a *AsyncManagerController) AsyncManagerInformer() AsyncManagerInformer {
	return a
}
func (a *AsyncManagerController) KubeConfigHandler() compute_target.ComputeTargetCredentialsHandler {
	return a
}

// default constructor for AsyncManagerController, mostly used for testing
func NewAsyncManagerControllerWithHandlers(
	handlers *AsyncManagerHandlerMap,
	managers *AsyncManagerMap,
	factory AsyncManagerFactory,
	kubeConverter kubeconfig.Converter,
) *AsyncManagerController {
	mcMgr := &AsyncManagerController{
		handlers:      handlers,
		managers:      managers,
		factory:       factory,
		kubeConverter: kubeConverter,
	}
	return mcMgr
}

func (a *AsyncManagerController) AddHandler(informer AsyncManagerHandler, name string) error {
	return a.handlers.SetHandler(name, informer)
}

func (a *AsyncManagerController) RemoveHandler(name string) error {
	if _, ok := a.handlers.GetHandler(name); !ok {
		return InformerNotRegisteredError
	}
	a.handlers.RemoveHandler(name)
	return nil
}

func (a *AsyncManagerController) ComputeTargetAdded(ctx context.Context, secret *k8s_core_types.Secret) error {
	// Only handle secrets representing k8s cluster credentials
	// TODO change this check to the new k8s cluster specific Secret type when migrating to skv2
	if secret.Type != k8s_core_types.SecretTypeOpaque {
		return nil
	}
	clusterName, config, err := a.kubeConverter.SecretToConfig(secret)
	if err != nil {
		return KubeConfigInvalidFormatError(err, secret.GetName(), secret.GetNamespace())
	}
	mgr, err := a.factory.New(ctx, config.RestConfig, AsyncManagerOptions{
		Cluster: clusterName,
	})
	if err != nil {
		return AsyncManagerFactoryError(err, clusterName)
	}
	if err := mgr.Start(); err != nil {
		return AsyncManagerStartError(err, clusterName)
	}
	for handlerName, handler := range a.handlers.ListHandlersByName() {
		if err := handler.ClusterAdded(mgr.Context(), mgr, clusterName); err != nil {
			return InformerAddFailedError(err, handlerName, clusterName)
		}
	}
	return a.managers.SetManager(clusterName, mgr)
}

func (a *AsyncManagerController) ComputeTargetRemoved(_ context.Context, secret *k8s_core_types.Secret) error {
	// Only handle secrets representing k8s cluster credentials
	// TODO change this check to the new k8s cluster specific Secret type when migrating to skv2
	if secret.Type != k8s_core_types.SecretTypeOpaque {
		return nil
	}
	cluster := secret.GetName()
	mgr, ok := a.managers.GetManager(cluster)
	if !ok {
		return NoManagerForClusterError(cluster)
	}
	mgr.Stop()
	for handlerName, handler := range a.handlers.ListHandlersByName() {
		if err := handler.ClusterRemoved(cluster); err != nil {
			return InformerDeleteFailedError(err, handlerName, cluster)
		}
	}
	a.managers.RemoveManager(cluster)
	return nil
}

func (a *AsyncManagerController) GetClientForCluster(_ context.Context, clusterName string, opts ...retry.Option) (client.Client, error) {
	var mgr AsyncManager

	// prepend default Option so it can be overridden by input opts
	opts = append([]retry.Option{retry.Attempts(DefaultClientGetterRetryAttempts)}, opts...)

	err := retry.Do(func() error {
		var ok bool
		mgr, ok = a.managers.GetManager(clusterName)
		if !ok {
			return ClientNotFoundError(clusterName)
		}
		return nil
	}, opts...)
	if err != nil {
		return nil, err
	}
	return mgr.Manager().GetClient(), nil
}
