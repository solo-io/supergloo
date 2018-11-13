package prometheus

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/types"
	"github.com/pkg/errors"
	"github.com/solo-io/supergloo/pkg/api/external/prometheus/v1"
)

func ConfigFromResource(cfg *v1.Config) (*PrometheusConfig, error) {
	if cfg == nil {
		return nil, nil
	}
	buf := &bytes.Buffer{}
	if err := (&jsonpb.Marshaler{OrigName: true}).Marshal(buf, cfg.Prometheus); err != nil {
		return nil, errors.Wrapf(err, "failed to marshal proto struct")
	}
	var c PrometheusConfig
	str := string(buf.String())
	if err := json.Unmarshal([]byte(str), &c); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal raw json to prometheus config")
	}
	return &c, nil
}

func ConfigToResource(cfg *PrometheusConfig) (*v1.Config, error) {
	if cfg == nil {
		return nil, nil
	}
	jsn, err := json.Marshal(cfg)
	if err != nil {
		return nil, errors.Wrapf(err, "marshalling cfg")
	}
	var s types.Struct
	if err := jsonpb.Unmarshal(bytes.NewBuffer(jsn), &s); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal jsn to proto struct")
	}
	return &v1.Config{Prometheus: &s}, nil
}

type PrometheusConfig struct {
	Global        Global         `json:"global,omitempty"`
	ScrapeConfigs []ScrapeConfig `json:"scrape_configs,omitempty"`
}

type Global struct {
	ScrapeInterval Duration `json:"scrape_interval,omitempty"`
}
type ScrapeConfig struct {
	JobName              string               `json:"job_name,omitempty"`
	KubernetesSdConfigs  []KubernetesSdConfig `json:"kubernetes_sd_configs,omitempty"`
	RelabelConfigs       []RelabelConfig      `json:"relabel_configs,omitempty"`
	ScrapeInterval       string               `json:"scrape_interval,omitempty"`
	MetricRelabelConfigs []RelabelConfig      `json:"metric_relabel_configs,omitempty,omitempty"`
	MetricsPath          string               `json:"metrics_path,omitempty"`
	BearerTokenFile      string               `json:"bearer_token_file,omitempty"`
	Scheme               string               `json:"scheme,omitempty"`
	TLSConfig            *TLSConfig           `json:"tls_config,omitempty"`
}

type TLSConfig struct {
	CaFile             string `json:"ca_file,omitempty"`
	CertFile           string `json:"cert_file,omitempty"`
	InsecureSkipVerify bool   `json:"insecure_skip_verify,omitempty"`
	KeyFile            string `json:"key_file,omitempty"`
}

type KubernetesSdConfig struct {
	Namespaces *Namespaces `json:"namespaces,omitempty"`
	Role       string      `json:"role,omitempty"`
}

type Namespaces struct {
	Names []string `json:"names,omitempty"`
}

// RelabelConfig is the configuration for relabeling of target label sets.
type RelabelConfig struct {
	// A list of labels from which values are taken and concatenated
	// with the configured separator in order.
	SourceLabels []string `json:"source_labels,flow,omitempty"`
	// Separator is the string between concatenated values from the source labels.
	Separator string `json:"separator,omitempty"`
	// Regex against which the concatenation is matched.
	// interface because prometheus decided it could be bool OR string ???
	Regex interface{} `json:"regex,omitempty"`
	// Modulus to take of the hash of concatenated values from the source labels.
	Modulus uint64 `json:"modulus,omitempty"`
	// TargetLabel is the label to which the resulting string is written in a replacement.
	// Regexp interpolation is allowed for the replace action.
	TargetLabel string `json:"target_label,omitempty"`
	// Replacement is the regex replacement pattern to be used.
	Replacement string `json:"replacement,omitempty"`
	// Action is the action to be performed for the relabeling.
	Action string `json:"action,omitempty"`
}

type Duration struct {
	time.Duration
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		d.Duration = time.Duration(value)
		return nil
	case string:
		var err error
		d.Duration, err = time.ParseDuration(value)
		if err != nil {
			return err
		}
		return nil
	default:
		return errors.New("invalid duration")
	}
}

