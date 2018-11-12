package v1

//go:generate ./generate.sh

//proteus:generate
type PrometheusConfig struct {
	Global        Global         `json:"global"`
	ScrapeConfigs []ScrapeConfig `json:"scrape_configs"`
}

type Global struct {
	ScrapeInterval string `json:"scrape_interval"`
}

type Namespaces struct {
	Names []string `json:"names"`
}

type KubernetesSdConfig struct {
	Namespaces Namespaces `json:"namespaces"`
	Role       string     `json:"role"`
}

type RelabelConfig struct {
	Action       string   `json:"action"`
	Regex        string   `json:"regex"`
	SourceLabels []string `json:"source_labels"`
}

type MetricRelabelConfig struct {
	Action       string   `json:"action"`
	Regex        string   `json:"regex"`
	SourceLabels []string `json:"source_labels"`
}

type TLSConfig struct {
	CaFile string `json:"ca_file"`
}

type ScrapeConfig struct {
	JobName              string                `json:"job_name"`
	KubernetesSdConfigs  []KubernetesSdConfig  `json:"kubernetes_sd_configs"`
	RelabelConfigs       []RelabelConfig       `json:"relabel_configs"`
	ScrapeInterval       string                `json:"scrape_interval,omitempty"`
	MetricRelabelConfigs []MetricRelabelConfig `json:"metric_relabel_configs,omitempty"`
	MetricsPath          string                `json:"metrics_path,omitempty"`
	BearerTokenFile      string                `json:"bearer_token_file,omitempty"`
	Scheme               string                `json:"scheme,omitempty"`
	TLSConfig            TLSConfig             `json:"tls_config,omitempty"`
}
