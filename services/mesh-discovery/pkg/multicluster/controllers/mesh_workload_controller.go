package controllers

import (
	discovery_controllers "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/controller"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/multicluster/manager"
)

func NewMeshWorkloadControllerFactory() MeshWorkloadControllerFactory {
	return &meshWorkloadControllerFactory{}
}

type meshWorkloadControllerFactory struct{}

func (d *meshWorkloadControllerFactory) Build(mgr mc_manager.AsyncManager, clusterName string) (discovery_controllers.MeshWorkloadController, error) {
	return discovery_controllers.NewMeshWorkloadController(clusterName, mgr.Manager())
}
