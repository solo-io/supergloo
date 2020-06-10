package strategies

import (
	"context"
	"fmt"

	"github.com/solo-io/go-utils/contextutils"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	"go.uber.org/zap"
)

// once the mesh services have had their federation metadata updated, call this function to write both that metadata and their new federation status to the cluster
func updateServices(ctx context.Context, federatedServices []*smh_discovery.MeshService, meshServiceClient smh_discovery.MeshServiceClient) error {
	logger := contextutils.LoggerFrom(ctx)

	for _, federatedService := range federatedServices {
		err := meshServiceClient.UpsertMeshServiceSpec(ctx, federatedService)
		if err != nil {
			logger.Errorw(fmt.Sprintf("Failed to set federation metadata on mesh service"),
				zap.Any("object_meta", federatedService.ObjectMeta),
				zap.Error(err),
			)
			return err
		}
	}

	return nil
}
