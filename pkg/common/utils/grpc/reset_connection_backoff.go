package grpc

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// WithResetConnectionBackoffUnary creates a Unary Interceptor which resets the underlying connection's retry backoff timer.
// This allows grpc connections to quickly attempt to reconnect to a server e.g. after a server restart.
func WithResetConnectionBackoffUnary() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		err := invoker(ctx, method, req, reply, cc, opts...)
		if status.Code(err) == codes.Unavailable {
			cc.ResetConnectBackoff()
			err = invoker(ctx, method, req, reply, cc, opts...)
		}
		return err
	}
}

// WithResetConnectionBackoffStream creates a Stream Interceptor which resets the underlying connection's retry backoff timer.
// This allows grpc connections to quickly attempt to reconnect to a server e.g. after a server restart.
func WithResetConnectionBackoffStream() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		cs, err := streamer(ctx, desc, cc, method, opts...)
		if status.Code(err) == codes.Unavailable {
			cc.ResetConnectBackoff()
			cs, err = streamer(ctx, desc, cc, method, opts...)
		}
		return cs, err
	}
}
