package health_check

import (
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"

	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

type HealthChecker interface {
	healthpb.HealthServer
	http.Handler
	Fail()
}

type healthChecker struct {
	*health.Server
	ok uint32
}

func NewHealthChecker() HealthChecker {
	ret := &healthChecker{
		ok:     1,
		Server: health.NewServer(),
	}
	ret.Server.SetServingStatus("mesh-apiserver", healthpb.HealthCheckResponse_SERVING)

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-sigterm
		atomic.StoreUint32(&ret.ok, 0)
		ret.Server.SetServingStatus("mesh-apiserver", healthpb.HealthCheckResponse_NOT_SERVING)
	}()

	return ret
}

func (hc *healthChecker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ok := atomic.LoadUint32(&hc.ok)
	if ok == 1 {
		w.Write([]byte("OK"))
	} else {
		w.WriteHeader(500)
	}
}

func (hc *healthChecker) Fail() {
	atomic.StoreUint32(&hc.ok, 0)
	hc.Server.SetServingStatus("mesh-apiserver", healthpb.HealthCheckResponse_NOT_SERVING)
}
