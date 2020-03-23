package kubernetes_core

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ConfigMapClientFactory func(client.Client) ConfigMapClient

func ConfigMapClientFactoryProvider() ConfigMapClientFactory {
	return NewConfigMapClient
}

func NewConfigMapClient(client client.Client) ConfigMapClient {
	return &configMapClient{
		client: client,
	}
}

type configMapClient struct {
	client client.Client
}

func (g *configMapClient) Create(ctx context.Context, configMap *corev1.ConfigMap) error {
	return g.client.Create(ctx, configMap)
}

func (g *configMapClient) Get(ctx context.Context, objKey client.ObjectKey) (*corev1.ConfigMap, error) {
	ConfigMap := corev1.ConfigMap{}
	err := g.client.Get(ctx, objKey, &ConfigMap)
	if err != nil {
		return nil, err
	}

	return &ConfigMap, nil
}

func (g *configMapClient) Update(ctx context.Context, configMap *corev1.ConfigMap) error {
	return g.client.Update(ctx, configMap)
}
