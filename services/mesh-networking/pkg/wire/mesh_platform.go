package wire

import (
	"github.com/google/wire"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/aws_creds"
	compute_target "github.com/solo-io/service-mesh-hub/services/common/compute-target"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/compute-target/k8s"
	aws2 "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/compute-target/aws"
	"github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/compute-target/aws"
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
	aws.NewNetworkingAwsCredsHandler,
	aws_creds.DefaultSecretAwsCredsConverter,
)
