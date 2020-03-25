package istio_networking

import (
	"context"

	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ServiceEntryClientFactory func(client client.Client) ServiceEntryClient

func NewServiceEntryClientFactory() ServiceEntryClientFactory {
	return NewServiceEntryClient
}

func NewServiceEntryClient(client client.Client) ServiceEntryClient {
	return &serviceEntryClient{
		client: client,
	}
}

type serviceEntryClient struct {
	client client.Client
}

func (s *serviceEntryClient) Create(ctx context.Context, ServiceEntry *v1alpha3.ServiceEntry) error {
	return s.client.Create(ctx, ServiceEntry)
}

func (s *serviceEntryClient) Get(ctx context.Context, objKey client.ObjectKey) (*v1alpha3.ServiceEntry, error) {
	ServiceEntry := v1alpha3.ServiceEntry{}
	err := s.client.Get(ctx, objKey, &ServiceEntry)
	if err != nil {
		return nil, err
	}

	return &ServiceEntry, nil
}

func (g *serviceEntryClient) Update(ctx context.Context, serviceEntry *v1alpha3.ServiceEntry) error {
	return g.client.Update(ctx, serviceEntry)
}
