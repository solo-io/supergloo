package controller_factories

import (
	networking_controller "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1/controller"
	mc_manager "github.com/solo-io/mesh-projects/services/common/multicluster/manager"
)

// given a manager that can talk to a cluster and a name for that cluster, produce a `VirtualMeshController`
type VirtualMeshControllerFactory func(
	mgr mc_manager.AsyncManager,
	clusterName string,
) (networking_controller.VirtualMeshController, error)

func NewVirtualMeshControllerFactory() VirtualMeshControllerFactory {
	return func(
		mgr mc_manager.AsyncManager,
		clusterName string,
	) (controller networking_controller.VirtualMeshController, err error) {
		// just directly return the generated autopilot implementation
		return networking_controller.NewVirtualMeshController(clusterName, mgr.Manager())
	}
}
