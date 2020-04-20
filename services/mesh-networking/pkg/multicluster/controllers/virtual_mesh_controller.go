package controller_factories

import (
	zephyr_networking_controller "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/controller"
	"github.com/solo-io/service-mesh-hub/services/common/multicluster/manager/k8s_manager"
)

func NewLocalVirtualMeshEventWatcher(mgr k8s_manager.AsyncManager) zephyr_networking_controller.VirtualMeshEventWatcher {
	return zephyr_networking_controller.NewVirtualMeshEventWatcher("management-plane-virtual-mesh-event-watcher", mgr.Manager())
}
