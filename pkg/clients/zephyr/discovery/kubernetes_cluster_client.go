package zephyr_discovery

import (
	"context"

	discovery_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/clients"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewKubernetesClusterClientForConfig(cfg *rest.Config) (KubernetesClusterClient, error) {
	if err := discovery_v1alpha1.AddToScheme(scheme.Scheme); err != nil {
		return nil, err
	}
	dynamicClient, err := client.New(cfg, client.Options{})
	if err != nil {
		return nil, err
	}
	return &kubernetesClusterClient{client: dynamicClient}, nil
}

func NewKubernetesClusterClient(client client.Client) KubernetesClusterClient {
	return &kubernetesClusterClient{client}
}

type kubernetesClusterClient struct {
	client client.Client
}

func (c *kubernetesClusterClient) Create(ctx context.Context, cluster *discovery_v1alpha1.KubernetesCluster) error {
	return c.client.Create(ctx, cluster)
}

func (c *kubernetesClusterClient) Update(ctx context.Context, cluster *discovery_v1alpha1.KubernetesCluster) error {
	return c.client.Update(ctx, cluster)
}

func (c *kubernetesClusterClient) Upsert(ctx context.Context, cluster *discovery_v1alpha1.KubernetesCluster) error {
	existing := &discovery_v1alpha1.KubernetesCluster{}
	err := c.client.Get(ctx, clients.ObjectMetaToObjectKey(cluster.ObjectMeta), existing)
	if errors.IsNotFound(err) {
		return c.Create(ctx, cluster)
	} else if err != nil {
		return err
	}

	existing.Spec = cluster.Spec
	return c.client.Update(ctx, existing)
}

func (c *kubernetesClusterClient) Get(ctx context.Context, key client.ObjectKey) (*discovery_v1alpha1.KubernetesCluster, error) {
	existing := &discovery_v1alpha1.KubernetesCluster{}
	err := c.client.Get(ctx, key, existing)
	return existing, err
}

func (c *kubernetesClusterClient) List(ctx context.Context, opts ...client.ListOption) (*discovery_v1alpha1.KubernetesClusterList, error) {
	list := &discovery_v1alpha1.KubernetesClusterList{}
	err := c.client.List(ctx, list, opts...)
	return list, err
}
