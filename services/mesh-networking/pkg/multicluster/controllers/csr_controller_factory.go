package controller_factories

import (
	zephyr_security_controller "github.com/solo-io/service-mesh-hub/pkg/api/security.zephyr.solo.io/v1alpha1/controller"
	"github.com/solo-io/service-mesh-hub/services/common/multicluster/manager/k8s_manager"
)

// given a manager that can talk to a cluster and a name for that cluster, produce a `CertificateSigningRequestController`
type VirtualMeshCSRControllerFactory func(
	mgr k8s_manager.AsyncManager,
	clusterName string,
) zephyr_security_controller.VirtualMeshCertificateSigningRequestEventWatcher

func NewVirtualMeshCSRControllerFactory() VirtualMeshCSRControllerFactory {
	return func(
		mgr k8s_manager.AsyncManager,
		clusterName string,
	) zephyr_security_controller.VirtualMeshCertificateSigningRequestEventWatcher {
		// just directly return the generated autopilot implementation
		return zephyr_security_controller.NewVirtualMeshCertificateSigningRequestEventWatcher(clusterName, mgr.Manager())
	}
}
