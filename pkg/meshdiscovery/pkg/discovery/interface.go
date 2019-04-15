package discovery

import (
	"context"

	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/pkg/config/common"
)

type MeshDiscoveryPlugins []MeshDiscovery

type MeshDiscovery interface {
	DiscoverMeshes(ctx context.Context, snapshot *v1.DiscoverySnapshot, enabled *common.EnabledConfigLoops) error
}
