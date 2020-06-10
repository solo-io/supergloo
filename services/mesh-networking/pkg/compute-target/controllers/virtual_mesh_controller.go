package controller_factories

import (
	smh_networking_controller "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/controller"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/compute-target/k8s"
)

func NewLocalVirtualMeshEventWatcher(mgr mc_manager.AsyncManager) smh_networking_controller.VirtualMeshEventWatcher {
	return smh_networking_controller.NewVirtualMeshEventWatcher("management-plane-virtual-mesh-event-watcher", mgr.Manager())
}
