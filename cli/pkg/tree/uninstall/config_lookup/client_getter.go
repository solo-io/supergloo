package config_lookup

import (
	"context"

	"github.com/avast/retry-go"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/kubeconfig"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/multicluster"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewDynamicClientGetter(kubeConfigLookup kubeconfig.KubeConfigLookup) multicluster.DynamicClientGetter {
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
