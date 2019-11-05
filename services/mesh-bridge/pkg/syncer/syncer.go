package syncer

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/mesh-projects/pkg/api/external/istio/networking/v1alpha3"
	v1 "github.com/solo-io/mesh-projects/pkg/api/v1"
	"github.com/solo-io/mesh-projects/services/mesh-bridge/pkg/setup/config"
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
	clients                config.ClientSet
	mbTranslator           translator.Translator
	serviceEntryReconciler v1alpha3.ServiceEntryReconciler
}

func NewMeshBridgeSyncer(clients config.ClientSet, mbTranslator translator.Translator,
	serviceEntryReconciler v1alpha3.ServiceEntryReconciler) v1.NetworkBridgeSyncer {
	return &networkBridgeSyncer{
		clients:                clients,
		mbTranslator:           mbTranslator,
		serviceEntryReconciler: serviceEntryReconciler,
	}
}

func (s *networkBridgeSyncer) Sync(ctx context.Context, snapshot *v1.NetworkBridgeSnapshot) error {
	ctx = contextutils.WithLoggerValues(ctx, zap.String("syncer", "operator"))
	contextutils.LoggerFrom(ctx).Infow("snapshot resources", zap.Int("mesh bridges", len(snapshot.MeshBridges)))
	var multiErr *multierror.Error

	serviceEntriesByNamespace, err := s.mbTranslator.Translate(ctx, s.getMeshBridgesByNamespace(snapshot.MeshBridges))
	if err != nil {
		return err
	}

	for namespace, serviceEntries := range serviceEntriesByNamespace {
		err = s.serviceEntryReconciler.Reconcile(namespace, serviceEntries, nil, clients.ListOpts{})
		multiErr = multierror.Append(multiErr, err)
	}

	return multiErr.ErrorOrNil()
}

func (s *networkBridgeSyncer) getMeshBridgesByNamespace(meshBridges v1.MeshBridgeList) translator.MeshBridgesByNamespace {

	meshBridgeMap := make(translator.MeshBridgesByNamespace)
	for _, v := range meshBridges {
		meshBridgeMap[v.Metadata.GetNamespace()] = append(meshBridgeMap[v.Metadata.GetNamespace()], v)
	}
	return meshBridgeMap
}
