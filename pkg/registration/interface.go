package registration

import (
	"context"

	"github.com/solo-io/solo-kit/pkg/api/v1/eventloop"
)

type EnabledConfigLoops struct {
	Istio   bool
	Gloo    bool
	AppMesh bool
}

type ConfigLoopStarters []ConfigLoopStarter
type ConfigLoopStarter func(ctx context.Context, enabled EnabledConfigLoops) (eventloop.EventLoop, error)
