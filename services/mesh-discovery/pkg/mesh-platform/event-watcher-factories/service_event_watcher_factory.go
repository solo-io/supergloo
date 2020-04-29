package event_watcher_factories

import (
	core_controllers "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1/controller"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/mesh-platform/k8s"
)

func NewServiceEventWatcherFactory() ServiceEventWatcherFactory {
	return &serviceEventWatcherFactory{}
}

type serviceEventWatcherFactory struct{}

func (d *serviceEventWatcherFactory) Build(mgr mc_manager.AsyncManager, clusterName string) core_controllers.ServiceEventWatcher {
	return core_controllers.NewServiceEventWatcher(clusterName, mgr.Manager())
}
