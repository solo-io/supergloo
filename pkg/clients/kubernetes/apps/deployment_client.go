package kubernetes_apps

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type deploymentClient struct {
	client client.Client
}

type DeploymentClientFactory func(client client.Client) DeploymentClient

func DeploymentClientFactoryProvider() DeploymentClientFactory {
	return NewDeploymentClient
}

func NewDeploymentClient(client client.Client) DeploymentClient {
	return &deploymentClient{client: client}
}

func (d *deploymentClient) GetDeployment(ctx context.Context, objectKey client.ObjectKey) (*appsv1.Deployment, error) {
	deployment := &appsv1.Deployment{}
	err := d.client.Get(ctx, objectKey, deployment)
	if err != nil {
		return nil, err
	}
	return deployment, nil
}
