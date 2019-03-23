package prometheus

import (
	"github.com/pkg/errors"
	"github.com/prometheus/prometheus/config"
	v1 "github.com/solo-io/supergloo/pkg/api/external/prometheus/v1"

	// note: this yaml library is required (ghodss/yaml will not work)
	// this is because the prometheus structs use yaml tags rather than json
	// in the struct annotations
	"gopkg.in/yaml.v2"
)

func ConfigFromResource(cfg *v1.PrometheusConfig) (*Config, error) {
	if cfg == nil {
		return nil, nil
	}
	var c Config
	if err := yaml.Unmarshal([]byte(cfg.Prometheus), &c); err != nil {
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
	return &v1.PrometheusConfig{Prometheus: string(yam)}, nil
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
func (cfg *Config) AddScrapeConfigs(scrapeConfigs []*config.ScrapeConfig) int {
	var updated int
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
		updated++
	}
	return updated
}
