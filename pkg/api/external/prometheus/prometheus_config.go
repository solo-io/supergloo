package prometheus

import (
	"bytes"

	jsontoyaml "github.com/ghodss/yaml"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/types"
	"github.com/pkg/errors"
	"github.com/prometheus/prometheus/config"
	v1 "github.com/solo-io/supergloo/pkg/api/external/prometheus/v1"
	yaml "gopkg.in/yaml.v2"
)

func ConfigFromResource(cfg *v1.PrometheusConfig) (*Config, error) {
	if cfg == nil {
		return nil, nil
	}
	buf := &bytes.Buffer{}
	if err := (&jsonpb.Marshaler{OrigName: true}).Marshal(buf, cfg.Prometheus); err != nil {
		return nil, errors.Wrapf(err, "failed to marshal proto struct")
	}
	yam, err := jsontoyaml.JSONToYAML(buf.Bytes())
	if err != nil {
		return nil, errors.Wrapf(err, "converting json to yaml")
	}
	var c Config
	if err := yaml.UnmarshalStrict(yam, &c); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal raw yaml to prometheus config")
	}
	return &c, nil
}

func ConfigToResource(cfg *Config) (*v1.PrometheusConfig, error) {
	if cfg == nil {
		return nil, nil
	}
	yam, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, errors.Wrapf(err, "marshalling to yaml")
	}

	jsn, err := jsontoyaml.YAMLToJSON([]byte(yam))
	if err != nil {
		return nil, errors.Wrapf(err, "converting yaml to json")
	}
	var s types.Struct
	if err := jsonpb.Unmarshal(bytes.NewBuffer(jsn), &s); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal jsn to proto struct")
	}
	return &v1.PrometheusConfig{Prometheus: &s}, nil
}

//type Config config.Config
type Config struct {
	GlobalConfig   *config.GlobalConfig   `yaml:"global"`
	AlertingConfig config.AlertingConfig  `yaml:"alerting,omitempty"`
	RuleFiles      []string               `yaml:"rule_files,omitempty"`
	ScrapeConfigs  []*config.ScrapeConfig `yaml:"scrape_configs,omitempty"`

	RemoteWriteConfigs []*config.RemoteWriteConfig `yaml:"remote_write,omitempty"`
	RemoteReadConfigs  []*config.RemoteReadConfig  `yaml:"remote_read,omitempty"`
}

// returns true if changed
func (cfg *Config) AddScrapeConfigs(scrapeConfigs []*config.ScrapeConfig) bool {
	var updated bool
	for _, desiredScrapeConfig := range scrapeConfigs {
		var found bool
		for _, sc := range cfg.ScrapeConfigs {
			if sc.JobName == desiredScrapeConfig.JobName {
				found = true
				break
			}
		}
		if found {
			continue
		}
		cfg.ScrapeConfigs = append(cfg.ScrapeConfigs, desiredScrapeConfig)
		updated = true
	}
	return updated
}
