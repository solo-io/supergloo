package deregister

import (
	"context"

	discovery_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
)

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_interfaces.go

// perform all the steps necessary to de-register this cluster from the SMH installation
type ClusterDeregistrationClient interface {
	Run(ctx context.Context, kubeCluster *discovery_v1alpha1.KubernetesCluster) error
}
