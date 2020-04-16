package kubernetes_core

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewServiceAccountClientForConfig(cfg *rest.Config) (ServiceAccountClient, error) {
	client, err := client.New(cfg, client.Options{})
	if err != nil {
		return nil, err
	}
	return &serviceAccountClient{client: client}, nil
}

type serviceAccountClient struct {
	client client.Client
}

func (s *serviceAccountClient) Create(ctx context.Context, serviceAccount *corev1.ServiceAccount) error {
	return s.client.Create(ctx, serviceAccount)
}

func (s *serviceAccountClient) Get(ctx context.Context, name, namespace string) (*corev1.ServiceAccount, error) {
	sa := corev1.ServiceAccount{}
	err := s.client.Get(ctx, client.ObjectKey{
		Namespace: namespace,
		Name:      name,
	}, &sa)
	return &sa, err
}

func (s *serviceAccountClient) Update(ctx context.Context, serviceAccount *corev1.ServiceAccount) error {
	return s.client.Update(ctx, serviceAccount)
}

func (s *serviceAccountClient) List(ctx context.Context, options ...client.ListOption) (*corev1.ServiceAccountList, error) {
	saList := corev1.ServiceAccountList{}
	err := s.client.List(ctx, &saList, options...)
	return &saList, err
}
