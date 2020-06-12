package wire

import (
	"github.com/google/wire"
	"github.com/solo-io/service-mesh-hub/pkg/common/aws/aws_creds"
	"github.com/solo-io/service-mesh-hub/pkg/common/aws/clients"
	compute_target "github.com/solo-io/service-mesh-hub/pkg/common/compute-target"
	mc_manager "github.com/solo-io/service-mesh-hub/pkg/common/compute-target/k8s"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/compute-target/aws"
)

func ComputeTargetCredentialsHandlersProvider(
	asyncManagerController *mc_manager.AsyncManagerController,
	awsCredsHandler compute_target.ComputeTargetCredentialsHandler,
) []compute_target.ComputeTargetCredentialsHandler {
	return []compute_target.ComputeTargetCredentialsHandler{
		asyncManagerController,
		awsCredsHandler,
	}
}

var AwsSet = wire.NewSet(
	aws.NewNetworkingAwsCredsHandler,
	aws_creds.DefaultSecretAwsCredsConverter,
	clients.STSClientFactoryProvider,
)
