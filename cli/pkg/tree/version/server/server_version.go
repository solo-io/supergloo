package server

import (
	"github.com/rotisserie/eris"
	common_config "github.com/solo-io/service-mesh-hub/cli/pkg/common/config"
	"github.com/solo-io/service-mesh-hub/cli/pkg/options"
	"github.com/solo-io/service-mesh-hub/pkg/common/docker"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

//go:generate mockgen -source ./server_version.go -destination mocks/server_version.go

type ServerVersion struct {
	Namespace  string
	Containers []*docker.Image
}

type ImageMeta struct {
	Tag      string
	Name     string
	Registry string
}

type DeploymentClient interface {
	GetDeployments(namespace string, labelSelector string) (*v1.DeploymentList, error)
}

type defaultDeploymentClient struct {
	opts   *options.Options
	loader common_config.KubeLoader
}

func NewDeploymentClient(loader common_config.KubeLoader, opts *options.Options) DeploymentClient {
	return &defaultDeploymentClient{
		opts:   opts,
		loader: loader,
	}
}

func (k *defaultDeploymentClient) GetDeployments(namespace string, labelSelector string) (*v1.DeploymentList, error) {
	cfg, err := k.loader.GetRestConfigForContext(k.opts.Root.KubeConfig, k.opts.Root.KubeContext)
	if err != nil {
		return nil, err
	}
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	deployments, err := client.AppsV1().Deployments(namespace).List(metav1.ListOptions{
		// search only for smh deployments based on labels
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, err
	}
	return deployments, nil
}

var (
	ConfigClientError = func(err error) error {
		return eris.Wrap(err, "error while getting kube config")
	}
)

type ServerVersionClient interface {
	GetServerVersion() (*ServerVersion, error)
}

// default ServerVersionClient
func DefaultServerVersionClientProvider(opts *options.Options, loader common_config.KubeLoader, imageNameParser docker.ImageNameParser) ServerVersionClient {
	return NewServerVersionClient(
		opts.Root.WriteNamespace,
		NewDeploymentClient(loader, opts),
		imageNameParser,
	)
}

type serverVersionClient struct {
	namespace       string
	configClient    DeploymentClient
	imageNameParser docker.ImageNameParser
}

func NewServerVersionClient(namespace string, configClient DeploymentClient, imageNameParser docker.ImageNameParser) *serverVersionClient {
	return &serverVersionClient{namespace: namespace, configClient: configClient, imageNameParser: imageNameParser}
}

func (k *serverVersionClient) GetServerVersion() (*ServerVersion, error) {
	deployments, err := k.configClient.GetDeployments(k.namespace, "app="+env.DefaultWriteNamespace)
	if err != nil {
		return nil, ConfigClientError(err)
	}
	if deployments == nil { // service-mesh-hub deployments available
		return nil, nil
	}

	var containers []*docker.Image
	for _, v := range deployments.Items {
		for _, container := range v.Spec.Template.Spec.Containers {
			parsedImage, err := k.imageNameParser.Parse(container.Image)
			if err != nil {
				return nil, err
			}
			containers = append(containers, parsedImage)
		}
	}
	if len(containers) == 0 {
		return nil, nil
	}
	serverVersion := &ServerVersion{
		Namespace:  k.namespace,
		Containers: containers,
	}

	return serverVersion, nil
}
