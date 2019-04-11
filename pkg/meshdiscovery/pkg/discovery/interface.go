package discovery

import (
	"context"

	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

//go:generate mockgen -destination mocks_test.go  -source interface.go -package discovery

type MeshDiscovery interface {
	DiscoverMeshes() (v1.MeshList, error)
	Init(ctx context.Context, snapshot *v1.DiscoverySnapshot)
}
