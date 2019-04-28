package prometheus

import (
	"context"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	"github.com/solo-io/supergloo/api/external/prometheus"
	prometheusv1 "github.com/solo-io/supergloo/pkg/api/external/prometheus/v1"
	"gopkg.in/yaml.v2"
	"k8s.io/client-go/kubernetes"

	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/configmap"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/utils/kubeutils"
	kubev1 "k8s.io/api/core/v1"
)

const configKey1 = "prometheus.yml"
const configKey2 = "prometheus.yaml"
const alertsKey = "rules"
const rulesKey = "rules"
const keyUsedAnnotation = "cfg_key_name"

type prometheusConfigmapConverter struct{}

func NewPrometheusConfigmapConverter() configmap.ConfigMapConverter {
	return &prometheusConfigmapConverter{}
}

func (c *prometheusConfigmapConverter) FromKubeConfigMap(ctx context.Context, rc *configmap.ResourceClient, configMap *kubev1.ConfigMap) (resources.Resource, error) {
	keyUsed := configKey1
	promYaml, ok := configMap.Data[configKey1]
	if !ok {
		keyUsed = configKey2
		promYaml, ok = configMap.Data[configKey2]
		if !ok {
			// not our resource
			return nil, nil
		}
	}
	alerts := configMap.Data[alertsKey]
	rules := configMap.Data[rulesKey]

	var cfg prometheus.Config
	if err := yaml.Unmarshal([]byte(promYaml), &cfg); err != nil {
		return nil, err
	}

	meta := kubeutils.FromKubeMeta(configMap.ObjectMeta)
	if meta.Annotations == nil {
		meta.Annotations = map[string]string{}
	}
	meta.Annotations[keyUsedAnnotation] = keyUsed

	return &prometheusv1.PrometheusConfig{PrometheusConfig: prometheus.PrometheusConfig{
		Metadata: meta,
		Config:   cfg,
		Rules:    rules,
		Alerts:   alerts,
	}}, nil
}

func (c *prometheusConfigmapConverter) ToKubeConfigMap(ctx context.Context, rc *configmap.ResourceClient, resource resources.Resource) (*kubev1.ConfigMap, error) {
	promCfg, ok := resource.(*prometheusv1.PrometheusConfig)
	if !ok {
		return nil, errors.Errorf("%T not type %T cannot convert", resource, &prometheusv1.PrometheusConfig{})
	}
	configYml, err := yaml.Marshal(promCfg.Config)
	if err != nil {
		return nil, err
	}
	// default to key prometheus.yml
	keyUsed := configKey1
	if promCfg.Metadata.Annotations != nil {
		key, ok := promCfg.Metadata.Annotations[keyUsedAnnotation]
		if ok {
			keyUsed = key
		}
	}
	data := map[string]string{
		keyUsed: string(configYml),
	}
	if promCfg.Alerts != "" {
		data[alertsKey] = promCfg.Alerts
	}
	if promCfg.Rules != "" {
		data[rulesKey] = promCfg.Rules
	}

	meta := kubeutils.ToKubeMeta(resource.GetMetadata())
	return &kubev1.ConfigMap{
		ObjectMeta: meta,
		Data:       data,
	}, nil
}

func ResourceClientFactory(kube kubernetes.Interface, kubeCache cache.KubeCoreCache) factory.ResourceClientFactory {
	return &factory.KubeConfigMapClientFactory{
		Clientset:       kube,
		Cache:           kubeCache,
		PlainConfigmaps: true,
		CustomConverter: NewPrometheusConfigmapConverter(),
	}
}
