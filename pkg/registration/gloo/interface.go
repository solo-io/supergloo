package gloo

import (
	"context"

	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

//go:generate mockgen -destination=./mocks/interface.go -source interface.go -package mocks

type GlooIngressPlugins []GlooIngressPlugin
type GlooIngressPlugin interface {
	HandleMeshes(ctx context.Context, ingress *v1.MeshIngress, meshes v1.MeshList) error
}
