package discovery

import (
	"context"
	"fmt"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

type meshDiscoverySyncer struct {
	meshClient v1.MeshClient
	reporter   reporter.Reporter
}

// calling this function with nil is valid and expected outside of tests
func NewMeshDiscoverySyncer(meshClient v1.MeshClient, reporter reporter.Reporter) v1.MeshdiscoverySyncer {
	return &meshDiscoverySyncer{
		meshClient: meshClient,
		reporter:   reporter,
	}
}

func (s *meshDiscoverySyncer) Sync(ctx context.Context, snap *v1.MeshdiscoverySnapshot) error {
	ctx = contextutils.WithLogger(ctx, fmt.Sprintf("istio-install-syncer-%v", snap.Hash()))
	logger := contextutils.LoggerFrom(ctx)
	logger.Infof("begin sync %v", snap.Stringer())
	defer logger.Infof("end sync %v", snap.Stringer())
	resourceErrs := make(reporter.ResourceErrors)

	// reporter should handle updates to the installs that happened during ensure
	return s.reporter.WriteReports(ctx, resourceErrs, nil)

}
