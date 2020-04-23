package internal_watcher

import (
	"context"
	"sync"

	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/kube"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/multicluster/manager"
	rest_api "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/rest-api"
	k8s_core_types "k8s.io/api/core/v1"
)

//go:generate mockgen -source ./membership.go -destination mocks/membership.go

var (
	PlatformAddError = func(err error, platformId string) error {
		return eris.Wrapf(err, "error during platform add handler for (%s)", platformId)
	}
	PlatformDeletionError = func(err error, platformId string) error {
		return eris.Wrapf(err, "error during platform delete handler for (%s)", platformId)
	}
	KubeConfigInvalidFormatError = func(err error, platform, name, namespace string) error {
		return eris.Wrapf(err, "invalid kube config for cluster %s in the secret %s in namespace %s",
			platform, name, namespace)
	}
)

// this interface is meant to abstract the platform add/delete logic for the secret watcher
type MeshPlatformSecretHandler interface {
	AddMemberMeshPlatform(ctx context.Context, s *k8s_core_types.Secret) (resync bool, err error)
	DeleteMemberMeshPlatform(ctx context.Context, s *k8s_core_types.Secret) (resync bool, err error)
}

type MeshPlatformMembershipHandler struct {
	kubeConfigReceiver   mc_manager.KubeConfigHandler
	restAPICredsHandlers []rest_api.RestAPICredsHandler
	lock                 sync.RWMutex
	platformByName       map[string]*remotePlatform
	kubeConverter        kube.Converter
}

func NewMeshPlatformMembershipHandler(
	kubeConfigReceiver mc_manager.KubeConfigHandler,
	restAPICredsHandlers []rest_api.RestAPICredsHandler,
	kubeConverter kube.Converter,
) *MeshPlatformMembershipHandler {
	return &MeshPlatformMembershipHandler{
		kubeConfigReceiver:   kubeConfigReceiver,
		restAPICredsHandlers: restAPICredsHandlers,
		platformByName:       make(map[string]*remotePlatform),
		kubeConverter:        kubeConverter,
	}
}

type remotePlatform struct {
	secretName  string
	kubeContext string
}

func (m *MeshPlatformMembershipHandler) AddMemberMeshPlatform(ctx context.Context, s *k8s_core_types.Secret) (resync bool, err error) {
	logger := contextutils.LoggerFrom(ctx)
	// TODO change this check when migrating to skv2
	if s.Type == k8s_core_types.SecretTypeOpaque {
		// New k8s cluster mesh platform
		clusterName, config, err := m.kubeConverter.SecretToConfig(s)
		if err != nil {
			return false, KubeConfigInvalidFormatError(err, clusterName, s.GetName(), s.GetNamespace())
		}

		err = m.kubeConfigReceiver.ClusterAdded(config.RestConfig, clusterName)
		if err != nil {
			return true, PlatformAddError(err, clusterName)
		}

		logger.Infof("Adding new cluster member: %s", clusterName)
		m.lock.Lock()
		m.platformByName[clusterName] = &remotePlatform{
			secretName:  s.GetName(),
			kubeContext: config.ApiConfig.CurrentContext,
		}
		m.lock.Unlock()
		m.lock.RLock()
		logger.Infof("Number of remote clusters: %d", len(m.platformByName))
		m.lock.RUnlock()
	} else {
		// New REST API mesh platform
		for _, restAPICredsHandler := range m.restAPICredsHandlers {
			err := restAPICredsHandler.RestAPIAdded(ctx, s)
			if err != nil {
				logger.Errorf(
					"Error initializing RestAPICredsHandler for secret %s.%s: %s",
					s.GetName(),
					s.GetNamespace(),
					err.Error())
			}
		}
	}
	return false, nil
}

func (m *MeshPlatformMembershipHandler) DeleteMemberMeshPlatform(ctx context.Context, s *k8s_core_types.Secret) (resync bool, err error) {
	logger := contextutils.LoggerFrom(ctx)
	for platformID, platform := range m.platforms() {
		if platform.secretName == s.GetName() {
			logger.Infof("Deleting platform member: %s", platformID)
			err := m.kubeConfigReceiver.ClusterRemoved(platformID)
			if err != nil {
				return true, PlatformDeletionError(err, platformID)
			}
			m.lock.Lock()
			delete(m.platformByName, platformID)
			m.lock.Unlock()
		}
	}
	m.lock.RLock()
	logger.Infof("Number of remote platforms: %d", len(m.platformByName))
	m.lock.RUnlock()
	return false, nil
}

// TODO REMOVE THIS
func (m *MeshPlatformMembershipHandler) platforms() map[string]*remotePlatform {
	m.lock.RLock()
	defer m.lock.RUnlock()
	result := make(map[string]*remotePlatform)
	for k, v := range m.platformByName {
		result[k] = v
	}
	return result
}
