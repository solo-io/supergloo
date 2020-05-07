package clients

import (
	"context"

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
		overwrite bool,
		useDevCsrAgentChart bool,
		localClusterDomainOverride string,
		remoteContextName string,
		clusterLabels map[string]string,
	) error
}
