package v1alpha1

import (
	"context"

	"github.com/servicemeshinterface/smi-sdk-go/pkg/apis/split/v1alpha1"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:generate mockgen -source ./interfaces.go -destination ./mock/mock_interfaces.go

type TrafficSplitClient interface {
	Get(ctx context.Context, key client.ObjectKey) (*v1alpha1.TrafficSplit, error)
	Create(ctx context.Context, trafficSplitClient *v1alpha1.TrafficSplit, options ...client.CreateOption) error
	List(ctx context.Context, options ...client.ListOption) (*v1alpha1.TrafficSplitList, error)
	Update(ctx context.Context, trafficSplitClient *v1alpha1.TrafficSplit, options ...client.UpdateOption) error
	// Create the TrafficSplit if it does not exist, otherwise update
	UpsertSpec(ctx context.Context, trafficSplitClient *v1alpha1.TrafficSplit) error
	Delete(ctx context.Context, key client.ObjectKey) error
}
