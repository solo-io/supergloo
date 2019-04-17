package registration

import (
	"context"

	"github.com/solo-io/supergloo/pkg/meshdiscovery/pkg/config"
)

type EnabledConfigLoops struct {
	Istio   bool
	Gloo    bool
	AppMesh bool
}

type ConfigLoopStarters []ConfigLoopStarter
type ConfigLoopStarter func(ctx context.Context, enabled EnabledConfigLoops) (config.EventLoop, error)
