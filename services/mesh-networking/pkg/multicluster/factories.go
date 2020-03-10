package networking_multicluster

import (
	zephyr_security "github.com/solo-io/mesh-projects/pkg/clients/zephyr/security"
	controller_factories "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/multicluster/controllers"
)

type ClientFactories struct {
	MeshGroupCSRClientFactory zephyr_security.MeshGroupCSRClientFactory
}

type ControllerFactories struct {
	MeshGroupCSRControllerFactory controller_factories.MeshGroupCSRControllerFactory
}
