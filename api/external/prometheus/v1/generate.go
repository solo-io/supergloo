package v1

import (
	config_util "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	sd_config "github.com/prometheus/prometheus/discovery/config"
	"net/url"
	"regexp"
)

//go:generate ./generate.sh

//proteus:generate
type Config struct {
	GlobalConfig   GlobalConfig    `json:"global"`
	AlertingConfig AlertingConfig  `json:"alerting,omitempty"`
	RuleFiles      []string        `json:"rule_files,omitempty"`
	ScrapeConfigs  []*ScrapeConfig `json:"scrape_configs,omitempty"`

	RemoteWriteConfigs []*RemoteWriteConfig `yaml:"remote_write,omitempty"`
	RemoteReadConfigs  []*RemoteReadConfig  `yaml:"remote_read,omitempty"`

	// original is the input from which the config was parsed.
	original string
}

type GlobalConfig struct {
	// How frequently to scrape targets by default.
	ScrapeInterval model.Duration `yaml:"scrape_interval,omitempty"`
	// The default timeout when scraping targets.
	ScrapeTimeout model.Duration `yaml:"scrape_timeout,omitempty"`
	// How frequently to evaluate rules by default.
	EvaluationInterval model.Duration `yaml:"evaluation_interval,omitempty"`
	// The labels to add to any timeseries that this Prometheus instance scrapes.
	ExternalLabels model.LabelSet `yaml:"external_labels,omitempty"`
}

type ScrapeConfig struct {
	// The job name to which the job label is set by default.
	JobName string `yaml:"job_name"`
	// Indicator whether the scraped metrics should remain unmodified.
	HonorLabels bool `yaml:"honor_labels,omitempty"`
	// A set of query parameters with which the target is scraped.
	Params url.Values `yaml:"params,omitempty"`
	// How frequently to scrape the targets of this scrape config.
	ScrapeInterval model.Duration `yaml:"scrape_interval,omitempty"`
	// The timeout for scraping targets of this config.
	ScrapeTimeout model.Duration `yaml:"scrape_timeout,omitempty"`
	// The HTTP resource path on which to fetch metrics from targets.
	MetricsPath string `yaml:"metrics_path,omitempty"`
	// The URL scheme with which to fetch metrics from targets.
	Scheme string `yaml:"scheme,omitempty"`
	// More than this many samples post metric-relabelling will cause the scrape to fail.
	SampleLimit uint `yaml:"sample_limit,omitempty"`

	// We cannot do proper Go type embedding below as the parser will then parse
	// values arbitrarily into the overflow maps of further-down types.

	ServiceDiscoveryConfig sd_config.ServiceDiscoveryConfig `yaml:",inline"`
	HTTPClientConfig       config_util.HTTPClientConfig     `yaml:",inline"`

	// List of target relabel configurations.
	RelabelConfigs []*RelabelConfig `yaml:"relabel_configs,omitempty"`
	// List of metric relabel configurations.
	MetricRelabelConfigs []*RelabelConfig `yaml:"metric_relabel_configs,omitempty"`
}


// RelabelConfig is the configuration for relabeling of target label sets.
type RelabelConfig struct {
	// A list of labels from which values are taken and concatenated
	// with the configured separator in order.
	SourceLabels model.LabelNames `yaml:"source_labels,flow,omitempty"`
	// Separator is the string between concatenated values from the source labels.
	Separator string `yaml:"separator,omitempty"`
	// Regex against which the concatenation is matched.
	Regex Regexp `yaml:"regex,omitempty"`
	// Modulus to take of the hash of concatenated values from the source labels.
	Modulus uint64 `yaml:"modulus,omitempty"`
	// TargetLabel is the label to which the resulting string is written in a replacement.
	// Regexp interpolation is allowed for the replace action.
	TargetLabel string `yaml:"target_label,omitempty"`
	// Replacement is the regex replacement pattern to be used.
	Replacement string `yaml:"replacement,omitempty"`
	// Action is the action to be performed for the relabeling.
	Action RelabelAction `yaml:"action,omitempty"`
}


