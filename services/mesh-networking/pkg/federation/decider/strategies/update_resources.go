package strategies

import (
	"context"
	"fmt"

	"github.com/solo-io/go-utils/contextutils"
	discovery_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	discovery_core "github.com/solo-io/service-mesh-hub/pkg/clients/zephyr/discovery"
	"go.uber.org/zap"
)

// once the mesh services have had their federation metadata updated, call this function to write both that metadata and their new federation status to the cluster
func updateServices(ctx context.Context, federatedServices []*discovery_v1alpha1.MeshService, meshServiceClient discovery_core.MeshServiceClient) error {
	logger := contextutils.LoggerFrom(ctx)

	for _, federatedService := range federatedServices {
		err := meshServiceClient.Update(ctx, federatedService)
		if err != nil {
			logger.Errorw(fmt.Sprintf("Failed to set federation metadata on mesh service"),
				zap.Any("opbject_meta", federatedService.ObjectMeta),
				zap.Error(err),
			)
			return err
		}
	}

	return nil
}
