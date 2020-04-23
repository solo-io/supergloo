package wire

import (
	"github.com/google/wire"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/aws_creds"
	rest_api "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/rest-api"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/rest-api/aws"
	appmesh_client "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/rest-api/aws/appmesh-client"
)

var AwsSet = wire.NewSet(
	aws.NewAwsCredsHandler,
	aws_creds.DefaultSecretAwsCredsConverter,
	appmesh_client.NewAppMeshClientFactory,
	aws.NewAppMeshDiscoveryReconcilerFactory,
)

func RestAPIHandlersProvider(
	awsHandler aws.AwsCredsHandler,
) []rest_api.RestAPICredsHandler {
	return []rest_api.RestAPICredsHandler{
		awsHandler,
	}
}
