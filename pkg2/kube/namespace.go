package kube

import (
	kubecore "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	kubemeta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// If you change this interface, you have to rerun mockgen
type NamespaceClient interface {
	CreateNamespaceIfNotExist(namespace string) error
	TryDeleteInstallNamespace(namespaceName string)
}

type KubeNamespaceClient struct {
	kube kubernetes.Interface
}

func NewKubeNamespaceClient(kube kubernetes.Interface) *KubeNamespaceClient {
	return &KubeNamespaceClient{
		kube: kube,
	}
}

func getNamespace(namespaceName string) *kubecore.Namespace {
	return &kubecore.Namespace{
		ObjectMeta: kubemeta.ObjectMeta{
			Name: namespaceName,
		},
	}
}

func (client *KubeNamespaceClient) CreateNamespaceIfNotExist(namespaceName string) error {
	_, err := client.kube.CoreV1().Namespaces().Create(getNamespace(namespaceName))
	if apierrors.IsAlreadyExists(err) {
		return nil
	}
	return err
}

func (client *KubeNamespaceClient) TryDeleteInstallNamespace(namespaceName string) {
	client.kube.CoreV1().Namespaces().Delete(namespaceName, &kubemeta.DeleteOptions{})
}
