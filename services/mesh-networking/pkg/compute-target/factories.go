package networking_multicluster

import (
	zephyr_security "github.com/solo-io/service-mesh-hub/pkg/api/security.zephyr.solo.io/v1alpha1"
	controller_factories "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/compute-target/controllers"
)

type ClientFactories struct {
	VirtualMeshCSRClientFactory zephyr_security.VirtualMeshCertificateSigningRequestClientFactory
}

type ControllerFactories struct {
	VirtualMeshCSRControllerFactory controller_factories.VirtualMeshCSRControllerFactory
}
