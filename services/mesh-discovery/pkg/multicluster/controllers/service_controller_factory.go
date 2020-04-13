package controllers

import (
	core_controllers "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1/controller"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/multicluster/manager"
)

func NewServiceControllerFactory() ServiceControllerFactory {
	return &serviceEventWatcherFactory{}
}

type serviceEventWatcherFactory struct{}

func (d *serviceEventWatcherFactory) Build(mgr mc_manager.AsyncManager, clusterName string) (core_controllers.ServiceEventWatcher, error) {
	return core_controllers.NewServiceController(clusterName, mgr.Manager())
}
