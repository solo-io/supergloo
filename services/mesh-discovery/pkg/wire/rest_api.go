package wire

import (
	"github.com/google/wire"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/aws_creds"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/rest"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/rest/aws"
	appmesh_client "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/rest/aws/clients/appmesh"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/rest/aws/discovery"
)

var AwsSet = wire.NewSet(
	aws.NewAwsAPIHandler,
	aws_creds.DefaultSecretAwsCredsConverter,
	appmesh_client.NewAppMeshClientFactory,
	discovery.NewAppMeshDiscoveryReconcilerFactory,
)

func RestAPIHandlersProvider(
	awsHandler aws.AwsCredsHandler,
) []rest.RestAPICredsHandler {
	return []rest.RestAPICredsHandler{
		awsHandler,
	}
}
