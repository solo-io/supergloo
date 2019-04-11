package discovery

import v1 "github.com/solo-io/supergloo/pkg/api/v1"

type MeshDiscovery interface {
	DiscoverMeshes() (v1.MeshList, error)
}
