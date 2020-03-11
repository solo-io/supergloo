package istio_networking

import (
	"context"

	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:generate mockgen -source ./interfaces.go -destination ./mock/mock_interfaces.go

type GatewayClient interface {
	Create(ctx context.Context, gateway *v1alpha3.Gateway) error
	Get(ctx context.Context, objKey client.ObjectKey) (*v1alpha3.Gateway, error)
	Update(ctx context.Context, gateway *v1alpha3.Gateway) error
}

type EnvoyFilterClient interface {
	Create(ctx context.Context, envoyFilter *v1alpha3.EnvoyFilter) error
	Get(ctx context.Context, objKey client.ObjectKey) (*v1alpha3.EnvoyFilter, error)
}

type ServiceEntryClient interface {
	Create(ctx context.Context, serviceEntry *v1alpha3.ServiceEntry) error
	Get(ctx context.Context, objKey client.ObjectKey) (*v1alpha3.ServiceEntry, error)
}

type VirtualServiceClient interface {
	Get(ctx context.Context, key client.ObjectKey) (*v1alpha3.VirtualService, error)
	Create(ctx context.Context, virtualService *v1alpha3.VirtualService, options ...client.CreateOption) error
	Update(ctx context.Context, virtualService *v1alpha3.VirtualService, options ...client.UpdateOption) error
	// Create the VirtualService if it does not exist, otherwise update
	Upsert(ctx context.Context, virtualService *v1alpha3.VirtualService) error
}

type DestinationRuleClient interface {
	Get(ctx context.Context, key client.ObjectKey) (*v1alpha3.DestinationRule, error)
	Create(ctx context.Context, destinationRule *v1alpha3.DestinationRule) error
	Update(ctx context.Context, destinationRule *v1alpha3.DestinationRule, options ...client.UpdateOption) error
	Upsert(ctx context.Context, destinationRule *v1alpha3.DestinationRule) error
}
