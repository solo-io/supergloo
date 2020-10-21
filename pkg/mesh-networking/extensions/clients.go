package extensions

import (
	"context"

	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/extensions/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/settings.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/common/utils/grpc"
)

// Clients provides a convenience wrapper for a set of clients to communicate with multiple Extension Servers
type Clients []v1alpha1.NetworkingExtensionsClient

// convenience function to construct a set of Extension Clients from eSettings
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
func (c Clients) WatchPushNotifications(ctx context.Context, pushFn func()) error {
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
func handlePushesForever(ctx context.Context, exClient v1alpha1.NetworkingExtensionsClient, pushFn func()) error {
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
		_, err := notifications.Recv()
		if err != nil {
			// attempt to restart the notification stream
			return handlePushesForever(ctx, exClient, pushFn)
		}
		pushFn()
	}
}
