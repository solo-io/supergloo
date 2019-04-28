package prometheus

import (
	"github.com/prometheus/prometheus/config"
)

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
