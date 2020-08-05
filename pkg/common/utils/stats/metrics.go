package stats

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

func AddMetrics(mux *http.ServeMux) {
	handler := promhttp.HandlerFor(metrics.Registry, promhttp.HandlerOpts{
		ErrorHandling: promhttp.HTTPErrorOnError,
	})
	profileDescriptions["/metrics"] = "Prometheus metrics"

	mux.Handle("/metrics", handler)
}
