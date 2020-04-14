package controllers

import (
	discovery_controllers "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/controller"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/multicluster/manager"
)

func NewMeshWorkloadEventWatcherFactory() MeshWorkloadEventWatcherFactory {
	return &meshWorkloadEventWatcherFactory{}
}

type meshWorkloadEventWatcherFactory struct{}

func (d *meshWorkloadEventWatcherFactory) Build(mgr mc_manager.AsyncManager, clusterName string) discovery_controllers.MeshWorkloadEventWatcher {
	return discovery_controllers.NewMeshWorkloadEventWatcher(clusterName, mgr.Manager())
}
