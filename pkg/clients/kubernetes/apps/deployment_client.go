package kubernetes_apps

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type controllerRuntimeDeploymentClient struct {
	client client.Client
}

type DeploymentClientFactory func(client client.Client) DeploymentClient

func ControllerRuntimeDeploymentClientFactoryProvider() DeploymentClientFactory {
	return NewControllerRuntimeDeploymentClient
}

type GeneratedDeploymentClientFactory func(cfg *rest.Config) (DeploymentClient, error)

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

func (c *controllerRuntimeDeploymentClient) List(ctx context.Context, options ...client.ListOption) (*appsv1.DeploymentList, error) {
	deploymentList := appsv1.DeploymentList{}
	err := c.client.List(ctx, &deploymentList, options...)
	if err != nil {
		return &deploymentList, err
	}
	return &deploymentList, nil
}

func NewGeneratedDeploymentClient(cfg *rest.Config) (DeploymentClient, error) {
	kubeInterface, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	return &generatedDeploymentClient{client: kubeInterface}, nil
}

type generatedDeploymentClient struct {
	client kubernetes.Interface
}

func (g *generatedDeploymentClient) Get(ctx context.Context, objectKey client.ObjectKey) (*appsv1.Deployment, error) {
	return g.client.AppsV1().Deployments(objectKey.Namespace).Get(objectKey.Name, v1.GetOptions{})
}

func (g *generatedDeploymentClient) List(ctx context.Context, options ...client.ListOption) (*appsv1.DeploymentList, error) {
	listOptions := &client.ListOptions{}
	for _, v := range options {
		v.ApplyToList(listOptions)
	}
	raw := v1.ListOptions{}
	if converted := listOptions.AsListOptions(); converted != nil {
		raw = *converted
	}
	return g.client.AppsV1().Deployments(listOptions.Namespace).List(raw)
}
