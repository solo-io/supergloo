package kubernetes_apiext

import (
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiext "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

type GeneratedCrdClientFactory func(cfg *rest.Config) (CustomResourceDefinitionClient, error)

func NewGeneratedCrdClientFactory() GeneratedCrdClientFactory {
	return GeneratedCrdClientFromRestConfig
}

func GeneratedCrdClientFromRestConfig(cfg *rest.Config) (CustomResourceDefinitionClient, error) {
	clientSet, err := apiext.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	return NewGeneratedCustomResourceDefinitionClient(clientSet), nil
}

func NewGeneratedCustomResourceDefinitionClient(client apiext.Interface) CustomResourceDefinitionClient {
	return &generatedCustomResourceDefinitionClient{
		client: client,
	}
}

type generatedCustomResourceDefinitionClient struct {
	client apiext.Interface
}

func (g *generatedCustomResourceDefinitionClient) Get(name string) (*v1beta1.CustomResourceDefinition, error) {
	return g.client.ApiextensionsV1beta1().CustomResourceDefinitions().Get(name, v1.GetOptions{})
}

func (g *generatedCustomResourceDefinitionClient) List() (*v1beta1.CustomResourceDefinitionList, error) {
	return g.client.ApiextensionsV1beta1().CustomResourceDefinitions().List(v1.ListOptions{})
}

func (g *generatedCustomResourceDefinitionClient) Delete(name string) error {
	return g.client.ApiextensionsV1beta1().CustomResourceDefinitions().Delete(name, nil)
}
