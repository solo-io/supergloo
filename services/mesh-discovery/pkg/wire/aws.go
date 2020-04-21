package wire

import (
	"github.com/google/wire"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/aws_creds"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/rest-api/aws"
)

var AwsSet = wire.NewSet(
	aws.NewAwsCredsHandler,
	aws_creds.DefaultSecretAwsCredsConverter,
	aws.NewAppMeshReconcilerFactory,
)
