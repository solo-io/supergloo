package grpc

import (
	"context"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/codes"
)

// DialOpts provides a common set of options for initiating connections to a gRPC server.
type DialOpts struct {
	// address of the gRPC server
	Address string

	// connect over plaintext or HTTPS
	Insecure bool

	// Enable fast reconnection attempts on network failures (gRPC code 14 'unavailable')
	ReconnectOnNetworkFailures bool
}

func (o DialOpts) Dial(ctx context.Context) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{grpc.WithBlock()}
	if o.Insecure {
		opts = append(opts, grpc.WithInsecure())
	}
	if o.ReconnectOnNetworkFailures {
		retryOpts := []grpc_retry.CallOption{
			grpc_retry.WithCodes(codes.Unavailable),
			grpc_retry.WithMax(10),
			grpc_retry.WithBackoff(grpc_retry.BackoffLinear(time.Second * 2)),
		}
		connectionBackoff := backoff.DefaultConfig
		connectionBackoff.MaxDelay = time.Second * 2
		opts = append(opts,
			grpc.WithConnectParams(grpc.ConnectParams{Backoff: connectionBackoff}),
			grpc.WithStreamInterceptor(grpc_middleware.ChainStreamClient(
				//WithResetConnectionBackoffStream(),
				grpc_retry.StreamClientInterceptor(retryOpts...),
			)),
			grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(
				//WithResetConnectionBackoffUnary(),
				grpc_retry.UnaryClientInterceptor(retryOpts...),
			)),
		)
	}
	cc, err := grpc.DialContext(ctx, o.Address, opts...)
	if err != nil {
		return nil, err
	}
	return cc, nil
}
