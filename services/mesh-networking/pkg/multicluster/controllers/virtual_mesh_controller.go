package controller_factories

import (
	zephyr_networking_controller "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/controller"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/mesh-platform/k8s"
)

func NewLocalVirtualMeshEventWatcher(mgr mc_manager.AsyncManager) zephyr_networking_controller.VirtualMeshEventWatcher {
	return zephyr_networking_controller.NewVirtualMeshEventWatcher("management-plane-virtual-mesh-event-watcher", mgr.Manager())
}
