package discovery

import (
	"context"

	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

type MeshDiscovery interface {
	DiscoverMeshes(ctx context.Context, snapshot *v1.DiscoverySnapshot) (v1.MeshList, error)
}
