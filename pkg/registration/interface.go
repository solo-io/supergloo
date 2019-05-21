package registration

import (
	"context"

	"github.com/solo-io/solo-kit/pkg/api/v1/eventloop"
)

type ConfigLoop interface {
	Enabled(enabled EnabledConfigLoops) bool
	Start(ctx context.Context, enabled EnabledConfigLoops) (eventloop.EventLoop, error)
}

type EnabledConfigLoops struct {
	Istio    bool
	IstioSmi bool
	Gloo     bool
	AppMesh  bool
	Linkerd  bool
}

type ConfigLoopStarter func(ctx context.Context, enabled EnabledConfigLoops) (eventloop.EventLoop, error)
