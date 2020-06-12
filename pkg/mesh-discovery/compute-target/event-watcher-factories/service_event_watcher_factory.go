package event_watcher_factories

import (
	core_controllers "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/controller"
	mc_manager "github.com/solo-io/service-mesh-hub/pkg/common/compute-target/k8s"
)

func NewServiceEventWatcherFactory() ServiceEventWatcherFactory {
	return &serviceEventWatcherFactory{}
}

type serviceEventWatcherFactory struct{}

func (d *serviceEventWatcherFactory) Build(mgr mc_manager.AsyncManager, clusterName string) core_controllers.ServiceEventWatcher {
	return core_controllers.NewServiceEventWatcher(clusterName, mgr.Manager())
}
