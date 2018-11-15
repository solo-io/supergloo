package istio

import (
	"github.com/ghodss/yaml"
	"github.com/solo-io/supergloo/pkg/api/external/prometheus"
)

var IstioScrapeConfigs []prometheus.ScrapeConfig

func init() {
	err := yaml.Unmarshal([]byte(istioScrapeConfigsYaml), &IstioScrapeConfigs)
	if err != nil {
		panic("failed to parse istioScrapeConfigsYaml: " + err.Error())
	}
}

const istioScrapeConfigsYaml = `# imported from istio-demo.yaml
- job_name: 'istio-mesh'
  # Override the global default and scrape targets from this job every 5 seconds.
  scrape_interval: 5s

  kubernetes_sd_configs:
  - role: endpoints
    namespaces:
      names:
      - istio-system

  relabel_configs:
  - source_labels: [__meta_kubernetes_service_name, __meta_kubernetes_endpoint_port_name]
    action: keep
    regex: istio-telemetry;prometheus


# Scrape config for envoy stats
- job_name: 'envoy-stats'
  metrics_path: /stats/prometheus
  kubernetes_sd_configs:
  - role: pod

  relabel_configs:
  - source_labels: [__meta_kubernetes_pod_container_port_name]
    action: keep
    regex: '.*-envoy-prom'
  - source_labels: [__address__, __meta_kubernetes_pod_annotation_prometheus_io_port]
    action: replace
    regex: ([^:]+)(?::\d+)?;(\d+)
    replacement: $1:15090
    target_label: __address__
  - action: labelmap
    regex: __meta_kubernetes_pod_label_(.+)
  - source_labels: [__meta_kubernetes_namespace]
    action: replace
    target_label: namespace
  - source_labels: [__meta_kubernetes_pod_name]
    action: replace
    target_label: pod_name

  metric_relabel_configs:
  # Exclude some of the envoy metrics that have massive cardinality
  # This list may need to be pruned further moving forward, as informed
  # by performance and scalability testing.
  - source_labels: [ cluster_name ]
    regex: '(outbound|inbound|prometheus_stats).*'
    action: drop
  - source_labels: [ tcp_prefix ]
    regex: '(outbound|inbound|prometheus_stats).*'
    action: drop
  - source_labels: [ listener_address ]
    regex: '(.+)'
    action: drop
  - source_labels: [ http_conn_manager_listener_prefix ]
    regex: '(.+)'
    action: drop
  - source_labels: [ http_conn_manager_prefix ]
    regex: '(.+)'
    action: drop
  - source_labels: [ __name__ ]
    regex: 'envoy_tls.*'
    action: drop
  - source_labels: [ __name__ ]
    regex: 'envoy_tcp_downstream.*'
    action: drop
  - source_labels: [ __name__ ]
    regex: 'envoy_http_(stats|admin).*'
    action: drop
  - source_labels: [ __name__ ]
    regex: 'envoy_cluster_(lb|retry|bind|internal|max|original).*'
    action: drop


- job_name: 'istio-policy'
  # Override the global default and scrape targets from this job every 5 seconds.
  scrape_interval: 5s
  # metrics_path defaults to '/metrics'
  # scheme defaults to 'http'.

  kubernetes_sd_configs:
  - role: endpoints
    namespaces:
      names:
      - istio-system


  relabel_configs:
  - source_labels: [__meta_kubernetes_service_name, __meta_kubernetes_endpoint_port_name]
    action: keep
    regex: istio-policy;http-monitoring

- job_name: 'istio-telemetry'
  # Override the global default and scrape targets from this job every 5 seconds.
  scrape_interval: 5s
  # metrics_path defaults to '/metrics'
  # scheme defaults to 'http'.

  kubernetes_sd_configs:
  - role: endpoints
    namespaces:
      names:
      - istio-system

  relabel_configs:
  - source_labels: [__meta_kubernetes_service_name, __meta_kubernetes_endpoint_port_name]
    action: keep
    regex: istio-telemetry;http-monitoring

- job_name: 'pilot'
  # Override the global default and scrape targets from this job every 5 seconds.
  scrape_interval: 5s
  # metrics_path defaults to '/metrics'
  # scheme defaults to 'http'.

  kubernetes_sd_configs:
  - role: endpoints
    namespaces:
      names:
      - istio-system

  relabel_configs:
  - source_labels: [__meta_kubernetes_service_name, __meta_kubernetes_endpoint_port_name]
    action: keep
    regex: istio-pilot;http-monitoring

- job_name: 'galley'
  # Override the global default and scrape targets from this job every 5 seconds.
  scrape_interval: 5s
  # metrics_path defaults to '/metrics'
  # scheme defaults to 'http'.

  kubernetes_sd_configs:
  - role: endpoints
    namespaces:
      names:
      - istio-system

  relabel_configs:
  - source_labels: [__meta_kubernetes_service_name, __meta_kubernetes_endpoint_port_name]
    action: keep
    regex: istio-galley;http-monitoring
`
