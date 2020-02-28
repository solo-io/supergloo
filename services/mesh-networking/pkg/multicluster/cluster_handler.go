package networking_multicluster

import (
	"context"

	"github.com/solo-io/mesh-projects/services/common"
	mc_manager "github.com/solo-io/mesh-projects/services/common/multicluster/manager"
)

// this is the main entrypoint for all mesh-group multi cluster logic
func NewMeshNetworkingClusterHandler(
	ctx context.Context,
	localManager mc_manager.AsyncManager,
	csrControllerFactory MeshGroupCertificateSigningRequestControllerFactory,
) (mc_manager.AsyncManagerHandler, error) {

	handler := &meshNetworkingClusterHandler{
		localManager:         localManager,
		csrControllerFactory: csrControllerFactory,
	}

	// be sure that we are also watching our local cluster
	err := handler.ClusterAdded(ctx, localManager, common.LocalClusterName)
	if err != nil {
		return nil, err
	}

	return handler, nil
}

type meshNetworkingClusterHandler struct {
	localManager         mc_manager.AsyncManager
	csrControllerFactory MeshGroupCertificateSigningRequestControllerFactory
}

func (m *meshNetworkingClusterHandler) ClusterAdded(ctx context.Context, mgr mc_manager.AsyncManager, clusterName string) error {
	return nil
}

func (m *meshNetworkingClusterHandler) ClusterRemoved(clusterName string) error {
	return nil
}