type AlertingConfig struct {
	AlertRelabelConfigs []*RelabelConfig      `yaml:"alert_relabel_configs,omitempty"`
	AlertmanagerConfigs []*AlertmanagerConfig `yaml:"alertmanagers,omitempty"`
}
type AlertmanagerConfig struct {
	// We cannot do proper Go type embedding below as the parser will then parse
	// values arbitrarily into the overflow maps of further-down types.

	ServiceDiscoveryConfig sd_config.ServiceDiscoveryConfig `yaml:",inline"`
	HTTPClientConfig       config_util.HTTPClientConfig     `yaml:",inline"`

	// The URL scheme to use when talking to Alertmanagers.
	Scheme string `yaml:"scheme,omitempty"`
	// Path prefix to add in front of the push endpoint path.
	PathPrefix string `yaml:"path_prefix,omitempty"`
	// The timeout used when sending alerts.
	Timeout model.Duration `yaml:"timeout,omitempty"`

	// List of Alertmanager relabel configurations.
	RelabelConfigs []*RelabelConfig `yaml:"relabel_configs,omitempty"`
}

// ClientCert contains client cert credentials.
type ClientCert struct {
	Cert string             `yaml:"cert"`
	Key  config_util.Secret `yaml:"key"`
}

// FileSDConfig is the configuration for file based discovery.
type FileSDConfig struct {
	Files           []string       `yaml:"files"`
	RefreshInterval model.Duration `yaml:"refresh_interval,omitempty"`
}

// RelabelAction is the action to be performed on relabeling.
type RelabelAction string

// Regexp encapsulates a regexp.Regexp and makes it YAML marshallable.
type Regexp struct {
	*regexp.Regexp
	original string
}

// RemoteWriteConfig is the configuration for writing to remote storage.
type RemoteWriteConfig struct {
	URL                 *config_util.URL `yaml:"url"`
	RemoteTimeout       model.Duration   `yaml:"remote_timeout,omitempty"`
	WriteRelabelConfigs []*RelabelConfig `yaml:"write_relabel_configs,omitempty"`

	// We cannot do proper Go type embedding below as the parser will then parse
	// values arbitrarily into the overflow maps of further-down types.
	HTTPClientConfig config_util.HTTPClientConfig `yaml:",inline"`
	QueueConfig      QueueConfig                  `yaml:"queue_config,omitempty"`
}

// QueueConfig is the configuration for the queue used to write to remote
// storage.
type QueueConfig struct {
	// Number of samples to buffer per shard before we start dropping them.
	Capacity int `yaml:"capacity,omitempty"`

	// Max number of shards, i.e. amount of concurrency.
	MaxShards int `yaml:"max_shards,omitempty"`

	// Maximum number of samples per send.
	MaxSamplesPerSend int `yaml:"max_samples_per_send,omitempty"`

	// Maximum time sample will wait in buffer.
	BatchSendDeadline model.Duration `yaml:"batch_send_deadline,omitempty"`

	// Max number of times to retry a batch on recoverable errors.
	MaxRetries int `yaml:"max_retries,omitempty"`

	// On recoverable errors, backoff exponentially.
	MinBackoff model.Duration `yaml:"min_backoff,omitempty"`
	MaxBackoff model.Duration `yaml:"max_backoff,omitempty"`
}

// RemoteReadConfig is the configuration for reading from remote storage.
type RemoteReadConfig struct {
	URL           *config_util.URL `yaml:"url"`
	RemoteTimeout model.Duration   `yaml:"remote_timeout,omitempty"`
	ReadRecent    bool             `yaml:"read_recent,omitempty"`
	// We cannot do proper Go type embedding below as the parser will then parse
	// values arbitrarily into the overflow maps of further-down types.
	HTTPClientConfig config_util.HTTPClientConfig `yaml:",inline"`

	// RequiredMatchers is an optional list of equality matchers which have to
	// be present in a selector to query the remote read endpoint.
	RequiredMatchers model.LabelSet `yaml:"required_matchers,omitempty"`
}
