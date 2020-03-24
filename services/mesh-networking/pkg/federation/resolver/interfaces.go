package resolver

import (
	"context"
)

// take federation decisions that have been written to mesh services and convert those decisions
// into concrete in-cluster resources to enable multicluster communication
type FederationResolver interface {
	Start(ctx context.Context) error
}
