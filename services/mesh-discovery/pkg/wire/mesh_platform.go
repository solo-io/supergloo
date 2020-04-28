package wire

import (
	"github.com/google/wire"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/aws_creds"
	mesh_platform "github.com/solo-io/service-mesh-hub/services/common/mesh-platform"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/mesh-platform/k8s"
	"github.com/solo-io/service-mesh-hub/services/common/mesh-platform/rest"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh/rest/aws"
	appmesh_client "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh/rest/aws/clients/appmesh"
)

var AwsSet = wire.NewSet(
	rest.NewAwsAPIHandler,
	aws_creds.DefaultSecretAwsCredsConverter,
	appmesh_client.NewAppMeshClientFactory,
	aws.NewAppMeshDiscoveryReconcilerFactory,
)

func MeshPlatformCredentialsHandlersProvider(
	asyncManagerController *mc_manager.AsyncManagerController,
	awsCredsHandler rest.AwsCredsHandler,
) []mesh_platform.MeshPlatformCredentialsHandler {
	return []mesh_platform.MeshPlatformCredentialsHandler{
		asyncManagerController,
		awsCredsHandler,
	}
}
