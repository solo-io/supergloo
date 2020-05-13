package cluster_registration

import (
	"context"

	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"k8s.io/client-go/tools/clientcmd"
)

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_interfaces.go

type ClusterRegistrationClient interface {
	// The remoteContextName must be passed explicitly rather than inferred from the remoteConfig because
	// of an open k8s bug, https://github.com/kubernetes/kubernetes/pull/87622.
	Register(
		ctx context.Context,
		remoteConfig clientcmd.ClientConfig,
		remoteClusterName string,
		remoteWriteNamespace string,
		remoteContextName string,
		discoverySource string,
		registerOpts ClusterRegisterOpts,
	) error
}

// perform all the steps necessary to de-register this cluster from the SMH installation
type ClusterDeregistrationClient interface {
	Deregister(ctx context.Context, kubeCluster *zephyr_discovery.KubernetesCluster) error
}
