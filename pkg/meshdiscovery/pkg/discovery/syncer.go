package discovery

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/pkg/discovery/istio"
)

type meshDiscoverySyncer struct {
	meshClient v1.MeshClient
	reporter   reporter.Reporter
}

// calling this function with nil is valid and expected outside of tests
func NewMeshDiscoverySyncer(meshClient v1.MeshClient, reporter reporter.Reporter) v1.DiscoverySyncer {
	return &meshDiscoverySyncer{
		meshClient: meshClient,
		reporter:   reporter,
	}
}

func (s *meshDiscoverySyncer) Sync(ctx context.Context, snap *v1.DiscoverySnapshot) error {
	multierr := &multierror.Error{}

	ctx = contextutils.WithLogger(ctx, fmt.Sprintf("mesh-discovery-syncer-%v", snap.Hash()))
	logger := contextutils.LoggerFrom(ctx)
	logger.Infof("begin sync %v", snap.Stringer())
	defer logger.Infof("end sync %v", snap.Stringer())

	var meshDiscoveryPlugins []MeshDiscovery
	var discoveredMeshes v1.MeshList

	pods := snap.Pods.List()
	meshes := snap.Meshes.List()

	meshDiscoveryPlugins = append(meshDiscoveryPlugins, istio.NewIstioMeshDiscovery(ctx, pods, meshes))

	for _, meshDiscoveryPlugin := range meshDiscoveryPlugins {
		meshes, err := meshDiscoveryPlugin.DiscoverMeshes()
		if err != nil {
			multierr = multierror.Append(multierr, err)
			logger.Errorf(err.Error())
		}
		discoveredMeshes = append(discoveredMeshes, meshes...)
	}

	// reconcile all discovered meshes

	return multierr.ErrorOrNil()

}
