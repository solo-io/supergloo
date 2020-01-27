package internal_watcher

import (
	"context"
	"sync"

	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/contextutils"
	mc_manager "github.com/solo-io/mesh-projects/services/common/multicluster/manager"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/clientcmd"
)

//go:generate mockgen -source ./membership.go -destination mocks/membership.go

var (
	ClusterAddError = func(err error, clusterId string) error {
		return eris.Wrapf(err, "error during cluster add handler for (%s)", clusterId)
	}
	ClusterDeletionError = func(err error, clusterId string) error {
		return eris.Wrapf(err, "error during cluster delete handler for (%s)", clusterId)
	}
	ClusterExistsError = func(cluster, name, namespace string) error {
		return eris.Errorf("Cluster %s in the secret %s in namespace %s already exists",
			cluster, name, namespace)
	}
	KubeConfigInvalidFormatError = func(err error, cluster, name, namespace string) error {
		return eris.Wrapf(err, "invalid kube config for cluster %s in the secret %s in namespace %s",
			cluster, name, namespace)
	}
)

// this interface is meant to abstract the cluster add/delete logic for the secret watcher
type ClusterSecretHandler interface {
	AddMemberCluster(ctx context.Context, s *v1.Secret) (resync bool, err error)
	DeleteMemberCluster(ctx context.Context, s *v1.Secret) (resync bool, err error)
}

type ClusterMembershipHandler struct {
	kubeConfigReceiver mc_manager.KubeConfigHandler
	lock               sync.RWMutex
	clusterByName      map[string]*remoteCluster
}

func NewClusterMembershipHandler(kubeConfigReceiver mc_manager.KubeConfigHandler) *ClusterMembershipHandler {
	return &ClusterMembershipHandler{
		kubeConfigReceiver: kubeConfigReceiver,
		clusterByName:      make(map[string]*remoteCluster),
	}
}

// remoteCluster defines cluster struct
type remoteCluster struct {
	secretName  string
	kubeContext string
}

func (c *ClusterMembershipHandler) AddMemberCluster(ctx context.Context, s *v1.Secret) (resync bool, err error) {
	logger := contextutils.LoggerFrom(ctx)
	for clusterID, kubeConfig := range s.Data {
		// clusterID must be unique even across multiple secrets
		c.lock.RLock()
		if _, ok := c.clusterByName[clusterID]; ok {
			return false, ClusterExistsError(clusterID, c.clusterByName[clusterID].secretName, s.ObjectMeta.Namespace)
		}
		c.lock.RUnlock()
		clusterConfig, err := clientcmd.Load(kubeConfig)
		if err != nil {
			return false, KubeConfigInvalidFormatError(err, clusterID, s.GetName(), s.GetNamespace())
		}

		clientConfig := clientcmd.NewDefaultClientConfig(*clusterConfig, &clientcmd.ConfigOverrides{})
		restConfig, err := clientConfig.ClientConfig()
		if err != nil {
			return false, KubeConfigInvalidFormatError(err, clusterID, s.GetName(), s.GetNamespace())
		}

		err = c.kubeConfigReceiver.ClusterAdded(restConfig, clusterID)
		if err != nil {
			return true, ClusterAddError(err, clusterID)
		}

		logger.Infof("Adding new cluster member: %s", clusterID)
		c.lock.Lock()
		c.clusterByName[clusterID] = &remoteCluster{
			secretName:  s.GetName(),
			kubeContext: clusterConfig.CurrentContext,
		}
		c.lock.Unlock()

	}
	c.lock.RLock()
	logger.Infof("Number of remote clusters: %d", len(c.clusterByName))
	c.lock.RUnlock()
	return false, nil
}

func (c *ClusterMembershipHandler) DeleteMemberCluster(ctx context.Context, s *v1.Secret) (resync bool, err error) {
	logger := contextutils.LoggerFrom(ctx)
	for clusterID, cluster := range c.clusters() {
		if cluster.secretName == s.GetName() {
			logger.Infof("Deleting cluster member: %s", clusterID)
			err := c.kubeConfigReceiver.ClusterRemoved(clusterID)
			if err != nil {
				return true, ClusterDeletionError(err, clusterID)
			}
			c.lock.Lock()
			delete(c.clusterByName, clusterID)
			c.lock.Unlock()
		}
	}
	c.lock.RLock()
	logger.Infof("Number of remote clusters: %d", len(c.clusterByName))
	c.lock.RUnlock()
	return false, nil
}

func (c *ClusterMembershipHandler) clusters() map[string]*remoteCluster {
	c.lock.RLock()
	defer c.lock.RUnlock()
	result := make(map[string]*remoteCluster)
	for k, v := range c.clusterByName {
		result[k] = v
	}
	return result
}
