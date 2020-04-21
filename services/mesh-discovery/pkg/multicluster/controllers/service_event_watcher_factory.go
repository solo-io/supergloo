package controllers

import (
	core_controllers "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1/controller"
	"github.com/solo-io/service-mesh-hub/services/common/multicluster/manager"
)

func NewServiceEventWatcherFactory() ServiceEventWatcherFactory {
	return &serviceEventWatcherFactory{}
}

type serviceEventWatcherFactory struct{}

func (d *serviceEventWatcherFactory) Build(mgr manager.AsyncManager, clusterName string) core_controllers.ServiceEventWatcher {
	return core_controllers.NewServiceEventWatcher(clusterName, mgr.Manager())
}
