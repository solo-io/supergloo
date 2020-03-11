package kubernetes_apps

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type controllerRuntimeDeploymentClient struct {
	client client.Client
}

type DeploymentClientFactory func(client client.Client) DeploymentClient

func ControllerRuntimeDeploymentClientFactoryProvider() DeploymentClientFactory {
	return NewControllerRuntimeDeploymentClient
}

type GeneratedDeploymentClientFactory func(client kubernetes.Interface) DeploymentClient

func GeneratedDeploymentClientFactoryProvider() GeneratedDeploymentClientFactory {
	return NewGeneratedDeploymentClient
}

func NewControllerRuntimeDeploymentClient(client client.Client) DeploymentClient {
	return &controllerRuntimeDeploymentClient{client: client}
}

func (d *controllerRuntimeDeploymentClient) Get(ctx context.Context, objectKey client.ObjectKey) (*appsv1.Deployment, error) {
	deployment := &appsv1.Deployment{}
	err := d.client.Get(ctx, objectKey, deployment)
	if err != nil {
		return nil, err
	}
	return deployment, nil
}

func NewGeneratedDeploymentClient(client kubernetes.Interface) DeploymentClient {
	return &generatedDeploymentClient{client: client}
}

type generatedDeploymentClient struct {
	client kubernetes.Interface
}

func (g *generatedDeploymentClient) Get(ctx context.Context, objectKey client.ObjectKey) (*appsv1.Deployment, error) {
	return g.client.AppsV1().Deployments(objectKey.Namespace).Get(objectKey.Name, v1.GetOptions{})
}
