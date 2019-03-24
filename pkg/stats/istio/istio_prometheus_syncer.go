package istio

import (
	"context"
	"fmt"

	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/go-utils/hashutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/supergloo/pkg/api/custom/clients/prometheus"

	"github.com/prometheus/prometheus/config"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	prometheusv1 "github.com/solo-io/supergloo/pkg/api/external/prometheus/v1"

	"github.com/solo-io/go-utils/contextutils"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

// registration-level syncer

type istioPrometheusSyncer struct {
	client prometheusv1.PrometheusConfigClient
}

func NewIstioPrometheusSyncer(client prometheusv1.PrometheusConfigClient) *istioPrometheusSyncer {
	return &istioPrometheusSyncer{client: client}
}

// Ensure all prometheus configs contain scrape configs for the meshes which target them
//
// TODO (ilackarms): figure out a way to have multiple syncers report on the same resource
// currently istioPrometheusSyncer returns errors instead of reporting
func (s *istioPrometheusSyncer) Sync(ctx context.Context, snap *v1.RegistrationSnapshot) error {
	ctx = contextutils.WithLogger(ctx, fmt.Sprintf("istio-stats-syncer-%v", snap.Hash()))
	logger := contextutils.LoggerFrom(ctx)
	logger.Infof("begin sync %v", snap.Stringer())
	defer logger.Infof("end sync %v", snap.Stringer())

	// a map of the prometheus configmap to the meshes it should scrape
	promConfigsWithMeshes := getPromCfgsWithMeshes(snap.Meshes.List())

	logger.Infof("syncing %v prometheus configs", len(promConfigsWithMeshes))

	if err := s.syncPrometheusConfigsWithMeshes(ctx, promConfigsWithMeshes); err != nil {
		return errors.Wrapf(err, "failed syncing prometheus configs with mesh scrape jobs")
	}

	return nil
}

// a simple tuple containing the
// istio mesh spec + the resource ref it came from
type istioMesh struct {
	ref   core.ResourceRef
	istio *v1.IstioMesh
}

// get all the prometheus configs that must scrape one or more meshes
func getPromCfgsWithMeshes(meshes v1.MeshList) map[core.ResourceRef][]istioMesh {

	// a map of the prometheus configmap to the meshes it should scrape
	promConfigsWithMeshes := make(map[core.ResourceRef][]istioMesh)

	for _, mesh := range meshes {
		// we only care about istio meshes
		istio := mesh.GetIstio()
		if istio == nil {
			continue
		}

		// we only care about meshes with monitoring configs
		monitoring := mesh.GetMonitoringConfig()
		if monitoring == nil {
			continue
		}
		// we only care about meshes with prometheus config stores
		promCfg := monitoring.GetPrometheusConfigmap()
		if promCfg == nil {
			continue
		}

		// add this mesh to the set this prometheus should scrape
		promConfigsWithMeshes[*promCfg] = append(promConfigsWithMeshes[*promCfg], istioMesh{
			ref:   mesh.Metadata.Ref(),
			istio: istio,
		})
	}

	return promConfigsWithMeshes
}

const superGlooScrapePrefix = "supergloo"

func scrapePrefix(mesh core.ResourceRef, meshType string) string {
	return fmt.Sprintf("%v-%v-%v-", superGlooScrapePrefix, meshType, mesh.Key())
}

func (s *istioPrometheusSyncer) syncPrometheusConfigsWithMeshes(ctx context.Context, configsWithMeshes map[core.ResourceRef][]istioMesh) error {
	// list all configs
	allPromConfigs, err := s.client.List("", clients.ListOpts{Ctx: ctx})
	if err != nil {
		return err
	}

	contextutils.LoggerFrom(ctx).Debugf("syncing prometheus configs with"+
		" mesh registrations: %#v", configsWithMeshes)

	// sync each prom config with the jobs it needs
	for _, originalCfg := range allPromConfigs {
		cfgRef := originalCfg.Metadata.Ref()

		// get the scrape jobs we want to ensure for this config
		meshesForThisConfig := configsWithMeshes[cfgRef]

		// generate scrape configs
		var scrapeConfigsToAdd []*config.ScrapeConfig
		for _, mesh := range meshesForThisConfig {
			scrapesForMesh, err := PrometheusScrapeConfigs(mesh.istio.InstallationNamespace)
			if err != nil {
				return errors.Wrapf(err, "internal error: generating scrape configs for mesh %v", mesh.ref)
			}
			// use mesh key as job prefix
			scrapesForMesh = prometheus.AddPrefix(scrapesForMesh, scrapePrefix(mesh.ref, "istio"))
			scrapeConfigsToAdd = append(scrapesForMesh, scrapeConfigsToAdd...)
		}

		// render config struct from solo kit resource
		promCfg, err := prometheus.ConfigFromResource(originalCfg)
		if err != nil {
			return errors.Wrapf(err, "internal error: failed converting %v to prometheus config. cannot sync",
				cfgRef)
		}

		// remove all scrape configs and start fresh
		promCfg.RemoveScrapeConfigs(superGlooScrapePrefix)

		// add all scrape configs
		added := promCfg.AddScrapeConfigs(scrapeConfigsToAdd)

		// compare with duplicate of original config, only update if diff
		// TODO (ilackarms): investigate a better way to duplicate promCfgs
		originalPromCfg, err := prometheus.ConfigFromResource(originalCfg)
		if err != nil {
			return errors.Wrapf(err, "internal error: failed converting %v to prometheus config. cannot sync",
				cfgRef)
		}
		if hashutils.HashAll(promCfg) == hashutils.HashAll(originalPromCfg) {
			return nil
		}

		// create a configmap from the new prom cfg and save it to storage
		promConfigMap, err := prometheus.ConfigToResource(promCfg)
		if err != nil {
			return errors.Wrapf(err, "internal error: failed converting %v from prometheus config to configmap. cannot sync",
				cfgRef)
		}

		// copy metadata for writing
		promConfigMap.Metadata = originalCfg.Metadata

		contextutils.LoggerFrom(ctx).Infof("prometheus %v syncing %v prometheus scrape configs", cfgRef.Key(), added)

		if _, err := s.client.Write(promConfigMap, clients.WriteOpts{Ctx: ctx, OverwriteExisting: true}); err != nil {
			return errors.Wrapf(err, "writing updated prometheus config to storage")
		}
	}
	return nil
}
