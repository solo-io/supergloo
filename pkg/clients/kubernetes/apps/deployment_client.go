package kubernetes_apps

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type deploymentClient struct {
	client client.Client
}

type DeploymentClientFactory func(client client.Client) DeploymentClient

func ControllerRuntimeDeploymentClientFactoryProvider() DeploymentClientFactory {
	return NewDeploymentClient
}

type DeploymentClientFactoryForConfig func(cfg *rest.Config) (DeploymentClient, error)

func DeploymentClientFactoryForConfigProvider() DeploymentClientFactoryForConfig {
	return NewDeploymentClientForConfig
}

func NewDeploymentClientForConfig(cfg *rest.Config) (DeploymentClient, error) {
	dynamicClient, err := client.New(cfg, client.Options{})
	if err != nil {
		return nil, err
	}
	return &deploymentClient{client: dynamicClient}, nil
}

func NewDeploymentClient(client client.Client) DeploymentClient {
	return &deploymentClient{client: client}
}

func (d *deploymentClient) Get(ctx context.Context, objectKey client.ObjectKey) (*appsv1.Deployment, error) {
	deployment := &appsv1.Deployment{}
	err := d.client.Get(ctx, objectKey, deployment)
	if err != nil {
		return nil, err
	}
	return deployment, nil
}

func (c *deploymentClient) List(ctx context.Context, options ...client.ListOption) (*appsv1.DeploymentList, error) {
	deploymentList := appsv1.DeploymentList{}
	err := c.client.List(ctx, &deploymentList, options...)
	if err != nil {
		return &deploymentList, err
	}
	return &deploymentList, nil
}
