package plugins

import (
	"context"

	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/webhook/patch"
	corev1 "k8s.io/api/core/v1"
)

type InjectionPlugin interface {
	Name() string
	GetAutoInjectMeshes(ctx context.Context) ([]*v1.Mesh, error)
	CheckMatch(ctx context.Context, candidatePod *corev1.Pod, meshes []*v1.Mesh) ([]*v1.Mesh, error)
	GetSidecarPatch(ctx context.Context, candidatePod *corev1.Pod, mesh []*v1.Mesh) ([]patch.JSONPatchOperation, error)
}

// Add new plugins here
func GetPlugins() []InjectionPlugin {
	return []InjectionPlugin{
		AppMeshInjectionPlugin{},
	}
}
