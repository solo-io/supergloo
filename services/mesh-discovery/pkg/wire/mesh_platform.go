package wire

import (
	"github.com/google/wire"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/aws_creds"
	mesh_platform "github.com/solo-io/service-mesh-hub/services/common/mesh-platform"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/mesh-platform/k8s"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh-workload/k8s/appmesh"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh/rest/aws"
	appmesh_client "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh/rest/aws/clients/appmesh"
	aws3 "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/mesh-platform/aws"
	aws_utils "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/mesh-platform/aws/parser"
)

var AwsSet = wire.NewSet(
	aws3.NewAwsAPIHandler,
	aws_creds.DefaultSecretAwsCredsConverter,
	appmesh_client.NewAppMeshClientFactory,
	aws.NewAppMeshDiscoveryReconcilerFactory,
	aws_utils.NewArnParser,
	aws_utils.NewAppMeshParser,
	appmesh.AppMeshWorkloadScannerFactoryProvider,
)

func MeshPlatformCredentialsHandlersProvider(
	asyncManagerController *mc_manager.AsyncManagerController,
	awsCredsHandler aws3.AwsCredsHandler,
) []mesh_platform.MeshPlatformCredentialsHandler {
	return []mesh_platform.MeshPlatformCredentialsHandler{
		asyncManagerController,
		awsCredsHandler,
	}
}
