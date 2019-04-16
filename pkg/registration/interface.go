package registration

import (
	"context"
)

type EnabledConfigLoops struct {
	Istio   bool
	Gloo    bool
	AppMesh bool
}

type ConfigLoop interface {
	Run(ctx context.Context, enabled EnabledConfigLoops) error
}
