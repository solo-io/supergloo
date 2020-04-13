package controller_factories

import (
	networking_controller "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/controller"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/multicluster/manager"
)

func NewLocalVirtualMeshEventWatcher(mgr mc_manager.AsyncManager) networking_controller.VirtualMeshEventWatcher {
	return networking_controller.NewVirtualMeshEventWatcher("management-plane-virtual-mesh-event-watcher", mgr.Manager())
}
