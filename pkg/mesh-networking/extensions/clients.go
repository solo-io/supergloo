package extensions

import (
	"context"
	"github.com/solo-io/go-utils/hashutils"
	"sync"

	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/extensions/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/settings.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/common/utils/grpc"
)

//go:generate mockgen -source ./clients.go -destination mocks/mock_clients.go

// A PushFunc handles push notifications from an Extensions Server
type PushFunc func(notification *v1alpha1.PushNotification)

// Clients provides a convenience wrapper for a set of clients to communicate with multiple Extension Servers
type Clients []v1alpha1.NetworkingExtensionsClient

func NewClientsFromSettings(ctx context.Context, extensionsServerOptions []*v1alpha2.NetworkingExtensionsServer) (Clients, error) {
	var extensionsClients Clients
	for _, extensionsServer := range extensionsServerOptions {
		extensionsServerAddr := extensionsServer.GetAddress()
		if extensionsServerAddr == "" {
			return nil, eris.Errorf("must specify extensions server address")
		}
		dialOpts := grpc.DialOpts{
			Address:                    extensionsServerAddr,
			Insecure:                   extensionsServer.GetInsecure(),
			ReconnectOnNetworkFailures: extensionsServer.GetReconnectOnNetworkFailures(),
		}
		grpcConnection, err := dialOpts.Dial(ctx)
		if err != nil {
			return nil, err
		}
		extensionsClients = append(extensionsClients, v1alpha1.NewNetworkingExtensionsClient(grpcConnection))
	}
	return extensionsClients, nil
}

// WatchPushNotifications watches push notifications from the available extension servers until the context is cancelled.
// Will call pushFn() when a notification is received.
func (c Clients) WatchPushNotifications(ctx context.Context, pushFn PushFunc) error {
	for _, exClient := range c {
		exClient := exClient // pike
		go func() {
			if err := handlePushesForever(ctx, exClient, pushFn); err != nil {
				contextutils.LoggerFrom(ctx).DPanicf("failed to start push notification watch with client %+v", exClient)
			}
		}()
	}
	return nil
}

// handles push notifications for an individual connection stream
func handlePushesForever(ctx context.Context, exClient v1alpha1.NetworkingExtensionsClient, pushFn PushFunc) error {
	notifications, err := exClient.WatchPushNotifications(ctx, &v1alpha1.WatchPushNotificationsRequest{})
	if err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		notification, err := notifications.Recv()
		if err != nil {
			// attempt to restart the notification stream
			return handlePushesForever(ctx, exClient, pushFn)
		}
		pushFn(notification)
	}
}

// Clientset provides a handle to fetching a set of cahced gRPC clients
type Clientset interface {
	// ConfigureServers updates the set of servers this Clientset is configured with.
	// Returns true if servers were updated, false if options have not changed
	ConfigureServers(extensionsServerOptions []*v1alpha2.NetworkingExtensionsServer) (bool, error)

	// GetClients returns the set of Extension clients that are cached with this Clientset.
	// Must be called after UpdateServers
	GetClients() Clients

	// WatchPushNotifications watches push notifications from the available extension servers until the Clientset's root context is cancelled.
	// Will call pushFn() when a notification is received.
	// Should be called after UpdateServers
	WatchPushNotifications(pushFn PushFunc) error
}

type clientset struct {
	rootCtx       context.Context
	cachedClients cachedClients
}

func NewClientset(rootCtx context.Context) *clientset {
	return &clientset{rootCtx: rootCtx}
}

type cachedClients struct {
	lock        sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	optionsHash uint64 // this hash used to keep track of changes to the server options
	clients     Clients
}

func (c *clientset) ConfigureServers(extensionsServerOptions []*v1alpha2.NetworkingExtensionsServer) (bool, error) {
	optionsHash, err := hashutils.HashAllSafe(nil, extensionsServerOptions)
	if err != nil {
		return false, err
	}
	if c.cachedClients.optionsHash == optionsHash {
		// nothing to do, options have remained the same
		return false, nil
	}

	newContext, newCancel := context.WithCancel(c.rootCtx)
	newClients, err := NewClientsFromSettings(newContext, extensionsServerOptions)
	if err != nil {
		return false, err
	}

	c.cachedClients.lock.Lock()

	// cancel previous clients
	if c.cachedClients.cancel != nil {
		c.cachedClients.cancel()
	}
	// update latest clients
	c.cachedClients.clients = newClients
	c.cachedClients.ctx = newContext
	c.cachedClients.cancel = newCancel
	c.cachedClients.optionsHash = optionsHash

	c.cachedClients.lock.Unlock()

	return true, nil
}

func (c *clientset) GetClients() Clients {
	c.cachedClients.lock.RLock()
	clients := c.cachedClients.clients
	c.cachedClients.lock.RUnlock()

	return clients
}

func (c *clientset) WatchPushNotifications(pushFn PushFunc) error {
	return c.GetClients().WatchPushNotifications(c.cachedClients.ctx, pushFn)
}
