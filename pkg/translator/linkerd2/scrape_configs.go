package linkerd2

import (
	"github.com/ghodss/yaml"
	"github.com/solo-io/supergloo/pkg/api/external/prometheus"
)

var linkerdScrapeConfigs []prometheus.ScrapeConfig

func init() {
	err := yaml.Unmarshal([]byte(linkerd2ScrapeConfigsYaml), &linkerdScrapeConfigs)
	if err != nil {
		panic("failed to parse linkerd2ScrapeConfigsYaml: " + err.Error())
	}
}

const linkerd2ScrapeConfigsYaml = `# imported from https://linkerd.io/2/observability/prometheus/
- job_name: 'linkerd-controller'
  kubernetes_sd_configs:
  - role: pod
    namespaces:
      names: ['{{.Namespace}}']
  relabel_configs:
  - source_labels:
    - __meta_kubernetes_pod_label_linkerd_io_control_plane_component
    - __meta_kubernetes_pod_container_port_name
    action: keep
    regex: (.*);admin-http$
  - source_labels: [__meta_kubernetes_pod_container_name]
    action: replace
    target_label: component

- job_name: 'linkerd-proxy'
  kubernetes_sd_configs:
  - role: pod
  relabel_configs:
  - source_labels:
    - __meta_kubernetes_pod_container_name
    - __meta_kubernetes_pod_container_port_name
    - __meta_kubernetes_pod_label_linkerd_io_control_plane_ns
    action: keep
    regex: ^linkerd-proxy;linkerd-metrics;{{.Namespace}}$
  - source_labels: [__meta_kubernetes_namespace]
    action: replace
    target_label: namespace
  - source_labels: [__meta_kubernetes_pod_name]
    action: replace
    target_label: pod
  # special case k8s' "job" label, to not interfere with prometheus' "job"
  # label
  # __meta_kubernetes_pod_label_linkerd_io_proxy_job=foo =>
  # k8s_job=foo
  - source_labels: [__meta_kubernetes_pod_label_linkerd_io_proxy_job]
    action: replace
    target_label: k8s_job
  # __meta_kubernetes_pod_label_linkerd_io_proxy_deployment=foo =>
  # deployment=foo
  - action: labelmap
    regex: __meta_kubernetes_pod_label_linkerd_io_proxy_(.+)
  # drop all labels that we just made copies of in the previous labelmap
  - action: labeldrop
    regex: __meta_kubernetes_pod_label_linkerd_io_proxy_(.+)
  # __meta_kubernetes_pod_label_linkerd_io_foo=bar =>
  # foo=bar
  - action: labelmap
    regex: __meta_kubernetes_pod_label_linkerd_io_(.+)`
