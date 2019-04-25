package registration

import (
	"context"

	"github.com/solo-io/solo-kit/pkg/api/v1/eventloop"
)

//go:generate mockgen -destination mocks/mocks.go github.com/solo-io/supergloo/pkg/registration ConfigLoop

type ConfigLoop interface {
	Enabled(enabled EnabledConfigLoops) bool
	Start(ctx context.Context, enabled EnabledConfigLoops) (eventloop.EventLoop, error)
}

type EnabledConfigLoops struct {
	Istio   bool
	Gloo    bool
	AppMesh bool
	Linkerd bool
}

type ConfigLoopStarters []ConfigLoopStarter
type ConfigLoopStarter func(ctx context.Context, enabled EnabledConfigLoops) (eventloop.EventLoop, error)
