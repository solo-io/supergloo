package kubernetes_core

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type PodClientFactory func(client client.Client) PodClient

func NewPodClientFactory() PodClientFactory {
	return NewPodClient
}

func NewPodClientForConfig(cfg *rest.Config) (PodClient, error) {
	dynamicClient, err := client.New(cfg, client.Options{})
	if err != nil {
		return nil, err
	}
	return &podClient{client: dynamicClient}, nil
}

func NewPodClient(client client.Client) PodClient {
	return &podClient{
		client: client,
	}
}

type podClient struct {
	client client.Client
}

func (s *podClient) Get(ctx context.Context, name string, namespace string) (*corev1.Pod, error) {
	pod := corev1.Pod{}
	err := s.client.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, &pod)
	if err != nil {
		return nil, err
	}
	return &pod, nil
}

func (s *podClient) List(ctx context.Context, options ...client.ListOption) (*corev1.PodList, error) {
	podList := corev1.PodList{}
	err := s.client.List(ctx, &podList, options...)
	if err != nil {
		return &podList, err
	}
	return &podList, nil
}
