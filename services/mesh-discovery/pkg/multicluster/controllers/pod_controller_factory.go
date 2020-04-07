package controllers

import (
	core_controllers "github.com/solo-io/service-mesh-hub/services/common/cluster/core/v1/controller"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/multicluster/manager"
)

func NewPodControllerFactory() PodControllerFactory {
	return &defaultPodControllerFactory{}
}

type defaultPodControllerFactory struct{}

func (d *defaultPodControllerFactory) Build(mgr mc_manager.AsyncManager, clusterName string) (core_controllers.PodController, error) {
	return core_controllers.NewPodController(clusterName, mgr.Manager())
}
