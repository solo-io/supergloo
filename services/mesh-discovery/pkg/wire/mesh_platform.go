package wire

import (
	"github.com/google/wire"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/aws_creds"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/multicluster/manager"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/rest/aws"
	appmesh_client "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/rest/aws/clients/appmesh"
)

var AwsSet = wire.NewSet(
	aws.NewAwsAPIHandler,
	aws_creds.DefaultSecretAwsCredsConverter,
	appmesh_client.NewAppMeshClientFactory,
	aws.NewAppMeshDiscoveryReconcilerFactory,
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
