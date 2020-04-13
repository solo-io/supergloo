package controllers

import (
	core_controllers "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1/controller"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/multicluster/manager"
)

func NewPodControllerFactory() PodEventWatcherFactory {
	return &defaultPodControllerFactory{}
}

type defaultPodControllerFactory struct{}

func (d *defaultPodControllerFactory) Build(mgr mc_manager.AsyncManager, clusterName string) (core_controllers.PodEventWatcher, error) {
	return core_controllers.NewPodController(clusterName, mgr.Manager())
}
