package config

import (
	"context"
)

type MeshConfigDiscovery interface {
	Run(ctx context.Context) (<-chan error, error)
	HandleError(ctx context.Context, err error)
}
