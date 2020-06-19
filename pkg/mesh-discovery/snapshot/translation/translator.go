package translation

import (
	"context"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/snapshots/output"
	"github.com/solo-io/smh/pkg/mesh-discovery/snapshot/input"
)

// the translator "reconciles the entire state of the world"
type Translator interface {
	// translates the Input Snapshot to an Output Snapshot
	Translate(ctx context.Context, in input.Snapshot) output.Snapshot
}
