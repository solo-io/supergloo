package internal_watcher

import (
	"context"

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
		kubeConverter:        kubeConverter,
	}
}

func (m *MeshPlatformMembershipHandler) AddMemberMeshPlatform(ctx context.Context, s *k8s_core_types.Secret) (resync bool, err error) {
	logger := contextutils.LoggerFrom(ctx)
	logger.Infof("Adding new mesh platform member: %s", s.GetName())
	// TODO change this check when migrating to skv2
	if s.Type == k8s_core_types.SecretTypeOpaque {
		// New k8s cluster mesh platform
		var clusterName string
		var config *kube.ConvertedConfigs
		clusterName, config, err = m.kubeConverter.SecretToConfig(s)
		if err != nil {
			return false, KubeConfigInvalidFormatError(err, clusterName, s.GetName(), s.GetNamespace())
		}
		err = m.kubeConfigReceiver.ClusterAdded(config.RestConfig, clusterName)
	} else {
		// New REST API mesh platform
		for _, restAPICredsHandler := range m.restAPICredsHandlers {
			err = restAPICredsHandler.RestAPIAdded(ctx, s)
			if err != nil {
				break
			}
		}
	}
	if err != nil {
		return false, PlatformAddError(err, s.GetName())
	}
	return false, nil
}

func (m *MeshPlatformMembershipHandler) DeleteMemberMeshPlatform(ctx context.Context, s *k8s_core_types.Secret) (resync bool, err error) {
	logger := contextutils.LoggerFrom(ctx)
	logger.Infof("Deleting mesh platform member: %s", s.GetName())
	// TODO change this check when migrating to skv2
	if s.Type == k8s_core_types.SecretTypeOpaque {
		err = m.kubeConfigReceiver.ClusterRemoved(s.GetName())
	} else {
		// New REST API mesh platform
		for _, restAPICredsHandler := range m.restAPICredsHandlers {
			err = restAPICredsHandler.RestAPIRemoved(ctx, s)
			if err != nil {
				break
			}
		}
	}
	if err != nil {
		return false, PlatformDeletionError(err, s.GetName())
	}
	return false, nil
}
