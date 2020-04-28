package wire

import (
	"context"

	"github.com/google/wire"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/aws_creds"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/multicluster/manager"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/rest/aws"
	v1 "k8s.io/api/core/v1"
)

func MeshPlatformCredentialsHandlersProvider(
	asyncManagerController *mc_manager.AsyncManagerController,
	awsCredsHandler aws.AwsCredsHandler,
) []mc_manager.MeshPlatformCredentialsHandler {
	return []mc_manager.MeshPlatformCredentialsHandler{
		asyncManagerController,
		awsCredsHandler,
	}
}

var AwsSet = wire.NewSet(
	NewNetworkingAwsCredsHandler,
	aws_creds.DefaultSecretAwsCredsConverter,
)

// Temporary stub until AppMesh translation is implemented
type networkingAwsCredsHandler struct {
}

func NewNetworkingAwsCredsHandler() aws.AwsCredsHandler {
	return &networkingAwsCredsHandler{}
}

func (n *networkingAwsCredsHandler) MeshPlatformAdded(ctx context.Context, secret *v1.Secret) error {
	return nil
}

func (n *networkingAwsCredsHandler) MeshPlatformRemoved(ctx context.Context, secret *v1.Secret) error {
	return nil
}
