package controller_factories

import (
	networking_controller "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/controller"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/multicluster/manager"
)

func NewLocalVirtualMeshController(mgr mc_manager.AsyncManager) (networking_controller.VirtualMeshController, error) {
	return networking_controller.NewVirtualMeshController("management-plane-virtual-mesh-controller", mgr.Manager())
}
