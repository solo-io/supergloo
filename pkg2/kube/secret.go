package kube

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// If you change this interface, you have to rerun mockgen
type SecretClient interface {
	Delete(namespace string, name string) error
}

type KubeSecretClient struct {
	kube kubernetes.Interface
}

func NewKubeSecretClient(kube kubernetes.Interface) *KubeSecretClient {
	return &KubeSecretClient{
		kube: kube,
	}
}

func (client *KubeSecretClient) Delete(namespace string, name string) error {
	return client.kube.CoreV1().Secrets(namespace).Delete(name, &v1.DeleteOptions{})
}