//type Config struct {
//	GlobalConfig   GlobalConfig    `json:"global"`
//	AlertingConfig AlertingConfig  `json:"alerting,omitempty"`
//	RuleFiles      []string        `json:"rule_files,omitempty"`
//	ScrapeConfigs  []*ScrapeConfig `json:"scrape_configs,omitempty"`
//
//	RemoteWriteConfigs []*RemoteWriteConfig `json:"remote_write,omitempty"`
//	RemoteReadConfigs  []*RemoteReadConfig  `json:"remote_read,omitempty"`
//
//	// original is the input from which the config was parsed.
//	original string
//}
//
//type GlobalConfig struct {
//	// How frequently to scrape targets by default.
//	ScrapeInterval Duration `json:"scrape_interval,omitempty"`
//	// The default timeout when scraping targets.
//	ScrapeTimeout Duration `json:"scrape_timeout,omitempty"`
//	// How frequently to evaluate rules by default.
//	EvaluationInterval Duration `json:"evaluation_interval,omitempty"`
//	// The labels to add to any timeseries that this Prometheus instance scrapes.
//	ExternalLabels model.LabelSet `json:"external_labels,omitempty"`
//}
//
//type ScrapeConfig struct {
//	// The job name to which the job label is set by default.
//	JobName string `json:"job_name"`
//	// Indicator whether the scraped metrics should remain unmodified.
//	HonorLabels bool `json:"honor_labels,omitempty"`
//	// A set of query parameters with which the target is scraped.
//	Params url.Values `json:"params,omitempty"`
//	// How frequently to scrape the targets of this scrape config.
//	ScrapeInterval Duration `json:"scrape_interval,omitempty"`
//	// The timeout for scraping targets of this config.
//	ScrapeTimeout Duration `json:"scrape_timeout,omitempty"`
//	// The HTTP resource path on which to fetch metrics from targets.
//	MetricsPath string `json:"metrics_path,omitempty"`
//	// The URL scheme with which to fetch metrics from targets.
//	Scheme string `json:"scheme,omitempty"`
//	// More than this many samples post metric-relabelling will cause the scrape to fail.
//	SampleLimit uint `json:"sample_limit,omitempty"`
//
//	// We cannot do proper Go type embedding below as the parser will then parse
//	// values arbitrarily into the overflow maps of further-down types.
//
//	ServiceDiscoveryConfig sd_config.ServiceDiscoveryConfig `json:",inline"`
//	HTTPClientConfig       config_util.HTTPClientConfig     `json:",inline"`
//
//	// List of target relabel configurations.
//	RelabelConfigs []*RelabelConfig `json:"relabel_configs,omitempty"`
//	// List of metric relabel configurations.
//	MetricRelabelConfigs []*RelabelConfig `json:"metric_relabel_configs,omitempty"`
//}
//
//// RelabelConfig is the configuration for relabeling of target label sets.
//type RelabelConfig struct {
//	// A list of labels from which values are taken and concatenated
//	// with the configured separator in order.
//	SourceLabels model.LabelNames `json:"source_labels,flow,omitempty"`
//	// Separator is the string between concatenated values from the source labels.
//	Separator string `json:"separator,omitempty"`
//	// Regex against which the concatenation is matched.
//	Regex Regexp `json:"regex,omitempty"`
//	// Modulus to take of the hash of concatenated values from the source labels.
//	Modulus uint64 `json:"modulus,omitempty"`
//	// TargetLabel is the label to which the resulting string is written in a replacement.
//	// Regexp interpolation is allowed for the replace action.
//	TargetLabel string `json:"target_label,omitempty"`
//	// Replacement is the regex replacement pattern to be used.
//	Replacement string `json:"replacement,omitempty"`
//	// Action is the action to be performed for the relabeling.
//	Action RelabelAction `json:"action,omitempty"`
//}
//
//type AlertingConfig struct {
//	AlertRelabelConfigs []*RelabelConfig      `json:"alert_relabel_configs,omitempty"`
//	AlertmanagerConfigs []*AlertmanagerConfig `json:"alertmanagers,omitempty"`
//}
//type AlertmanagerConfig struct {
//	// We cannot do proper Go type embedding below as the parser will then parse
//	// values arbitrarily into the overflow maps of further-down types.
//
//	ServiceDiscoveryConfig sd_config.ServiceDiscoveryConfig `json:",inline"`
//	HTTPClientConfig       config_util.HTTPClientConfig     `json:",inline"`
//
//	// The URL scheme to use when talking to Alertmanagers.
//	Scheme string `json:"scheme,omitempty"`
//	// Path prefix to add in front of the push endpoint path.
//	PathPrefix string `json:"path_prefix,omitempty"`
//	// The timeout used when sending alerts.
//	Timeout Duration `json:"timeout,omitempty"`
//
//	// List of Alertmanager relabel configurations.
//	RelabelConfigs []*RelabelConfig `json:"relabel_configs,omitempty"`
//}
//
//// ClientCert contains client cert credentials.
//type ClientCert struct {
//	Cert string             `json:"cert"`
//	Key  config_util.Secret `json:"key"`
//}
//
//// FileSDConfig is the configuration for file based discovery.
//type FileSDConfig struct {
//	Files           []string `json:"files"`
//	RefreshInterval Duration `json:"refresh_interval,omitempty"`
//}
//
//// RelabelAction is the action to be performed on relabeling.
//type RelabelAction string
//
//// Regexp encapsulates a regexp.Regexp and makes it YAML marshallable.
//type Regexp interface{}
//
//// RemoteWriteConfig is the configuration for writing to remote storage.
//type RemoteWriteConfig struct {
//	URL                 *config_util.URL `json:"url"`
//	RemoteTimeout       Duration         `json:"remote_timeout,omitempty"`
//	WriteRelabelConfigs []*RelabelConfig `json:"write_relabel_configs,omitempty"`
//
//	// We cannot do proper Go type embedding below as the parser will then parse
//	// values arbitrarily into the overflow maps of further-down types.
//	HTTPClientConfig config_util.HTTPClientConfig `json:",inline"`
//	QueueConfig      QueueConfig                  `json:"queue_config,omitempty"`
//}
//
//// QueueConfig is the configuration for the queue used to write to remote
//// storage.
//type QueueConfig struct {
//	// Number of samples to buffer per shard before we start dropping them.
//	Capacity int `json:"capacity,omitempty"`
//
//	// Max number of shards, i.e. amount of concurrency.
//	MaxShards int `json:"max_shards,omitempty"`
//
//	// Maximum number of samples per send.
//	MaxSamplesPerSend int `json:"max_samples_per_send,omitempty"`
//
//	// Maximum time sample will wait in buffer.
//	BatchSendDeadline Duration `json:"batch_send_deadline,omitempty"`
//
//	// Max number of times to retry a batch on recoverable errors.
//	MaxRetries int `json:"max_retries,omitempty"`
//
//	// On recoverable errors, backoff exponentially.
//	MinBackoff Duration `json:"min_backoff,omitempty"`
//	MaxBackoff Duration `json:"max_backoff,omitempty"`
//}
//
//// RemoteReadConfig is the configuration for reading from remote storage.
//type RemoteReadConfig struct {
//	URL           *config_util.URL `json:"url"`
//	RemoteTimeout Duration         `json:"remote_timeout,omitempty"`
//	ReadRecent    bool             `json:"read_recent,omitempty"`
//	// We cannot do proper Go type embedding below as the parser will then parse
//	// values arbitrarily into the overflow maps of further-down types.
//	HTTPClientConfig config_util.HTTPClientConfig `json:",inline"`
//
//	// RequiredMatchers is an optional list of equality matchers which have to
//	// be present in a selector to query the remote read endpoint.
//	RequiredMatchers model.LabelSet `json:"required_matchers,omitempty"`
//}
//
