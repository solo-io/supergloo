package kubernetes_discovery

import (
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/kubernetes"
)

func NewGeneratedServerVersionClient(client kubernetes.Interface) ServerVersionClient {
	return &generatedServerVersionClient{
		client: client,
	}
}

type generatedServerVersionClient struct {
	client kubernetes.Interface
}

func (g *generatedServerVersionClient) Get() (*version.Info, error) {
	return g.client.Discovery().ServerVersion()
}
