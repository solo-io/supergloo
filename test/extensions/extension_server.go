package extensions

import (
	"context"
	"fmt"
	"net"

	"go.uber.org/atomic"

	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/output/istio"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/extensions"
	"google.golang.org/grpc"

	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/extensions/v1alpha1"
)

const ExtensionsServerPort = 2345

type testExtensionsServer struct {
	createMeshPatches func(ctx context.Context, mesh *v1alpha2.MeshSpec) (istio.Builder, error)
	hasConnected      *atomic.Bool
}

func NewTestExtensionsServer() *testExtensionsServer {
	return &testExtensionsServer{createMeshPatches: getCreateMeshPatchesFunc(), hasConnected: &atomic.Bool{}}
}

// Runs an e2e implementation of a grpc extensions service for Networking
// that adds a route to a simple "HelloWorld" server running on the local machine (reachable via `host.docker.internal` from inside KinD)
func (t *testExtensionsServer) Run() error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%v", ExtensionsServerPort))
	if err != nil {
		return err
	}
	grpcSrv := grpc.NewServer()
	v1alpha1.RegisterNetworkingExtensionsServer(grpcSrv, t)
	return grpcSrv.Serve(l)
}

func (t *testExtensionsServer) GetMeshPatches(ctx context.Context, request *v1alpha1.MeshPatchRequest) (*v1alpha1.PatchList, error) {
	outputs, err := t.createMeshPatches(ctx, request.Mesh.Spec)
	if err != nil {
		return nil, err
	}

	return &v1alpha1.PatchList{
		PatchedResources: extensions.OutputsToProto(outputs),
	}, nil
}

func (t *testExtensionsServer) GetTrafficTargetPatches(ctx context.Context, request *v1alpha1.TrafficTargetPatchRequest) (*v1alpha1.PatchList, error) {
	return &v1alpha1.PatchList{}, nil
}

func (t *testExtensionsServer) GetWorkloadPatches(ctx context.Context, request *v1alpha1.WorkloadPatchRequest) (*v1alpha1.PatchList, error) {
	return &v1alpha1.PatchList{}, nil
}

func (t *testExtensionsServer) WatchPushNotifications(request *v1alpha1.WatchPushNotificationsRequest, server v1alpha1.NetworkingExtensions_WatchPushNotificationsServer) error {

	// one to start
	if err := server.Send(&v1alpha1.PushNotification{}); err != nil {
		return err
	}

	// client has connected
	t.hasConnected.Store(true)

	// sleep forever
	select {
	case <-server.Context().Done():
		return nil
	}
}

// returns true if a client has connected to this server
func (t *testExtensionsServer) HasConnected() bool {
	return t.hasConnected.Load()
}
