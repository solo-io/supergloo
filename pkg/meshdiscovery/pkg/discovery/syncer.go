package discovery

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/solo-io/go-utils/contextutils"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/pkg/clientset"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/pkg/config"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/pkg/config/common"
	"go.uber.org/zap"
)

type meshDiscoverySyncer struct {
	rootCtx        context.Context
	cs             *clientset.Clientset
	plugins        MeshDiscoveryPlugins
	meshReconciler v1.MeshReconciler
}

// calling this function with nil is valid and expected outside of tests
func NewMeshDiscoverySyncer(ctx context.Context, cs *clientset.Clientset, plugins ...MeshDiscovery) v1.DiscoverySyncer {
	meshReconciler := v1.NewMeshReconciler(cs.Discovery.Mesh)
	return &meshDiscoverySyncer{
		rootCtx:        ctx,
		cs:             cs,
		plugins:        plugins,
		meshReconciler: meshReconciler,
	}
}

func (s *meshDiscoverySyncer) Sync(ctx context.Context, snap *v1.DiscoverySnapshot) error {
	multierr := &multierror.Error{}
	ctx = contextutils.WithLogger(ctx, fmt.Sprintf("mesh-discovery-syncer-%v", snap.Hash()))
	fields := []interface{}{
		zap.Int("meshes", len(snap.Meshes.List())),
		zap.Int("installs", len(snap.Installs.List())),
		zap.Int("pods", len(snap.Pods.List())),
	}
	logger := contextutils.LoggerFrom(ctx)
	logger.Infow("begin sync", fields...)
	defer logger.Infow("end sync", fields...)

	enabled := &common.EnabledConfigLoops{}

	for _, meshDiscoveryPlugin := range s.plugins {
		err := meshDiscoveryPlugin.DiscoverMeshes(ctx, snap, enabled)
		if err != nil {
			multierr = multierror.Append(multierr, err)
			logger.Errorf(err.Error())
		}
	}

	if multierr.ErrorOrNil() != nil {
		return multierr.ErrorOrNil()
	}

	return config.RunAdvancedDiscoverySyncers(ctx, s.cs, enabled)
}
