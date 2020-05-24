package multicluster

import (
	"context"

	"github.com/avast/retry-go"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_interfaces.go

// Simple map get interface to expose the map of dynamic clients to the local controller
type DynamicClientGetter interface {
	// Return (client, true) if found, otherwise (nil, false)
	GetClientForCluster(ctx context.Context, clusterName string, opts ...retry.Option) (client.Client, error)
}
