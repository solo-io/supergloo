package server

import (
	"context"
	"fmt"
	"net"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	rpc_v1 "github.com/solo-io/service-mesh-hub/services/apiserver/pkg/api/v1"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/stats/view"
	"google.golang.org/grpc"
)

func init() {
	view.Register(ocgrpc.DefaultServerViews...)
}

type GrpcServer interface {
	Run() error
	Stop()
}

type grpcServer struct {
	server *grpc.Server
}

func NewGrpcServer(
	ctx context.Context,
	kubeClusterApiServer rpc_v1.KubernetesClusterApiServer,
	meshServer rpc_v1.MeshApiServer,
	meshWorkloadApiServer rpc_v1.MeshWorkloadApiServer,
	meshServiceApiServer rpc_v1.MeshServiceApiServer,
) GrpcServer {

	logger := contextutils.LoggerFrom(ctx)
	server := grpc.NewServer(
		grpc.StatsHandler(&ocgrpc.ServerHandler{}),
		grpc_middleware.WithUnaryServerChain(grpc_zap.UnaryServerInterceptor(logger.Desugar())),
	)

	rpc_v1.RegisterKubernetesClusterApiServer(server, kubeClusterApiServer)
	rpc_v1.RegisterMeshApiServer(server, meshServer)
	rpc_v1.RegisterMeshServiceApiServer(server, meshServiceApiServer)
	rpc_v1.RegisterMeshWorkloadApiServer(server, meshWorkloadApiServer)

	return &grpcServer{server: server}
}

func (g *grpcServer) Run() error {
	grpcPort := env.GetGrpcPort()
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		return eris.Wrapf(err, "failed to setup listener")
	}
	return g.server.Serve(listener)
}

func (g *grpcServer) Stop() {
	g.server.GracefulStop()
}
