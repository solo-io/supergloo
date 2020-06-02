package server

import (
	"github.com/rotisserie/eris"
	"github.com/solo-io/service-mesh-hub/cli/pkg/options"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/container-runtime"
	"github.com/solo-io/service-mesh-hub/pkg/container-runtime/docker"
	"github.com/solo-io/service-mesh-hub/pkg/kube/kubeconfig"
	k8s_apps "k8s.io/api/apps/v1"
	k8s_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	GetDeployments(namespace string, labelSelector string) (*k8s_apps.DeploymentList, error)
}

type defaultDeploymentClient struct {
	opts   *options.Options
	loader kubeconfig.KubeLoader
}

func NewDeploymentClient(loader kubeconfig.KubeLoader, opts *options.Options) DeploymentClient {
	return &defaultDeploymentClient{
		opts:   opts,
		loader: loader,
	}
}

func (k *defaultDeploymentClient) GetDeployments(namespace string, labelSelector string) (*k8s_apps.DeploymentList, error) {
	cfg, err := k.loader.GetRestConfigForContext(k.opts.Root.KubeConfig, k.opts.Root.KubeContext)
	if err != nil {
		return nil, err
	}
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	deployments, err := client.AppsV1().Deployments(namespace).List(k8s_meta.ListOptions{
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
func DefaultServerVersionClientProvider(opts *options.Options, loader kubeconfig.KubeLoader, imageNameParser docker.ImageNameParser) ServerVersionClient {
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
	deployments, err := k.configClient.GetDeployments(k.namespace, "app="+container_runtime.GetWriteNamespace())
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
