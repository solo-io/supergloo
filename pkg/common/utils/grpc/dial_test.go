package grpc_test

import (
	"context"
	"net"
	"time"

	"github.com/hashicorp/go-multierror"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/solo-io/service-mesh-hub/pkg/common/utils/grpc"
	"github.com/solo-io/service-mesh-hub/pkg/common/utils/grpc/test_api"
	"google.golang.org/grpc"
)

//go:generate protoc -I=. --go_out=plugins=grpc:. test.proto

var req = &test_api.Request{}

// the purpose of this test is to verify that gRPC options can handle
// reconnecting the client to a disconnected server
var _ = Describe("Dial Integration Test", func() {
	It("reconnects automatically when the server restarts", func() {
		s := serverImpl{}
		ctx := context.Background()
		serverCtx, stopServer := context.WithCancel(ctx)
		addr := "localhost:1234"
		err := s.GoListen(serverCtx, addr)
		Expect(err).NotTo(HaveOccurred())

		cli, err := newClient(ctx, addr)
		Expect(err).NotTo(HaveOccurred())

		_, err = cli.Invoke(ctx, req)
		Expect(err).NotTo(HaveOccurred())

		// disconnect server, expect client to fail
		stopServer()

		time.Sleep(time.Second / 4)

		go func() {
			defer GinkgoRecover()

			// sleep to grow retry backoff
			time.Sleep(time.Second * 15)

			// start server back up
			err = s.GoListen(ctx, addr)
			Expect(err).NotTo(HaveOccurred())
		}()

		// expect client to retry and succeed
		_, err = cli.Invoke(ctx, req)
		Expect(err).NotTo(HaveOccurred())

	})
})

type serverImpl struct {
	c int
}

func (s *serverImpl) GoListen(ctx context.Context, addr string) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	srv := grpc.NewServer()
	test_api.RegisterTestServiceServer(srv, s)

	go func() {
		defer GinkgoRecover()
		<-ctx.Done()
		err := l.Close()
		Expect(err).NotTo(HaveOccurred())
		srv.Stop()
	}()
	go func() {
		select {
		case <-ctx.Done():
			return
		default:
		}
		s.c++
		defer GinkgoRecover()
		err := srv.Serve(l)
		if err != nil {
			time.Sleep(time.Second / 2)
			select {
			case <-ctx.Done():
				return
			default:
			}
			err := srv.Serve(l)
			Expect(err).NotTo(HaveOccurred())
		}
	}()

	checkSrvStarted := func() error {
		cli, err := newClient(ctx, addr)
		if err != nil {
			return err
		}
		_, err = cli.Invoke(ctx, req)
		return err
	}

	// try to connect for a max of 4 attempts
	var errs error
	for i := 0; i < 4; i++ {
		if err := checkSrvStarted(); err != nil {
			errs = multierror.Append(errs, err)
			// sleep before retry
			time.Sleep(time.Second / 2)
			continue
		}
		// succcess
		return nil
	}
	return errs
}

func newClient(ctx context.Context, addr string) (test_api.TestServiceClient, error) {
	cc, err := DialOpts{Address: addr, Insecure: true, ReconnectOnNetworkFailures: true}.Dial(ctx)
	if err != nil {
		return nil, err
	}
	return test_api.NewTestServiceClient(cc), nil
}

func (s *serverImpl) Invoke(ctx context.Context, request *test_api.Request) (*test_api.Response, error) {
	return &test_api.Response{}, nil
}
