package config

import (
	"context"
)

type AdvancedMeshDiscovery interface {
	Run(ctx context.Context) (<-chan error, error)
	HandleError(ctx context.Context, err error)
}
