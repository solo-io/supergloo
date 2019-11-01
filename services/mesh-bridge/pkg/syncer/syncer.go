package syncer

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/solo-io/go-utils/contextutils"
	v1 "github.com/solo-io/mesh-projects/pkg/api/v1"
	"github.com/solo-io/mesh-projects/services/mesh-bridge/pkg/setup/config"
	"github.com/solo-io/mesh-projects/services/mesh-bridge/pkg/translator"
	"go.uber.org/zap"
)

var (
	FailedToReconcileResourcesError = func(err error) error {
		return errors.Wrapf(err, "error reconciling resources")
	}

	FailedToPurgeResourcesError = func(err error) error {
		return errors.Wrapf(err, "error purging resources")
	}

	ResourcesNotAvailableDescription = "Resources failed to become available!"
)

type networkBridgeSyncer struct {
	clients      *config.ClientSet
	mbTranslator translator.Translator
}

func NewMeshBridgeSyncer(clients *config.ClientSet,
	mbTranslator translator.Translator) v1.NetworkBridgeSyncer {
	return &networkBridgeSyncer{
		clients:      clients,
		mbTranslator: mbTranslator,
	}
}

func (s *networkBridgeSyncer) Sync(ctx context.Context, snapshot *v1.NetworkBridgeSnapshot) error {
	ctx = contextutils.WithLoggerValues(ctx, zap.String("syncer", "operator"))
	contextutils.LoggerFrom(ctx).Infow("snapshot resources", zap.Int("mesh bridges", len(snapshot.MeshBridges)))
	var multiErr *multierror.Error

	return multiErr.ErrorOrNil()
}
