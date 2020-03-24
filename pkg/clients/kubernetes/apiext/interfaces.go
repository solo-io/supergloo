package kubernetes_apiext

import (
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/rest"
)

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_interfaces.go

type ServerVersionClient interface {
	Get() (*version.Info, error)
}

type ApiExtensionsInterfaceFactory func(c *rest.Config) (*clientset.Clientset, error)

func ApiExtensionsInterfaceFactoryProvider() ApiExtensionsInterfaceFactory {
	return clientset.NewForConfig
}

type CustomResourceDefinitionClient interface {
	Get(name string) (*v1beta1.CustomResourceDefinition, error)
	List() (*v1beta1.CustomResourceDefinitionList, error)
	Delete(name string) error
}
