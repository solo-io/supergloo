package controllers

import (
	core_controllers "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1/controller"
	"github.com/solo-io/service-mesh-hub/services/common/multicluster/manager/k8s_manager"
)

func NewPodEventWatcherFactory() PodEventWatcherFactory {
	return &defaultPodEventWatcherFactory{}
}

type defaultPodEventWatcherFactory struct{}

func (d *defaultPodEventWatcherFactory) Build(mgr k8s_manager.AsyncManager, clusterName string) core_controllers.PodEventWatcher {
	return core_controllers.NewPodEventWatcher(clusterName, mgr.Manager())
}
