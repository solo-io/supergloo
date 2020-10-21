package extensions

import (
	"context"

	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/extensions/v1alpha1"
)

// testExtensionsServer is an e2e implementation of a grpc extensions service for Networking
// that adds a route to an echo server running on the local machine (reachable via `host.docker.internal` from inside KinD)
type testExtensionsServer struct{}

func (t *testExtensionsServer) GetTrafficTargetPatches(ctx context.Context, request *v1alpha1.TrafficTargetPatchRequest) (*v1alpha1.PatchList, error) {
	return nil, nil
}

func (t *testExtensionsServer) GetWorkloadPatches(ctx context.Context, request *v1alpha1.WorkloadPatchRequest) (*v1alpha1.PatchList, error) {
	return nil, nil
}

func (t *testExtensionsServer) GetMeshPatches(ctx context.Context, request *v1alpha1.MeshPatchRequest) (*v1alpha1.PatchList, error) {
	panic("implement me")
}

func (t *testExtensionsServer) WatchPushNotifications(request *v1alpha1.WatchPushNotificationsRequest, server v1alpha1.NetworkingExtensions_WatchPushNotificationsServer) error {

}
