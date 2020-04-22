package deregister

import (
	"context"

	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
)

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_interfaces.go

// perform all the steps necessary to de-register this cluster from the SMH installation
type ClusterDeregistrationClient interface {
	// results in deleting the kubeCluster that's passed in
	Run(ctx context.Context, kubeCluster *zephyr_discovery.KubernetesCluster) error
}
