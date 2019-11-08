package main

import (
	"context"

	"go.uber.org/zap/zapcore"

	"github.com/solo-io/mesh-projects/services/mesh-config/pkg/setup"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/mesh-projects/pkg/version"
	"go.uber.org/zap"
)

func main() {

	ctx := createRootContext()
	contextutils.SetLogLevel(zapcore.DebugLevel)
	if err := run(ctx); err != nil {
		contextutils.LoggerFrom(ctx).Fatalw("err in main",
			zap.Error(err),
			zap.Any("process", version.MeshConfigAppName))
	}
}

func run(ctx context.Context) error {
	errs := make(chan error)
	go func() {
		errs <- setup.Main(ctx, nil)
	}()
	return <-errs
}

func createRootContext() context.Context {
	rootCtx := context.Background()
	rootCtx = contextutils.WithLogger(rootCtx, version.MeshConfigAppName)
	loggingContext := []interface{}{"version", version.Version}
	rootCtx = contextutils.WithLoggerValues(rootCtx, loggingContext...)
	return rootCtx
}
