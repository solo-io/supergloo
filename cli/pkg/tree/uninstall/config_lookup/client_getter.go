package config_lookup

import (
	"context"

	"github.com/avast/retry-go"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/mesh-platform/k8s"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewDynamicClientGetter(kubeConfigLookup KubeConfigLookup) mc_manager.DynamicClientGetter {
	return &dynamicClientGetter{kubeConfigLookup: kubeConfigLookup}
}

type dynamicClientGetter struct {
	kubeConfigLookup KubeConfigLookup
}

func (d *dynamicClientGetter) GetClientForCluster(ctx context.Context, clusterName string, opts ...retry.Option) (client.Client, error) {
	config, err := d.kubeConfigLookup.FromCluster(ctx, clusterName)
	if err != nil {
		return nil, err
	}

	return client.New(config.RestConfig, client.Options{})
}
