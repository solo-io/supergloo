package networking_multicluster

import (
	zephyr_security "github.com/solo-io/service-mesh-hub/pkg/clients/zephyr/security"
	controller_factories "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/multicluster/controllers"
)

type ClientFactories struct {
	VirtualMeshCSRClientFactory zephyr_security.VirtualMeshCSRClientFactory
}

type ControllerFactories struct {
	VirtualMeshCSRControllerFactory controller_factories.VirtualMeshCSRControllerFactory
}
