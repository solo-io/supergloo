package prometheus

import (
	"context"
	"strings"

	"github.com/prometheus/prometheus/config"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/pkg/api/custom/clients/prometheus"
	v1 "github.com/solo-io/supergloo/pkg/api/external/prometheus/v1"
)

/*
Ensures the target configmap contains scrape configs
*/
func EnsureScrapeConfigs(ctx context.Context, meshId string, targetConfigmap core.ResourceRef, scrapeConfigs []*config.ScrapeConfig, client v1.PrometheusConfigClient) error {
	// get the existing prometheus config
	currentConfig, err := client.Read(targetConfigmap.Namespace, targetConfigmap.Name, clients.ReadOpts{Ctx: ctx})
	if err != nil {
		return err
	}
	promCfg, err := prometheus.ConfigFromResource(currentConfig)
	if err != nil {
		return err
	}

	// prepend each job with the given mesh id
	var prefixedScrapeConfigs []*config.ScrapeConfig
	// prepend prefix to our jobs
	// this way we can also remove our jobs later
	for _, job := range scrapeConfigs {
		// shallow copy to prevent modifying the input configs
		job := *job
		job.JobName = meshId + "-" + job.JobName
		prefixedScrapeConfigs = append(prefixedScrapeConfigs, &job)
	}

	// update the promcfg
	updated := promCfg.AddScrapeConfigs(prefixedScrapeConfigs)
	if updated == 0 {
		return nil
	}

	contextutils.LoggerFrom(ctx).Infof("added %v: %v scrape configs", targetConfigmap, updated)

	// write to storage
	updatedConfig, err := prometheus.ConfigToResource(promCfg)
	if err != nil {
		return err
	}
	updatedConfig.Metadata = currentConfig.Metadata

	_, err = client.Write(updatedConfig, clients.WriteOpts{Ctx: ctx, OverwriteExisting: true})
	return err
}

func RemoveScrapeConfigs(ctx context.Context, meshId string, targetConfigmap core.ResourceRef, client v1.PrometheusConfigClient) error {
	// get the existing prometheus config
	currentConfig, err := client.Read(targetConfigmap.Namespace, targetConfigmap.Name, clients.ReadOpts{Ctx: ctx})
	if err != nil {
		return err
	}
	promCfg, err := prometheus.ConfigFromResource(currentConfig)
	if err != nil {
		return err
	}

	// filter out jobs with the mesh id prefix
	var notOurJobs []*config.ScrapeConfig
	for _, job := range promCfg.ScrapeConfigs {
		if strings.HasPrefix(job.JobName, meshId+"-") {
			continue
		}
		notOurJobs = append(notOurJobs, job)
	}
	// nothing to delete
	if scrapeConfigsEqual(notOurJobs, promCfg.ScrapeConfigs) {
		return nil
	}

	// overwrite with filtered list
	promCfg.ScrapeConfigs = notOurJobs

	// write to storage
	updatedConfig, err := prometheus.ConfigToResource(promCfg)
	if err != nil {
		return err
	}
	updatedConfig.Metadata = currentConfig.Metadata

	_, err = client.Write(updatedConfig, clients.WriteOpts{Ctx: ctx, OverwriteExisting: true})
	return err
}

func scrapeConfigsEqual(list1, list2 []*config.ScrapeConfig) bool {
	if len(list1) != len(list2) {
		return false
	}
	for i := range list1 {
		if list1[i] != list2[i] {
			return false
		}
	}
	return true
}
