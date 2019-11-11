package main

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/mesh-projects/pkg/version"
	"github.com/solo-io/mesh-projects/services/internal/config"
	"github.com/solo-io/mesh-projects/services/mesh-bridge/pkg/setup"
	"go.uber.org/zap"
)

func main() {
	ctx := config.CreateRootContext(nil, "mesh-bridge")
	if err := run(ctx); err != nil {
		contextutils.LoggerFrom(ctx).Fatalw("err in main",
			zap.Error(err),
			zap.Any("process", version.MeshBridgeAppName))
	}
}

func run(ctx context.Context) error {
	errs := make(chan error)
	go func() {
		errs <- setup.Main(ctx, nil)
	}()
	return <-errs
}
