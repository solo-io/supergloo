package server

import (
	"strings"

	"github.com/solo-io/mesh-projects/pkg/env"

	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/mesh-projects/pkg/common/docker"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

//go:generate mockgen -source ./server_version.go -destination mocks/server_version.go

type ServerVersion struct {
	Namespace  string
	Containers []*ImageMeta
}

type ImageMeta struct {
	Tag      string
	Name     string
	Registry string
}

type DeploymentClient interface {
	GetDeployments(namespace string, labelSelector string) (*v1.DeploymentList, error)
}

type defaultDeploymentClient struct{}

func NewDeploymentClient() DeploymentClient {
	return &defaultDeploymentClient{}
}

func (k *defaultDeploymentClient) GetDeployments(namespace string, labelSelector string) (*v1.DeploymentList, error) {
	cfg, err := kubeutils.GetConfig("", "")
	if err != nil {
		// kubecfg is missing, therefore no cluster is present, only print client version
		return nil, nil
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
func DefaultServerVersionClientProvider(namespace string) ServerVersionClient {
	return NewServerVersionClient(namespace, NewDeploymentClient())
}

type serverVersionClient struct {
	namespace    string
	configClient DeploymentClient
}

func NewServerVersionClient(namespace string, kubeConfigClient DeploymentClient) ServerVersionClient {
	return &serverVersionClient{namespace: namespace, configClient: kubeConfigClient}
}

func (k *serverVersionClient) GetServerVersion() (*ServerVersion, error) {
	deployments, err := k.configClient.GetDeployments(k.namespace, "app="+env.DefaultWriteNamespace)
	if err != nil {
		return nil, ConfigClientError(err)
	}
	if deployments == nil { // service-mesh-hub deployments available
		return nil, nil
	}
	imageParser := docker.NewImageNameParser()
	var containers []*ImageMeta
	for _, v := range deployments.Items {
		for _, container := range v.Spec.Template.Spec.Containers {
			parsedImage, err := imageParser.Parse(container.Image)
			if err != nil {
				return nil, err
			}
			splitRepoName := strings.Split(parsedImage.Path, "/")
			containers = append(containers, &ImageMeta{
				Tag:      parsedImage.Tag,
				Name:     container.Image,
				Registry: parsedImage.Domain + "/" + strings.Join(splitRepoName[:len(splitRepoName)-1], "/"),
			})
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
