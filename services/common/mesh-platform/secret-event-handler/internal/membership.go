package internal_watcher

import (
	"context"

	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/contextutils"
	mesh_platform "github.com/solo-io/service-mesh-hub/services/common/mesh-platform"
	k8s_core_types "k8s.io/api/core/v1"
)

var (
	PlatformAddError = func(err error, platformId string) error {
		return eris.Wrapf(err, "error during platform add handler for (%s)", platformId)
	}
	PlatformRemoveError = func(err error, platformId string) error {
		return eris.Wrapf(err, "error during platform delete handler for (%s)", platformId)
	}
)

type MeshPlatformMembershipHandler struct {
	meshPlatformCredentialsHandlers []mesh_platform.MeshPlatformCredentialsHandler
}

func NewMeshPlatformMembershipHandler(
	meshPlatformCredentialsHandlers []mesh_platform.MeshPlatformCredentialsHandler,
) *MeshPlatformMembershipHandler {
	return &MeshPlatformMembershipHandler{
		meshPlatformCredentialsHandlers: meshPlatformCredentialsHandlers,
	}
}

func (m *MeshPlatformMembershipHandler) MeshPlatformSecretAdded(ctx context.Context, s *k8s_core_types.Secret) (resync bool, err error) {
	logger := contextutils.LoggerFrom(ctx)
	logger.Infof("Adding new mesh platform with name: %s", s.GetName())
	for _, credsHandler := range m.meshPlatformCredentialsHandlers {
		if err = credsHandler.MeshPlatformAdded(ctx, s); err != nil {
			break
		}
	}
	if err != nil {
		return false, PlatformAddError(err, s.GetName())
	}
	return false, nil
}

func (m *MeshPlatformMembershipHandler) MeshPlatformSecretRemoved(ctx context.Context, s *k8s_core_types.Secret) (resync bool, err error) {
	logger := contextutils.LoggerFrom(ctx)
	logger.Infof("Removing mesh platform with name: %s", s.GetName())
	for _, credsHandler := range m.meshPlatformCredentialsHandlers {
		if err = credsHandler.MeshPlatformRemoved(ctx, s); err != nil {
			break
		}
	}
	if err != nil {
		return false, PlatformRemoveError(err, s.GetName())
	}
	return false, nil
}
