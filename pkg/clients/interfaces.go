package clients

import (
	"context"

	"k8s.io/client-go/tools/clientcmd"
)

type ClusterRegistrationClient interface {
	Register(
		ctx context.Context,
		remoteConfig clientcmd.ClientConfig,
		remoteClusterName string,
		remoteWriteNamespace string,
		overwrite bool,
		useDevCsrAgentChart bool,
		localClusterDomainOverride string,
	) error
}
