package event_watcher_factories

import (
	zephyr_discovery_controller "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/controller"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/mesh-platform/k8s"
)

func NewMeshWorkloadEventWatcherFactory() MeshWorkloadEventWatcherFactory {
	return &meshWorkloadEventWatcherFactory{}
}

type meshWorkloadEventWatcherFactory struct{}

func (d *meshWorkloadEventWatcherFactory) Build(mgr mc_manager.AsyncManager, clusterName string) zephyr_discovery_controller.MeshWorkloadEventWatcher {
	return zephyr_discovery_controller.NewMeshWorkloadEventWatcher(clusterName, mgr.Manager())
}
