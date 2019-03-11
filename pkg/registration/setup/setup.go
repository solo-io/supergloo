package setup

import (
	"context"
	"time"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/supergloo/pkg/api/clientset"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/registration/istio"
)

func RunRegistrationEventLoop(ctx context.Context, cs *clientset.Clientset, customErrHandler func(error)) error {
	ctx = contextutils.WithLogger(ctx, "registration-event-loop")
	logger := contextutils.LoggerFrom(ctx)

	errHandler := func(err error) {
		if err == nil {
			return
		}
		logger.Errorf("config error: %v", err)
		if customErrHandler != nil {
			customErrHandler(err)
		}
	}

	configSyncers := createRegistrationSyncers(cs, errHandler)

	if err := runRegistrationEventLoop(ctx, errHandler, cs, configSyncers); err != nil {
		return err
	}

	return nil
}

// Add registration syncers here
func createRegistrationSyncers(clientset *clientset.Clientset, errHandler func(error)) v1.RegistrationSyncer {
	return v1.RegistrationSyncers{
		istio.NewIstioRegistrationSyncer(clientset, errHandler),
	}
}

// start the istio config event loop
func runRegistrationEventLoop(ctx context.Context, errHandler func(err error), clientset *clientset.Clientset, syncers v1.RegistrationSyncer) error {
	configEmitter := v1.NewRegistrationEmitter(clientset.Input.Mesh)
	configEventLoop := v1.NewRegistrationEventLoop(configEmitter, syncers)

	watchOpts := clients.WatchOpts{
		Ctx:         ctx,
		RefreshRate: time.Second * 1,
	}

	configEventLoopErrs, err := configEventLoop.Run(nil, watchOpts)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case err := <-configEventLoopErrs:
				errHandler(err)
			case <-ctx.Done():
			}
		}
	}()
	return nil
}
