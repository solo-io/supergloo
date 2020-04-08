package controllers

import (
	core_controllers "github.com/solo-io/service-mesh-hub/services/common/cluster/core/v1/controller"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/multicluster/manager"
)

func NewServiceControllerFactory() ServiceControllerFactory {
	return &serviceControllerFactory{}
}

type serviceControllerFactory struct{}

func (d *serviceControllerFactory) Build(mgr mc_manager.AsyncManager, clusterName string) (core_controllers.ServiceController, error) {
	return core_controllers.NewServiceController(clusterName, mgr.Manager())
}
