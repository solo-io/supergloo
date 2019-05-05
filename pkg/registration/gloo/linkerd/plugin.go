package linkerd

import (
	"context"

	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

type glooLinkerdMtlsPlugin struct {
}

func (pl *glooLinkerdMtlsPlugin) HandleMeshes(ctx context.Context, meshes v1.MeshList) error {
	panic("implement me")
}
