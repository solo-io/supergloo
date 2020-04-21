package wire

import (
	"github.com/google/wire"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/aws_creds"
	"github.com/solo-io/service-mesh-hub/services/common/multicluster/manager/rest_watcher/aws"
	rest_api "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/rest-api"
)

var AwsSet = wire.NewSet(
	aws.NewAwsCredsHandler,
	aws_creds.DefaultSecretAwsCredsConverter,
	rest_api.NewAppMeshAPIReconcilerFactory,
)
