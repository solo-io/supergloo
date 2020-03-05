package istio_networking

import (
	"context"

	"istio.io/client-go/pkg/apis/networking/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:generate mockgen -source ./interfaces.go -destination mocks/mock_clients.go

type VirtualServiceClient interface {
	Get(ctx context.Context, key client.ObjectKey) (*v1beta1.VirtualService, error)
	Create(ctx context.Context, virtualService *v1beta1.VirtualService, options ...client.CreateOption) error
	Update(ctx context.Context, virtualService *v1beta1.VirtualService, options ...client.UpdateOption) error
	// Create the VirtualService if it does not exist, otherwise update
	Upsert(ctx context.Context, virtualService *v1beta1.VirtualService) error
}
