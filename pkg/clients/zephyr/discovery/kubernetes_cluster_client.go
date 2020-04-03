package zephyr_discovery

import (
	"context"

	discovery_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1/clientset/versioned"
	discovery_types "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1/clientset/versioned/typed/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/mesh-projects/pkg/clients"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewGeneratedKubernetesClusterClient(config *rest.Config) (KubernetesClusterClient, error) {
	clientSet, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &kubernetesClusterClient{clientSet: clientSet.DiscoveryV1alpha1()}, nil
}

type kubernetesClusterClient struct {
	clientSet discovery_types.DiscoveryV1alpha1Interface
}

func (k *kubernetesClusterClient) Get(ctx context.Context, key client.ObjectKey) (*discovery_v1alpha1.KubernetesCluster, error) {
	return k.clientSet.KubernetesClusters(key.Namespace).Get(key.Name, metav1.GetOptions{})
}

func (k *kubernetesClusterClient) List(
	ctx context.Context,
	opts ...client.ListOption,
) (*discovery_v1alpha1.KubernetesClusterList, error) {
	listOptions := &client.ListOptions{}
	for _, v := range opts {
		v.ApplyToList(listOptions)
	}
	raw := metav1.ListOptions{}
	if converted := listOptions.AsListOptions(); converted != nil {
		raw = *converted
	}
	return k.clientSet.KubernetesClusters(listOptions.Namespace).List(raw)
}

func (k *kubernetesClusterClient) Create(ctx context.Context, cluster *discovery_v1alpha1.KubernetesCluster) error {
	existing, err := k.clientSet.KubernetesClusters(cluster.GetNamespace()).Create(cluster)
	if err != nil {
		return err
	}
	*cluster = *existing
	return nil
}

func (k *kubernetesClusterClient) Update(ctx context.Context, cluster *discovery_v1alpha1.KubernetesCluster) error {
	existing, err := k.clientSet.KubernetesClusters(cluster.GetNamespace()).Update(cluster)
	if err != nil {
		return err
	}
	*cluster = *existing
	return nil
}

func (k *kubernetesClusterClient) Upsert(ctx context.Context, cluster *discovery_v1alpha1.KubernetesCluster) error {
	err := k.Create(ctx, cluster)
	if err != nil {
		if !errors.IsAlreadyExists(err) {
			return err
		}
		existing, err := k.Get(ctx, client.ObjectKey{
			Namespace: cluster.GetNamespace(),
			Name:      cluster.GetName(),
		})
		if err != nil {
			return err
		}
		existing.Spec = cluster.Spec
		return k.Update(ctx, existing)
	}
	return nil
}

func NewControllerRuntimeKubernetesClusterClient(client client.Client) KubernetesClusterClient {
	return &controllerRuntimeKubernetesClusterClient{client}
}

type controllerRuntimeKubernetesClusterClient struct {
	client client.Client
}

func (c *controllerRuntimeKubernetesClusterClient) Create(ctx context.Context, cluster *discovery_v1alpha1.KubernetesCluster) error {
	return c.client.Create(ctx, cluster)
}

func (c *controllerRuntimeKubernetesClusterClient) Update(ctx context.Context, cluster *discovery_v1alpha1.KubernetesCluster) error {
	return c.client.Update(ctx, cluster)
}

func (c *controllerRuntimeKubernetesClusterClient) Upsert(ctx context.Context, cluster *discovery_v1alpha1.KubernetesCluster) error {
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

func (c *controllerRuntimeKubernetesClusterClient) Get(ctx context.Context, key client.ObjectKey) (*discovery_v1alpha1.KubernetesCluster, error) {
	existing := &discovery_v1alpha1.KubernetesCluster{}
	err := c.client.Get(ctx, key, existing)
	return existing, err
}

func (c *controllerRuntimeKubernetesClusterClient) List(ctx context.Context, opts ...client.ListOption) (*discovery_v1alpha1.KubernetesClusterList, error) {
	list := &discovery_v1alpha1.KubernetesClusterList{}
	err := c.client.List(ctx, list, opts...)
	return list, err
}
