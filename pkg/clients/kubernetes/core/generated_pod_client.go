package kubernetes_core

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type GeneratedPodClientFactory func(client kubernetes.Interface) PodClient

func GeneratedPodClientFactoryProvider() GeneratedPodClientFactory {
	return NewGeneratedPodClient
}

func NewGeneratedPodClient(client kubernetes.Interface) PodClient {
	return &generatedPodClient{
		client: client,
	}
}

type generatedPodClient struct {
	client kubernetes.Interface
}

func (s *generatedPodClient) Get(ctx context.Context, name string, namespace string) (*corev1.Pod, error) {
	return s.client.CoreV1().Pods(namespace).Get(name, v1.GetOptions{})
}

func (s *generatedPodClient) List(ctx context.Context, options ...client.ListOption) (*corev1.PodList, error) {
	listOptions := &client.ListOptions{}
	for _, v := range options {
		v.ApplyToList(listOptions)
	}
	raw := v1.ListOptions{}
	if converted := listOptions.AsListOptions(); converted != nil {
		raw = *converted
	}
	return s.client.CoreV1().Pods(listOptions.Namespace).List(raw)
}
