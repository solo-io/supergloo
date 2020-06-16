package networking_multicluster

import (
	smh_security_providers "github.com/solo-io/service-mesh-hub/pkg/api/security.smh.solo.io/v1alpha1/providers"
	controller_factories "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/compute-target/controllers"
)

type ClientFactories struct {
	VirtualMeshCSRClientFactory smh_security_providers.VirtualMeshCertificateSigningRequestClientFactory
}

type ControllerFactories struct {
	VirtualMeshCSRControllerFactory controller_factories.VirtualMeshCSRControllerFactory
}
