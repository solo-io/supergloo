package server

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/golang/sync/errgroup"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	rpc_v1 "github.com/solo-io/service-mesh-hub/services/apiserver/pkg/api/v1"
	"github.com/solo-io/service-mesh-hub/services/apiserver/pkg/server/health_check"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/stats/view"
	"google.golang.org/grpc"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

func init() {
	view.Register(ocgrpc.DefaultServerViews...)
}

type GrpcServer interface {
	Run(ctx context.Context) error
	Stop()
}

type grpcServer struct {
	server        *grpc.Server
	healthChecker health_check.HealthChecker
}

func NewGrpcServer(
	ctx context.Context,
	kubeClusterApiServer rpc_v1.KubernetesClusterApiServer,
	meshServer rpc_v1.MeshApiServer,
	meshWorkloadApiServer rpc_v1.MeshWorkloadApiServer,
	meshServiceApiServer rpc_v1.MeshServiceApiServer,
	virtualMeshApiServer rpc_v1.VirtualMeshApiServer,
	healthChecker health_check.HealthChecker,
) GrpcServer {

	logger := contextutils.LoggerFrom(ctx)
	server := grpc.NewServer(
		grpc.StatsHandler(&ocgrpc.ServerHandler{}),
		grpc_middleware.WithUnaryServerChain(grpc_zap.UnaryServerInterceptor(logger.Desugar())),
	)
	// register grpc health check
	healthpb.RegisterHealthServer(server, healthChecker)
	// register handlers
	rpc_v1.RegisterKubernetesClusterApiServer(server, kubeClusterApiServer)
	rpc_v1.RegisterMeshApiServer(server, meshServer)
	rpc_v1.RegisterMeshServiceApiServer(server, meshServiceApiServer)
	rpc_v1.RegisterMeshWorkloadApiServer(server, meshWorkloadApiServer)
	rpc_v1.RegisterVirtualMeshApiServer(server, virtualMeshApiServer)

	return &grpcServer{
		healthChecker: healthChecker,
		server:        server,
	}
}

func (g *grpcServer) Run(ctx context.Context) error {
	eg, _ := errgroup.WithContext(ctx)
	// Start http health check
	healthCheckPort := env.GetHealthCheckPort()
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", healthCheckPort))
	if err != nil {
		return eris.Wrapf(err, "failed to setup health check listener")
	}
	eg.Go(func() error {
		return http.Serve(listener, g.healthChecker)
	})
	// start grpc listener
	grpcPort := env.GetGrpcPort()
	listener, err = net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		return eris.Wrapf(err, "failed to setup grpc listener")
	}
	eg.Go(func() error {
		return g.server.Serve(listener)
	})
	return eg.Wait()
}

func (g *grpcServer) Stop() {
	g.server.GracefulStop()
}
