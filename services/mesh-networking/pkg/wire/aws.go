package wire

import (
	"context"

	"github.com/google/wire"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/aws_creds"
	"github.com/solo-io/service-mesh-hub/services/common/multicluster/manager/rest_watcher/aws"
	v1 "k8s.io/api/core/v1"
)

var AwsSet = wire.NewSet(
	NewNetworkingAwsCredsHandler,
	aws_creds.DefaultSecretAwsCredsConverter,
)

type networkingAwsCredsHandler struct {
}

// Temporary stub until AppMesh translation is implemented
func NewNetworkingAwsCredsHandler() aws.AwsCredsHandler {
	return &networkingAwsCredsHandler{}
}

func (n networkingAwsCredsHandler) RestAPIAdded(ctx context.Context, secret *v1.Secret) error {
	// create map of apiName to secret or client
	return nil
}

func (n networkingAwsCredsHandler) RestAPIRemoved(ctx context.Context, apiName string) error {
	// remove from map of apiName to secret or client
	return nil
}
