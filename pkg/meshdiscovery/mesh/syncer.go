package mesh

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"go.uber.org/zap"
)

type meshDiscoverySyncer struct {
	meshClient     v1.MeshClient
	plugins        MeshDiscoveryPlugins
	meshReconciler v1.MeshReconciler
}

// calling this function with nil is valid and expected outside of tests
func NewMeshDiscoverySyncer(meshClient v1.MeshClient, plugins ...MeshDiscovery) v1.DiscoverySyncer {
	meshReconciler := v1.NewMeshReconciler(meshClient)
	return &meshDiscoverySyncer{
		meshClient:     meshClient,
		plugins:        plugins,
		meshReconciler: meshReconciler,
	}
}

func (s *meshDiscoverySyncer) Sync(ctx context.Context, snap *v1.DiscoverySnapshot) error {
	multierr := &multierror.Error{}
	ctx = contextutils.WithLogger(ctx, fmt.Sprintf("mesh-discovery-syncer-%v", snap.Hash()))
	fields := []interface{}{
		zap.Int("installs", len(snap.Installs.List())),
		zap.Int("pods", len(snap.Pods.List())),
	}
	logger := contextutils.LoggerFrom(ctx)
	logger.Infow("begin sync", fields...)
	defer logger.Infow("end sync", fields...)

	var discoveredMeshes v1.MeshList

	for _, meshDiscoveryPlugin := range s.plugins {
		meshes, err := meshDiscoveryPlugin.DiscoverMeshes(ctx, snap)
		if err != nil {
			multierr = multierror.Append(multierr, err)
			logger.Errorf(err.Error())
		}
		discoveredMeshes = append(discoveredMeshes, meshes...)
	}

	// reconcile all discovered meshes
	err := s.meshReconciler.Reconcile("", discoveredMeshes, func(original, desired *v1.Mesh) (b bool, e error) {
		return false, nil
	}, clients.ListOpts{})
	if err != nil {
		multierr = multierror.Append(multierr, err)
	}

	return multierr.ErrorOrNil()
}
