package internal_watcher

import (
	"context"

	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/contextutils"
	compute_target "github.com/solo-io/service-mesh-hub/pkg/common/compute-target"
	k8s_core_types "k8s.io/api/core/v1"
)

var (
	ComputeTargetAddError = func(err error, computeTargetID string) error {
		return eris.Wrapf(err, "error during compute target add handler for (%s)", computeTargetID)
	}
	ComputeTargetRemoveError = func(err error, computeTargetID string) error {
		return eris.Wrapf(err, "error during compute target delete handler for (%s)", computeTargetID)
	}
)

type ComputeTargetMembershipHandler struct {
	computeTargetCredentialsHandlers []compute_target.ComputeTargetCredentialsHandler
}

func NewComputeTargetMembershipHandler(
	computeTargetCredentialsHandlers []compute_target.ComputeTargetCredentialsHandler,
) *ComputeTargetMembershipHandler {
	return &ComputeTargetMembershipHandler{
		computeTargetCredentialsHandlers: computeTargetCredentialsHandlers,
	}
}

func (m *ComputeTargetMembershipHandler) ComputeTargetSecretAdded(ctx context.Context, s *k8s_core_types.Secret) (resync bool, err error) {
	logger := contextutils.LoggerFrom(ctx)
	logger.Infof("Adding new compute target with name: %s", s.GetName())
	for _, credsHandler := range m.computeTargetCredentialsHandlers {
		if err = credsHandler.ComputeTargetAdded(ctx, s); err != nil {
			break
		}
	}
	if err != nil {
		return false, ComputeTargetAddError(err, s.GetName())
	}
	return false, nil
}

func (m *ComputeTargetMembershipHandler) ComputeTargetSecretRemoved(ctx context.Context, s *k8s_core_types.Secret) (resync bool, err error) {
	logger := contextutils.LoggerFrom(ctx)
	logger.Infof("Removing compute target with name: %s", s.GetName())
	for _, credsHandler := range m.computeTargetCredentialsHandlers {
		if err = credsHandler.ComputeTargetRemoved(ctx, s); err != nil {
			break
		}
	}
	if err != nil {
		return false, ComputeTargetRemoveError(err, s.GetName())
	}
	return false, nil
}
