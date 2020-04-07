package config

import (
	"context"

	"github.com/go-logr/zapr"
	"sigs.k8s.io/controller-runtime/pkg/log"
	zaputil "sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/service-mesh-hub/pkg/version"
	"go.uber.org/zap"
)

func CreateRootContext(customCtx context.Context, name string) context.Context {
	setupLogging(name, "version", version.Version)

	rootCtx := customCtx
	if rootCtx == nil {
		rootCtx = context.Background()
	}

	return rootCtx
}

func setupLogging(name string, values ...interface{}) {
	// set up zap logger for all loggers
	logconfig := zap.NewDevelopmentConfig()
	logconfig.Level.SetLevel(zap.DebugLevel)

	baseLogger := zaputil.NewRaw(
		zaputil.Level(&logconfig.Level),
		zaputil.UseDevMode(true),
	).Named(name).Sugar().With(values).Desugar()

	// klog
	zap.ReplaceGlobals(baseLogger)
	// controller-runtime
	log.SetLogger(zapr.NewLogger(baseLogger))
	// go-utils
	contextutils.SetFallbackLogger(baseLogger.Sugar())
}
