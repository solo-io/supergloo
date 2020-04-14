package controllers

import (
	discovery_controllers "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/controller"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/multicluster/manager"
)

func NewMeshControllerFactory() MeshControllerFactory {
	return &meshControllerFactory{}
}

type meshControllerFactory struct{}

func (d *meshControllerFactory) Build(mgr mc_manager.AsyncManager, clusterName string) (discovery_controllers.MeshController, error) {
	return discovery_controllers.NewMeshController(clusterName, mgr.Manager())
}
