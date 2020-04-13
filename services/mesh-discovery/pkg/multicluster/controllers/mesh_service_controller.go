package controllers

import (
	discovery_controllers "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/controller"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/multicluster/manager"
)

func NewMeshServiceEventWatcherFactory() MeshServiceEventWatcherFactory {
	return &MeshServiceEventWatcherFactory{}
}

type MeshServiceEventWatcherFactory struct{}

func (d *MeshServiceEventWatcherFactory) Build(
	mgr mc_manager.AsyncManager,
	clusterName string,
) (discovery_controllers.MeshServiceEventWatcher, error) {
	return discovery_controllers.NewMeshServiceEventWatcher(clusterName, mgr.Manager())
}
