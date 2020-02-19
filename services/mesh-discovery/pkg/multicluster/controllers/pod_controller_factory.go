package controllers

import (
	core_controllers "github.com/solo-io/mesh-projects/services/common/cluster/core/v1/controller"
	mc_manager "github.com/solo-io/mesh-projects/services/common/multicluster/manager"
)

func NewPodControllerFactory() PodControllerFactory {
	return &defaultPodControllerFactory{}
}

type defaultPodControllerFactory struct{}

func (d *defaultPodControllerFactory) Build(mgr mc_manager.AsyncManager, clusterName string) (core_controllers.PodController, error) {
	return core_controllers.NewPodController(clusterName, mgr.Manager())
}
