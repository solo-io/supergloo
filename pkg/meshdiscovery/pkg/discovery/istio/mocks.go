package istio

import (
	"context"

	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

type mockIstioMeshDiscovery struct {
}

func NewMockIstioMeshDiscovery() *mockIstioMeshDiscovery {
	return &mockIstioMeshDiscovery{}
}

func (imd *mockIstioMeshDiscovery) DiscoverMeshes(ctx context.Context, snapshot *v1.DiscoverySnapshot) (v1.MeshList, error) {
	return nil, nil
}
