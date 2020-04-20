package controllers

import (
	discovery_controllers "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/controller"
	"github.com/solo-io/service-mesh-hub/services/common/multicluster/manager/k8s_manager"
)

func NewMeshControllerFactory() MeshEventWatcherFactory {
	return &meshControllerFactory{}
}

type meshControllerFactory struct{}

func (d *meshControllerFactory) Build(mgr k8s_manager.AsyncManager, clusterName string) discovery_controllers.MeshEventWatcher {
	return discovery_controllers.NewMeshEventWatcher(clusterName, mgr.Manager())
}
