package config

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/mesh-projects/pkg/version"
	"go.uber.org/zap"
)

func CreateRootContext(customCtx context.Context, name string) context.Context {
	rootCtx := customCtx
	if rootCtx == nil {
		rootCtx = context.Background()
	}
	rootCtx = contextutils.WithLogger(rootCtx, name)
	loggingContext := []interface{}{"version", version.Version}
	rootCtx = contextutils.WithLoggerValues(rootCtx, loggingContext...)

	// >:(
	// the default global zap logger, which controller-runtime uses, is a no-op logger
	// https://github.com/uber-go/zap/blob/5dab9368974ab1352e4245f9d33e5bce4c23a034/global.go#L41
	zap.ReplaceGlobals(contextutils.LoggerFrom(rootCtx).Desugar())

	return rootCtx
}
