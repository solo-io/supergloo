package controller_factories

import (
	cert_controller "github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1/controller"
	mc_manager "github.com/solo-io/mesh-projects/services/common/multicluster/manager"
)

// given a manager that can talk to a cluster and a name for that cluster, produce a `CertificateSigningRequestController`
type MeshGroupCertificateSigningRequestControllerFactory func(
	mgr mc_manager.AsyncManager,
	clusterName string,
) (cert_controller.MeshGroupCertificateSigningRequestController, error)

func NewMeshGroupCertificateSigningRequestControllerFactory() MeshGroupCertificateSigningRequestControllerFactory {
	return func(
		mgr mc_manager.AsyncManager,
		clusterName string,
	) (controller cert_controller.MeshGroupCertificateSigningRequestController, err error) {
		// just directly return the generated autopilot implementation
		return cert_controller.NewMeshGroupCertificateSigningRequestController(clusterName, mgr.Manager())
	}
}
