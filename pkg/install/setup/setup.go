package setup

import (
	"context"
	"time"

	"github.com/solo-io/supergloo/pkg/api/clientset"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	istioinstall "github.com/solo-io/supergloo/pkg/install/istio"
)

func RunInstallEventLoop(ctx context.Context, cs *clientset.Clientset, customErrHandler func(error)) error {
	ctx = contextutils.WithLogger(ctx, "install-event-loop")
	logger := contextutils.LoggerFrom(ctx)

	errHandler := func(err error) {
		if err == nil {
			return
		}
		logger.Errorf("install error: %v", err)
		if customErrHandler != nil {
			customErrHandler(err)
		}
	}

	installSyncers := createInstallSyncers(cs, errHandler)

	if err := startEventLoop(ctx, errHandler, cs, installSyncers); err != nil {
		return err
	}

	return nil
}

// Add install syncers here
func createInstallSyncers(clientset *clientset.Clientset, errHandler func(error)) v1.InstallSyncers {
	return v1.InstallSyncers{
		istioinstall.NewInstallSyncer(
			nil,
			clientset.Input.Mesh,
			reporter.NewReporter("istio-install-reporter", clientset.Input.Install.BaseClient()),
		),
	}
}

// start the install event loop
func startEventLoop(ctx context.Context, errHandler func(err error), c *clientset.Clientset, syncers v1.InstallSyncers) error {
	installEmitter := v1.NewInstallEmitter(c.Input.Install)
	installEventLoop := v1.NewInstallEventLoop(installEmitter, syncers)

	watchOpts := clients.WatchOpts{
		Ctx:         ctx,
		RefreshRate: time.Second * 1,
	}

	installEventLoopErrs, err := installEventLoop.Run(nil, watchOpts)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case err := <-installEventLoopErrs:
				errHandler(err)
			case <-ctx.Done():
			}
		}
	}()
	return nil
}
