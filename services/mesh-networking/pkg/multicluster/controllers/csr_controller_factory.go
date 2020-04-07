package controller_factories

import (
	cert_controller "github.com/solo-io/service-mesh-hub/pkg/api/security.zephyr.solo.io/v1alpha1/controller"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/multicluster/manager"
)

// given a manager that can talk to a cluster and a name for that cluster, produce a `CertificateSigningRequestController`
type VirtualMeshCSRControllerFactory func(
	mgr mc_manager.AsyncManager,
	clusterName string,
) (cert_controller.VirtualMeshCertificateSigningRequestController, error)

func NewVirtualMeshCSRControllerFactory() VirtualMeshCSRControllerFactory {
	return func(
		mgr mc_manager.AsyncManager,
		clusterName string,
	) (controller cert_controller.VirtualMeshCertificateSigningRequestController, err error) {
		// just directly return the generated autopilot implementation
		return cert_controller.NewVirtualMeshCertificateSigningRequestController(clusterName, mgr.Manager())
	}
}
