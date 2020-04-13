package controller_factories

import (
	security_controller "github.com/solo-io/service-mesh-hub/pkg/api/security.zephyr.solo.io/v1alpha1/controller"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/multicluster/manager"
)

// given a manager that can talk to a cluster and a name for that cluster, produce a `CertificateSigningRequestController`
type VirtualMeshCSRControllerFactory func(
	mgr mc_manager.AsyncManager,
	clusterName string,
) security_controller.VirtualMeshCertificateSigningRequestEventWatcher

func NewVirtualMeshCSRControllerFactory() VirtualMeshCSRControllerFactory {
	return func(
		mgr mc_manager.AsyncManager,
		clusterName string,
	) security_controller.VirtualMeshCertificateSigningRequestEventWatcher {
		// just directly return the generated autopilot implementation
		return security_controller.NewVirtualMeshCertificateSigningRequestEventWatcher(clusterName, mgr.Manager())
	}
}
