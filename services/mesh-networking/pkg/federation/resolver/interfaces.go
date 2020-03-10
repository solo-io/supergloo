package resolver

import (
	"context"

	"github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1/controller"
)

// take federation decisions that have been written to mesh services and convert those decisions
// into concrete in-cluster resources to enable multicluster communication
type FederationResolver interface {
	Start(ctx context.Context, meshServiceController controller.MeshServiceController)
}
