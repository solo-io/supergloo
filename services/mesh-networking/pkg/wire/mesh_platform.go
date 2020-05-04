package wire

import (
	"context"

	"github.com/google/wire"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/aws_creds"
	compute_target "github.com/solo-io/service-mesh-hub/services/common/compute-target"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/compute-target/k8s"
	aws2 "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/compute-target/aws"
	v1 "k8s.io/api/core/v1"
)

func ComputeTargetCredentialsHandlersProvider(
	asyncManagerController *mc_manager.AsyncManagerController,
	awsCredsHandler aws2.AwsCredsHandler,
) []compute_target.ComputeTargetCredentialsHandler {
	return []compute_target.ComputeTargetCredentialsHandler{
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

func NewNetworkingAwsCredsHandler() aws2.AwsCredsHandler {
	return &networkingAwsCredsHandler{}
}

func (n *networkingAwsCredsHandler) ComputeTargetAdded(ctx context.Context, secret *v1.Secret) error {
	return nil
}

func (n *networkingAwsCredsHandler) ComputeTargetRemoved(ctx context.Context, secret *v1.Secret) error {
	return nil
}
