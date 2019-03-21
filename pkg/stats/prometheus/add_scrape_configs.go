package prometheus

import (
	"context"

	"github.com/prometheus/prometheus/config"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/pkg/api/custom/clients/prometheus"
	v1 "github.com/solo-io/supergloo/pkg/api/external/prometheus/v1"
)

/*
Syncs a prometheus configmap, ensuring that each provided scrape config
Is present in the target configmap
*/
type PrometheusConfigUpdater struct {
	targetConfigmap core.ResourceRef
	scrapeConfigs   []*config.ScrapeConfig
	client          v1.PrometheusConfigClient
}

func NewPrometheusConfigUpdater(targetConfigmap core.ResourceRef, scrapeConfigs []*config.ScrapeConfig, client v1.PrometheusConfigClient) *PrometheusConfigUpdater {
	return &PrometheusConfigUpdater{targetConfigmap: targetConfigmap, scrapeConfigs: scrapeConfigs, client: client}
}

func (s *PrometheusConfigUpdater) EnsureScrapeConfigs(ctx context.Context) error {
	oldConfig, err := s.client.Read(s.targetConfigmap.Namespace, s.targetConfigmap.Name, clients.ReadOpts{Ctx: ctx})
	if err != nil {
		return err
	}
	promCfg, err := prometheus.ConfigFromResource(oldConfig)
	if err != nil {
		return err
	}
	updated := promCfg.AddScrapeConfigs(s.scrapeConfigs)
	if !updated {
		return nil
	}
	updatedConfig, err := prometheus.ConfigToResource(promCfg)
	if err != nil {
		return err
	}
	updatedConfig.Metadata = oldConfig.Metadata
	contextutils.LoggerFrom(ctx).Infof("updating configmap %v with %v scrape configs", s.targetConfigmap, len(s.scrapeConfigs))
	_, err = s.client.Write(updatedConfig, clients.WriteOpts{Ctx: ctx, OverwriteExisting: true})
	return err
}
