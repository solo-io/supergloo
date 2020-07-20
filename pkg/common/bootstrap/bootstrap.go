package bootstrap

import (
	"context"

	"github.com/go-logr/zapr"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/service-mesh-hub/pkg/common/schemes"
	"github.com/solo-io/skv2/pkg/multicluster"
	"github.com/solo-io/skv2/pkg/multicluster/watch"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/log"
	zaputil "sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	// required import to enable kube client-go auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

type StartReconciler func(
	ctx context.Context,
	masterManager manager.Manager,
	mcClient multicluster.Client,
	clusters multicluster.ClusterSet,
	mcWatcher multicluster.ClusterWatcher,
) error

// bootstrap options for starting discovery
// TODO: wire these up to Settings CR
type Options struct {
	// MetricsBindAddress is the TCP address that the controller should bind to
	// for serving prometheus metrics.
	// It can be set to "0" to disable the metrics serving.
	MetricsBindAddress string

	// MasterNamespace if specified restricts the Master manager's cache to watch objects in
	// the desired namespace Defaults to all namespaces
	//
	// Note: If a namespace is specified, controllers can still Watch for a
	// cluster-scoped resource (e.g Node).  For namespaced resources the cache
	// will only hold objects from the desired namespace.
	MasterNamespace string

	// enables debug mode
	DebugMode bool
}

// the mesh-discovery controller is the Kubernetes Controller/Operator
// which processes k8s storage events to produce
// discovered resources.
func Start(ctx context.Context, rootLogger string, start StartReconciler, opts Options) error {
	setupLogging(opts.DebugMode)

	ctx = contextutils.WithLogger(ctx, rootLogger)
	mgr, err := makeMasterManager(opts)
	if err != nil {
		return err
	}

	clusterWatcher := watch.NewClusterWatcher(ctx, manager.Options{
		Namespace: "", // TODO (ilackarms): support configuring specific watch namespaces on remote clusters
		Scheme:    mgr.GetScheme(),
	})

	mcClient := multicluster.NewClient(clusterWatcher)

	if err := start(ctx, mgr, mcClient, clusterWatcher, clusterWatcher); err != nil {
		return err
	}

	if err := clusterWatcher.Run(mgr); err != nil {
		return err
	}

	contextutils.LoggerFrom(ctx).Infof("starting manager with options %+v", opts)
	return mgr.Start(ctx.Done())
}

// get the manager for the local cluster; we will use this as our "master" cluster
func makeMasterManager(opts Options) (manager.Manager, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	mgr, err := manager.New(cfg, manager.Options{
		Namespace:          opts.MasterNamespace, // TODO (ilackarms): support configuring multiple watch namespaces on master cluster
		MetricsBindAddress: opts.MetricsBindAddress,
	})
	if err != nil {
		return nil, err
	}

	if err := schemes.SchemeBuilder.AddToScheme(mgr.GetScheme()); err != nil {
		return nil, err
	}
	return mgr, nil
}

func setupLogging(debugMode bool) {
	level := zapcore.InfoLevel
	if debugMode {
		level = zapcore.DebugLevel
	}
	atomicLevel := zap.NewAtomicLevelAt(level)
	baseLogger := zaputil.NewRaw(
		zaputil.Level(&atomicLevel),
		// Only set debug mode if specified. This will use a non-json (human readable) encoder which makes it impossible
		// to use any json parsing tools for the log. Should only be enabled explicitly
		zaputil.UseDevMode(debugMode),
	)

	// klog
	zap.ReplaceGlobals(baseLogger)
	// controller-runtime
	log.SetLogger(zapr.NewLogger(baseLogger))
	// go-utils
	contextutils.SetFallbackLogger(baseLogger.Sugar())
}
