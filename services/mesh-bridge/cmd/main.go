package main

import (
	"context"

	"github.com/solo-io/mesh-projects/pkg/version"
	setup2 "github.com/solo-io/mesh-projects/services/mesh-bridge/pkg/setup"

	"github.com/solo-io/go-utils/contextutils"
	"go.uber.org/zap"
)

func main() {
	ctx := getInitialContext()
	// get loop set
	loops, err := setup2.MustInitializeMeshBridge(ctx)
	if err != nil {
		contextutils.LoggerFrom(ctx).Fatalw("Failed to initialize operator event loops", zap.Error(err))
	}

	// run mesh bridge loop
	contextutils.LoggerFrom(ctx).Infow("Starting event loop",
		zap.Any("watchOpts", loops.WatchOpts),
		zap.Strings("watchNamespaces", loops.WatchNamespaces))
	errs, err := loops.OperatorLoop.Run(loops.WatchNamespaces, loops.WatchOpts)
	if err != nil {
		contextutils.LoggerFrom(ctx).Fatalw("Failed to start event loop", zap.Error(err))
	}
	for err := range errs {
		contextutils.LoggerFrom(ctx).Errorw("Error during event loop", zap.Error(err))
	}
}

func getInitialContext() context.Context {
	loggingContext := []interface{}{"version", version.Version}
	ctx := contextutils.WithLogger(context.Background(), "sm-marketplace-operator")
	ctx = contextutils.WithLoggerValues(ctx, loggingContext...)
	return ctx
}
