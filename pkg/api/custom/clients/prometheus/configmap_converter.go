package prometheus

import (
	"context"

	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/go-utils/protoutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/configmap"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/utils/kubeutils"
	v1 "github.com/solo-io/supergloo/pkg/api/external/prometheus/v1"
	kubev1 "k8s.io/api/core/v1"
)

const prometheusConfigmapKey = "prometheus.yml"

type prometheusConfigmapConverter struct{}

func NewPrometheusConfigmapConverter() configmap.ConfigMapConverter {
	return &prometheusConfigmapConverter{}
}

func (c *prometheusConfigmapConverter) FromKubeConfigMap(ctx context.Context, rc *configmap.ResourceClient, configMap *kubev1.ConfigMap) (resources.Resource, error) {
	resource := rc.NewResource()
	// we only care about prometheus configs
	if _, isPrometheusConfig := configMap.Data[prometheusConfigmapKey]; !isPrometheusConfig {
		return nil, nil
	}
	// only works for string fields
	resourceMap := make(map[string]interface{})
	for k, v := range configMap.Data {
		resourceMap[k] = v
	}

	if err := protoutils.UnmarshalMap(resourceMap, resource); err != nil {
		return nil, errors.Wrapf(err, "reading configmap data into %v", rc.Kind())
	}
	resource.SetMetadata(kubeutils.FromKubeMeta(configMap.ObjectMeta))

	return resource, nil
}

func (c *prometheusConfigmapConverter) ToKubeConfigMap(ctx context.Context, rc *configmap.ResourceClient, resource resources.Resource) (*kubev1.ConfigMap, error) {
	if _, isPrometheusConfig := resource.(*v1.PrometheusConfig); !isPrometheusConfig {
		return nil, errors.Errorf("cannot convert %v to configmap", resources.Kind(resource))
	}
	resourceMap, err := protoutils.MarshalMapEmitZeroValues(resource)
	if err != nil {
		return nil, errors.Wrapf(err, "marshalling resource as map")
	}
	configMapData := make(map[string]string)
	for k, v := range resourceMap {
		// metadata comes from ToKubeMeta
		// status not supported
		if k == "metadata" {
			continue
		}
		switch val := v.(type) {
		case string:
			configMapData[k] = val
		}
	}
	meta := kubeutils.ToKubeMeta(resource.GetMetadata())
	return &kubev1.ConfigMap{
		ObjectMeta: meta,
		Data:       configMapData,
	}, nil
}
