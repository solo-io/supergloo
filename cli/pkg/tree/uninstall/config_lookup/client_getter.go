package config_lookup

import (
	"context"

	"github.com/avast/retry-go"
	"github.com/solo-io/service-mesh-hub/pkg/kube/kubeconfig"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/compute-target/k8s"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewDynamicClientGetter(kubeConfigLookup kubeconfig.KubeConfigLookup) mc_manager.DynamicClientGetter {
	return &dynamicClientGetter{kubeConfigLookup: kubeConfigLookup}
}

type dynamicClientGetter struct {
	kubeConfigLookup kubeconfig.KubeConfigLookup
}

func (d *dynamicClientGetter) GetClientForCluster(ctx context.Context, clusterName string, opts ...retry.Option) (client.Client, error) {
	config, err := d.kubeConfigLookup.FromCluster(ctx, clusterName)
	if err != nil {
		return nil, err
	}

	return client.New(config.RestConfig, client.Options{})
}
