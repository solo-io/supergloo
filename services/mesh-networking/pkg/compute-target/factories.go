package networking_multicluster

import (
	smh_security "github.com/solo-io/service-mesh-hub/pkg/api/security.smh.solo.io/v1alpha1"
	controller_factories "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/compute-target/controllers"
)

type ClientFactories struct {
	VirtualMeshCSRClientFactory smh_security.VirtualMeshCertificateSigningRequestClientFactory
}

type ControllerFactories struct {
	VirtualMeshCSRControllerFactory controller_factories.VirtualMeshCSRControllerFactory
}
