package syncer

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/mesh-projects/pkg/api/external/istio/networking/v1alpha3"
	v1 "github.com/solo-io/mesh-projects/pkg/api/v1"
	"github.com/solo-io/mesh-projects/services/common"
	"github.com/solo-io/mesh-projects/services/internal/config"
	"github.com/solo-io/mesh-projects/services/mesh-bridge/pkg/translator"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
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
	clients      config.MeshBridgeClientSet
	mbTranslator translator.Translator
}

func NewMeshBridgeSyncer(clients config.MeshBridgeClientSet, mbTranslator translator.Translator) v1.NetworkBridgeSyncer {
	return &networkBridgeSyncer{
		clients:      clients,
		mbTranslator: mbTranslator,
	}
}

func (s *networkBridgeSyncer) Sync(ctx context.Context, snapshot *v1.NetworkBridgeSnapshot) error {
	ctx = contextutils.WithLoggerValues(ctx, zap.String("syncer", "operator"))
	contextutils.LoggerFrom(ctx).Infow("snapshot resources", zap.Int("mesh bridges", len(snapshot.MeshBridges)))
	// Lazy load this reconciler because it requires an istio client, which we do not want to load until after
	// the CRDs are registered
	serviceEntryReconciler := v1alpha3.NewServiceEntryReconciler(s.clients.ServiceEntry())

	var multiErr *multierror.Error

	serviceEntries, err := s.mbTranslator.Translate(ctx, snapshot)
	if err != nil {
		return err
	}

	err = serviceEntryReconciler.Reconcile("", serviceEntries, nil, clients.ListOpts{
		Selector: common.OwnerLabels,
	})
	multiErr = multierror.Append(multiErr, err)

	return multiErr.ErrorOrNil()
}
