package v1alpha2

import (
	"context"

	"github.com/linkerd/linkerd2/controller/gen/apis/serviceprofile/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:generate mockgen -source ./interfaces.go -destination ./mock/mock_interfaces.go

type ServiceProfileClient interface {
	Get(ctx context.Context, key client.ObjectKey) (*v1alpha2.ServiceProfile, error)
	List(ctx context.Context, options ...client.ListOption) (*v1alpha2.ServiceProfileList, error)
	Create(ctx context.Context, serviceProfile *v1alpha2.ServiceProfile, options ...client.CreateOption) error
	Update(ctx context.Context, serviceProfile *v1alpha2.ServiceProfile, options ...client.UpdateOption) error
	// Create the ServiceProfile if it does not exist, otherwise update
	UpsertSpec(ctx context.Context, serviceProfile *v1alpha2.ServiceProfile) error
	Delete(ctx context.Context, key client.ObjectKey) error
}
