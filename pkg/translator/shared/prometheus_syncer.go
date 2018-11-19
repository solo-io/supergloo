package shared

import (
	"context"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/solo-kit/pkg/utils/contextutils"
	"github.com/solo-io/supergloo/pkg/api/external/prometheus"
	prometheusv1 "github.com/solo-io/supergloo/pkg/api/external/prometheus/v1"
	"github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/translator/kube"
	"go.uber.org/multierr"
	"k8s.io/client-go/kubernetes"
)

type MeshWithPrometheus interface {
	GetPrometheusConfigmap() *core.ResourceRef
}

type PrometheusSyncer struct {
	// it's okay for this to be nil, it's only used
	// if the prometheus crd tells us to restart pods
	// if the selector is set and this is nil, we don't have kube support enabled
	Kube                 kubernetes.Interface
	PrometheusClient     prometheusv1.ConfigClient // for reading/writing configmaps
	DesiredScrapeConfigs []prometheus.ScrapeConfig
	// the implementing syncer puts in a function to return the configmap resource ref
	// if the configmap ref is nil, we return, assume there is no work for us to do on this mesh
	GetConfigMap func(*v1.Mesh) *core.ResourceRef
}

func (s *PrometheusSyncer) Sync(ctx context.Context, snap *v1.TranslatorSnapshot) error {
	ctx = contextutils.WithLogger(ctx, "prometheus-syncer")
	logger := contextutils.LoggerFrom(ctx)
	meshes := snap.Meshes.List()
	upstreams := snap.Upstreams.List()

	logger.Infof("begin sync %v (%v meshes)", snap.Hash(),
		len(meshes), len(upstreams))
	defer logger.Infof("end sync %v", snap.Hash())
	logger.Debugf("%v", snap)

	var errs error
	for _, mesh := range snap.Meshes.List() {
		logger.Infof("syncing mesh %v", mesh.Metadata.Ref())
		if err := s.syncMesh(ctx, mesh); err != nil {
			errs = multierr.Append(errs, errors.Wrapf(err, "syncing mesh %v failed", mesh.Metadata.Ref()))
			continue
		}
	}
	return errs
}

func (s *PrometheusSyncer) syncMesh(ctx context.Context, mesh *v1.Mesh) error {
	if mesh.Observability == nil {
		return nil
	}
	if mesh.Observability.Prometheus == nil {
		return nil
	}
	if !mesh.Observability.Prometheus.EnableMetrics {
		return nil
	}

	configMap := s.GetConfigMap(mesh)
	if configMap == nil {
		// nothing to configure for this mesh
		return nil
	}
	prometheusConfig, err := s.getPrometheusConfig(ctx, *configMap)
	if err != nil {
		return errors.Wrapf(err, "retrieving existing prometheus config")
	}

	// TODO (ilackarms): make this syncer take scrape configs as an argument
	changed := prometheusConfig.AddScrapeConfigs(s.DesiredScrapeConfigs)
	if !changed {
		return nil
	}

	contextutils.LoggerFrom(ctx).Infof("syncing prometheus config for mesh %v", mesh.Metadata.Ref())

	if err := s.writePrometheusConfig(ctx, *configMap, prometheusConfig); err != nil {
		return errors.Wrapf(err, "writing prometheus config")
	}
	// no pod labels specified, nothing to restart
	if len(mesh.Observability.Prometheus.PodLabels) < 1 {
		return nil
	}

	selector := mesh.Observability.Prometheus.PodLabels

	// got this far, it means we're on kube and they want us to restart pods
	if err := kube.RestartPods(s.Kube, configMap.Namespace, selector); err != nil {
		return errors.Wrapf(err, "restarting prometheus pods")
	}

	return nil
}

func (s *PrometheusSyncer) getPrometheusConfig(ctx context.Context, ref core.ResourceRef) (*prometheus.PrometheusConfig, error) {
	cfg, err := s.PrometheusClient.Read(ref.Namespace, ref.Name, clients.ReadOpts{
		Ctx: ctx,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "reading prometheus config from %v", ref)
	}
	promCfg, err := prometheus.ConfigFromResource(cfg)
	if err != nil {
		return nil, errors.Wrapf(err, "parsing prometheus config from %v", ref)
	}
	return promCfg, nil
}

func (s *PrometheusSyncer) writePrometheusConfig(ctx context.Context, ref core.ResourceRef, cfg *prometheus.PrometheusConfig) error {
	originalCfg, err := s.PrometheusClient.Read(ref.Namespace, ref.Name, clients.ReadOpts{
		Ctx: ctx,
	})
	if err != nil {
		return errors.Wrapf(err, "fetching prometheus config from %v for update", ref)
	}
	desiredCfg, err := prometheus.ConfigToResource(cfg)
	if err != nil {
		return errors.Wrapf(err, "converting prometheus config to resource %v", ref)
	}
	desiredCfg.SetMetadata(originalCfg.Metadata)
	if _, err := s.PrometheusClient.Write(desiredCfg, clients.WriteOpts{Ctx: ctx, OverwriteExisting: true}); err != nil {
		return errors.Wrapf(err, "updating prometheus configmap")
	}
	return nil
}
