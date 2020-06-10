package event_watcher_factories

import (
	smh_discovery_controller "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/controller"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/compute-target/k8s"
)

func NewMeshWorkloadEventWatcherFactory() MeshWorkloadEventWatcherFactory {
	return &meshWorkloadEventWatcherFactory{}
}

type meshWorkloadEventWatcherFactory struct{}

func (d *meshWorkloadEventWatcherFactory) Build(mgr mc_manager.AsyncManager, clusterName string) smh_discovery_controller.MeshWorkloadEventWatcher {
	return smh_discovery_controller.NewMeshWorkloadEventWatcher(clusterName, mgr.Manager())
}
