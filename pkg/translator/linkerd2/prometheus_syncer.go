package linkerd2

import (
	"context"
	"go.uber.org/multierr"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/solo-kit/pkg/utils/contextutils"
	"github.com/solo-io/supergloo/pkg/api/external/prometheus"
	prometheusv1 "github.com/solo-io/supergloo/pkg/api/external/prometheus/v1"
	"github.com/solo-io/supergloo/pkg/api/v1"
)

type PrometheusSyncer struct {
	// it's okay for this to be nil, it's only used
	// if the prometheus crd tells us to restart pods
	// if the selector is set and this is nil, we don't have kube support enabled
	Kube             kubernetes.Interface
	PrometheusClient prometheusv1.ConfigClient // for reading/writing configmaps
}

func (s *PrometheusSyncer) Sync(ctx context.Context, snap *v1.TranslatorSnapshot) error {
	ctx = contextutils.WithLogger(ctx, "prometheus-syncer")
	logger := contextutils.LoggerFrom(ctx)
	logger.Infof("begin sync %v (%v meshes, %v upstreams)", snap.Hash(),
		len(snap.Meshes), len(snap.Upstreams))
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
	if mesh.TargetMesh == nil {
		// ilackarms (todo): reporting for this error
		return errors.Errorf("invalid mesh %v: target_mesh required", mesh.Metadata.Ref())
	}
	if mesh.TargetMesh.MeshType != v1.MeshType_LINKERD2 {
		return nil
	}
	if mesh.Observability == nil {
		return nil
	}
	if mesh.Observability.Prometheus == nil {
		return nil
	}
	if !mesh.Observability.Prometheus.EnableMetrics {
		return nil
	}
	if mesh.Observability.Prometheus.PrometheusConfigMap == nil {
		// ilackarms (todo): reporting for this error
		return errors.Errorf("invalid mesh %v: must provide a reference to the target prometheus config map", mesh.Metadata.Ref())
	}
	configMap := *mesh.Observability.Prometheus.PrometheusConfigMap
	prometheusConfig, err := s.getPrometheusConfig(ctx, configMap)
	if err != nil {
		return errors.Wrapf(err, "retrieving existing prometheus config")
	}

	// TODO (ilackarms): make this syncer take scrape configs as an argument
	changed := prometheusConfig.AddScrapeConfigs(LinkerdScrapeConfigs)
	if !changed {
		return nil
	}

	contextutils.LoggerFrom(ctx).Infof("syncing prometheus config for mesh %v", mesh.Metadata.Ref())

	if err := s.writePrometheusConfig(ctx, configMap, prometheusConfig); err != nil {
		return errors.Wrapf(err, "writing prometheus config")
	}
	// no pod labels specified, nothing to restart
	if len(mesh.Observability.Prometheus.PodLabels) < 1 {
		return nil
	}

	selector := mesh.Observability.Prometheus.PodLabels

	// got this far, it means we're on kube and they want us to restart pods
	if err := s.restartPods(ctx, configMap.Namespace, selector); err != nil {
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

// TODO (ilackarms / rickducott): generalize this function

func (s *PrometheusSyncer) restartPods(ctx context.Context, namespace string, selector map[string]string) error {
	if s.Kube == nil {
		return errors.Errorf("kubernetes suppport is currently disabled. see SuperGloo documentation" +
			" for utilizing pod restarts")
	}
	if err := s.Kube.CoreV1().Pods(namespace).DeleteCollection(nil, metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(selector).String(),
	}); err != nil {
		return errors.Wrapf(err, "restarting pods with selector %v", selector)
	}
	return nil
}
