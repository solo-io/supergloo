package internal_watcher

import (
	"context"
	"sync"

	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/aws_creds"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/kube"
	"github.com/solo-io/service-mesh-hub/services/common/multicluster/manager/k8s_manager"
	"github.com/solo-io/service-mesh-hub/services/common/multicluster/manager/rest_watcher/aws"
	v1 "k8s.io/api/core/v1"
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
type MeshAPISecretHandler interface {
	AddMemberMeshAPI(ctx context.Context, s *v1.Secret) (resync bool, err error)
	DeleteMemberCluster(ctx context.Context, s *v1.Secret) (resync bool, err error)
}

type MeshAPIMembershipHandler struct {
	kubeConfigReceiver k8s_manager.KubeConfigHandler
	awsCredsHandler    aws.AwsCredsHandler
	lock               sync.RWMutex
	clusterByName      map[string]*remoteCluster
	kubeConverter      kube.Converter
}

func NewClusterMembershipHandler(
	kubeConfigReceiver k8s_manager.KubeConfigHandler,
	awsCredsHandler aws.AwsCredsHandler,
	kubeConverter kube.Converter,
) *MeshAPIMembershipHandler {
	return &MeshAPIMembershipHandler{
		kubeConfigReceiver: kubeConfigReceiver,
		awsCredsHandler:    awsCredsHandler,
		clusterByName:      make(map[string]*remoteCluster),
		kubeConverter:      kubeConverter,
	}
}

// remoteCluster defines cluster struct
type remoteCluster struct {
	secretName  string
	kubeContext string
}

func (m *MeshAPIMembershipHandler) AddMemberMeshAPI(ctx context.Context, s *v1.Secret) (resync bool, err error) {
	logger := contextutils.LoggerFrom(ctx)
	if s.Type == aws_creds.AWSSecretType {
		err := m.awsCredsHandler.RestAPIAdded(ctx, s)
		if err != nil {
			logger.Errorf("Error initialize AwsCredsHandler for secret %s.%s: %s", s.GetName(), s.GetNamespace(), err.Error())
		}
	} else {

	}

	// Kubernetes cluster
	clusterName, config, err := m.kubeConverter.SecretToConfig(s)
	if err != nil {
		return false, KubeConfigInvalidFormatError(err, clusterName, s.GetName(), s.GetNamespace())
	}

	err = m.kubeConfigReceiver.ClusterAdded(config.RestConfig, clusterName)
	if err != nil {
		return true, ClusterAddError(err, clusterName)
	}

	logger.Infof("Adding new cluster member: %s", clusterName)
	m.lock.Lock()
	m.clusterByName[clusterName] = &remoteCluster{
		secretName:  s.GetName(),
		kubeContext: config.ApiConfig.CurrentContext,
	}
	m.lock.Unlock()

	m.lock.RLock()
	logger.Infof("Number of remote clusters: %d", len(m.clusterByName))
	m.lock.RUnlock()
	return false, nil
}

func (m *MeshAPIMembershipHandler) DeleteMemberCluster(ctx context.Context, s *v1.Secret) (resync bool, err error) {
	logger := contextutils.LoggerFrom(ctx)
	for clusterID, cluster := range m.clusters() {
		if cluster.secretName == s.GetName() {
			logger.Infof("Deleting cluster member: %s", clusterID)
			err := m.kubeConfigReceiver.ClusterRemoved(clusterID)
			if err != nil {
				return true, ClusterDeletionError(err, clusterID)
			}
			m.lock.Lock()
			delete(m.clusterByName, clusterID)
			m.lock.Unlock()
		}
	}
	m.lock.RLock()
	logger.Infof("Number of remote clusters: %d", len(m.clusterByName))
	m.lock.RUnlock()
	return false, nil
}

func (m *MeshAPIMembershipHandler) clusters() map[string]*remoteCluster {
	m.lock.RLock()
	defer m.lock.RUnlock()
	result := make(map[string]*remoteCluster)
	for k, v := range m.clusterByName {
		result[k] = v
	}
	return result
}
