package multicluster

import (
	"context"

	"github.com/solo-io/mesh-projects/services/common"
	mc_manager "github.com/solo-io/mesh-projects/services/common/multicluster/manager"
)

// this is the main entrypoint for all mesh-group multi cluster logic
func NewMeshGroupClusterHandler(
	ctx context.Context,
	localManager mc_manager.AsyncManager,
) (mc_manager.AsyncManagerHandler, error) {

	handler := &meshDiscoveryClusterHandler{
		ctx:          ctx,
		localManager: localManager,
	}

	// be sure that we are also watching our local cluster
	err := handler.ClusterAdded(localManager, common.LocalClusterName)
	if err != nil {
		return nil, err
	}

	return handler, nil
}

type meshDiscoveryClusterHandler struct {
	localManager mc_manager.AsyncManager
	ctx          context.Context
}

func (m *meshDiscoveryClusterHandler) ClusterAdded(mgr mc_manager.AsyncManager, clusterName string) error {
	return nil
}

func (m *meshDiscoveryClusterHandler) ClusterRemoved(clusterName string) error {
	return nil
}
