package v1alpha3

import (
	"context"

	"github.com/servicemeshinterface/smi-sdk-go/pkg/apis/split/v1alpha3"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:generate mockgen -source ./interfaces.go -destination ./mock/mock_interfaces.go

type TrafficSplitClient interface {
	Get(ctx context.Context, key client.ObjectKey) (*v1alpha3.TrafficSplit, error)
	Create(ctx context.Context, trafficSplitClient *v1alpha3.TrafficSplit, options ...client.CreateOption) error
	Update(ctx context.Context, trafficSplitClient *v1alpha3.TrafficSplit, options ...client.UpdateOption) error
	// Create the TrafficSplit if it does not exist, otherwise update
	UpsertSpec(ctx context.Context, trafficSplitClient *v1alpha3.TrafficSplit) error
	Delete(ctx context.Context, key client.ObjectKey) error
}
