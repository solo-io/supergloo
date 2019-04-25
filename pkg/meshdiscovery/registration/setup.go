package registration

import (
	"context"
	"time"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/clientset"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/config/istio"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/config/linkerd"
	"github.com/solo-io/supergloo/pkg/registration"
)

func RunRegistrationEventLoop(ctx context.Context, cs *clientset.Clientset, customErrHandler func(error)) error {
	ctx = contextutils.WithLogger(ctx, "mesh-discovery-registration-event-loop")
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

	manager := registration.NewManager()
	newDiscoveryConfigLoops(ctx, cs, manager)

	registrationSyncers := createRegistrationSyncers(manager)

	if err := runRegistrationEventLoop(ctx, errHandler, cs, registrationSyncers); err != nil {
		return err
	}

	return nil
}

func newDiscoveryConfigLoops(ctx context.Context, clientset *clientset.Clientset, manager *registration.Manager) {
	istio.StartIstioDiscoveryConfigLoop(ctx, clientset, manager)
	linkerd.StartLinkerdDiscoveryConfigLoop(ctx, clientset, manager)
}

// Add registration syncers here
func createRegistrationSyncers(manager *registration.Manager) v1.RegistrationSyncer {
	return v1.RegistrationSyncers{
		registration.NewRegistrationSyncer(manager),
	}
}

// start the registration event loop
func runRegistrationEventLoop(ctx context.Context, errHandler func(err error), clientset *clientset.Clientset, syncers v1.RegistrationSyncer) error {
	registrationEmitter := v1.NewRegistrationEmitter(clientset.Discovery.Mesh, clientset.Input.MeshIngress)
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
