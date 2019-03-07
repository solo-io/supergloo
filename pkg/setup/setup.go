package setup

import (
	"context"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/go-utils/stats"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/install/istio"
)

func Main(customErrHandler func(error)) error {
	if os.Getenv("START_STATS_SERVER") != "" {
		stats.StartStatsServer()
	}

	rootCtx := createRootContext()
	logger := contextutils.LoggerFrom(rootCtx)

	errHandler := func(err error) {
		if err == nil {
			return
		}
		logger.Errorf("install event loop error: %v", err)
		if customErrHandler != nil {
			customErrHandler(err)
		}
	}

	clients, err := createClients(rootCtx)
	if err != nil {
		return errors.Wrap(err, "initializing clients")
	}

	return runInstallEventLoop(rootCtx, errHandler, clients, createInstallSyncers(clients.InstallClient, clients.MeshClient))
}

func createRootContext() context.Context {
	rootCtx := context.Background()
	rootCtx = contextutils.WithLogger(rootCtx, "supergloo")
	return rootCtx
}

type clientset struct {
	InstallClient v1.InstallClient
	MeshClient    v1.MeshClient
}

func createClients(ctx context.Context) (*clientset, error) {
	kubeCache := kube.NewKubeCache(ctx)
	restConfig, err := kubeutils.GetConfig("", "")
	if err != nil {
		return nil, err
	}

	installClient, err := v1.NewInstallClient(&factory.KubeResourceClientFactory{
		Crd:         v1.InstallCrd,
		Cfg:         restConfig,
		SharedCache: kubeCache,
	})
	if err != nil {
		return nil, err
	}
	if err := installClient.Register(); err != nil {
		return nil, err
	}

	meshClient, err := v1.NewMeshClient(&factory.KubeResourceClientFactory{
		Crd:         v1.MeshCrd,
		Cfg:         restConfig,
		SharedCache: kubeCache,
	})
	if err != nil {
		return nil, err
	}
	if err := meshClient.Register(); err != nil {
		return nil, err
	}

	return &clientset{
		InstallClient: installClient,
		MeshClient:    meshClient,
	}, nil
}

func runInstallEventLoop(ctx context.Context, errHandler func(err error), c *clientset, syncers v1.InstallSyncers) error {
	installEmitter := v1.NewInstallEmitter(c.InstallClient)
	installEventLoop := v1.NewInstallEventLoop(installEmitter, syncers)

	watchOpts := clients.WatchOpts{
		Ctx:         ctx,
		RefreshRate: time.Second * 1,
	}

	installEventLoopErrs, err := installEventLoop.Run(nil, watchOpts)
	if err != nil {
		return err
	}
	for {
		select {
		case err := <-installEventLoopErrs:
			errHandler(err)
		case <-ctx.Done():
			return nil
		}
	}
}

// Add install syncers here
func createInstallSyncers(installClient v1.InstallClient, meshClient v1.MeshClient) v1.InstallSyncers {
	return v1.InstallSyncers{
		istio.NewInstallSyncer(nil,
			meshClient,
			reporter.NewReporter("istio-install-reporter", installClient.BaseClient())),
	}
}
