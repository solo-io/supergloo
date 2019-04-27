package prometheus

import (
	"encoding/json"
	"reflect"
	"sort"
	"strings"

	"github.com/prometheus/prometheus/config"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
)

type PrometheusConfig struct {
	Metadata core.Metadata
	Config
	Alerts string // inline as a string, not processed by supergloo
	Rules  string // inline as a string, not processed by supergloo
}

func (c *PrometheusConfig) GetMetadata() core.Metadata {
	return c.Metadata
}

func (c *PrometheusConfig) SetMetadata(meta core.Metadata) {
	c.Metadata = meta
}

func (c *PrometheusConfig) Equal(that interface{}) bool {
	return reflect.DeepEqual(c, that)
}

func (c *PrometheusConfig) Clone() *PrometheusConfig {
	raw, err := json.Marshal(c)
	if err != nil {
		panic(err)
	}
	var newC PrometheusConfig
	if err := json.Unmarshal(raw, &newC); err != nil {
		panic(err)
	}
	return &newC
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

func SortConfigs(scrapeConfigs []*config.ScrapeConfig) {
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
	SortConfigs(cfg.ScrapeConfigs)
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
	SortConfigs(cfg.ScrapeConfigs)
	return removed
}
