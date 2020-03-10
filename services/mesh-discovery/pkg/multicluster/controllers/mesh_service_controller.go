package controllers

import (
	discovery_controllers "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1/controller"
	mc_manager "github.com/solo-io/mesh-projects/services/common/multicluster/manager"
)

func NewMeshServiceControllerFactory() MeshServiceControllerFactory {
	return &meshServiceControllerFactory{}
}

type meshServiceControllerFactory struct{}

func (d *meshServiceControllerFactory) Build(
	mgr mc_manager.AsyncManager,
	clusterName string,
) (discovery_controllers.MeshServiceController, error) {
	return discovery_controllers.NewMeshServiceController(clusterName, mgr.Manager())
}
