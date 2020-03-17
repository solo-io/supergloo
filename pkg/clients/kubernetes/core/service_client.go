package kubernetes_core

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ServiceClientFactory func(client client.Client) ServiceClient

func ServiceClientFactoryProvider() ServiceClientFactory {
	return NewServiceClient
}

func NewServiceClient(client client.Client) ServiceClient {
	return &serviceClient{
		client: client,
	}
}

type serviceClient struct {
	client client.Client
}

func (s *serviceClient) Get(ctx context.Context, name string, namespace string) (*corev1.Service, error) {
	service := corev1.Service{}
	err := s.client.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, &service)
	if err != nil {
		return nil, err
	}
	return &service, nil
}

func (s *serviceClient) List(ctx context.Context, options ...client.ListOption) (*corev1.ServiceList, error) {
	serviceList := corev1.ServiceList{}
	err := s.client.List(ctx, &serviceList, options...)
	if err != nil {
		return &serviceList, err
	}
	return &serviceList, nil
}
