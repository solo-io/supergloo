package setup

import (
	"context"
	"time"

	"github.com/solo-io/supergloo/pkg/registration"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/supergloo/pkg/api/clientset"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

func RunRegistrationEventLoop(ctx context.Context, cs *clientset.Clientset, customErrHandler func(error)) error {
	ctx = contextutils.WithLogger(ctx, "registration-event-loop")
	logger := contextutils.LoggerFrom(ctx)

	errHandler := func(err error) {
		if err == nil {
			return
		}
		logger.Errorf("registration error: %v", err)
		if customErrHandler != nil {
			customErrHandler(err)
		}
	}

	registrationSyncers := createRegistrationSyncers(cs, errHandler)

	if err := runRegistrationEventLoop(ctx, errHandler, cs, registrationSyncers); err != nil {
		return err
	}

	return nil
}

// Add registration syncers here
func createRegistrationSyncers(clientset *clientset.Clientset, errHandler func(error)) v1.RegistrationSyncer {
	return v1.RegistrationSyncers{
		registration.NewRegistrationSyncer(clientset, errHandler),
	}
}

// start the registration event loop
func runRegistrationEventLoop(ctx context.Context, errHandler func(err error), clientset *clientset.Clientset, syncers v1.RegistrationSyncer) error {
	registrationEmitter := v1.NewRegistrationEmitter(clientset.Input.Mesh)
	registrationEventLoop := v1.NewRegistrationEventLoop(registrationEmitter, syncers)

	watchOpts := clients.WatchOpts{
		Ctx:         ctx,
		RefreshRate: time.Second * 1,
	}

	registrationEventLoopErrs, err := registrationEventLoop.Run(nil, watchOpts)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case err := <-registrationEventLoopErrs:
				errHandler(err)
			case <-ctx.Done():
			}
		}
	}()
	return nil
}
