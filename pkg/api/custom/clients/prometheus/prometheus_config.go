package prometheus

import (
	"sort"
	"strings"

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

func sortConfigs(scrapeConfigs []*config.ScrapeConfig) {
	sort.SliceStable(scrapeConfigs, func(i, j int) bool {
		return scrapeConfigs[i].JobName < scrapeConfigs[j].JobName
	})
}

// returns number of added
func (cfg *Config) AddScrapeConfigs(scrapeConfigs []*config.ScrapeConfig) int {
	var added int
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
		added++
	}
	sortConfigs(cfg.ScrapeConfigs)
	return added
}

// returns number of removed
func (cfg *Config) RemoveScrapeConfigs(namePrefix string) int {
	var removed int

	// filter out jobs with the name prefix
	var filteredJobs []*config.ScrapeConfig
	for _, job := range cfg.ScrapeConfigs {
		if strings.HasPrefix(job.JobName, namePrefix) {
			removed++
			continue
		}
		filteredJobs = append(filteredJobs, job)
	}
	cfg.ScrapeConfigs = filteredJobs
	sortConfigs(cfg.ScrapeConfigs)
	return removed
}

func AddPrefix(scrapeConfigs []*config.ScrapeConfig, prefix string) []*config.ScrapeConfig {
	// prepend each job with the given mesh id
	var prefixedScrapeConfigs []*config.ScrapeConfig
	// prepend prefix to our jobs
	// this way we can also remove our jobs later
	for _, job := range scrapeConfigs {
		// shallow copy to prevent modifying the input configs
		job := *job
		job.JobName = prefix + job.JobName
		prefixedScrapeConfigs = append(prefixedScrapeConfigs, &job)
	}
	return prefixedScrapeConfigs
}
