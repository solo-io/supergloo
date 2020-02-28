package networking_multicluster

import (
	"github.com/google/wire"
	zephyr_security "github.com/solo-io/mesh-projects/pkg/clients/zephyr/security"
	controller_factories "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/multicluster/controllers"
)

var (
	ClientFactoryProviderSet = wire.NewSet(
		zephyr_security.MeshGroupCertificateSigningRequestClientFactoryProvider,
		NewClientFactories,
	)
)

type ClientFactories struct {
	CSRClientFactory zephyr_security.MeshGroupCertificateSigningRequestClientFactory
}

func NewClientFactories(CSRClientFactory zephyr_security.MeshGroupCertificateSigningRequestClientFactory) *ClientFactories {
	return &ClientFactories{CSRClientFactory: CSRClientFactory}
}

var (
	ControllerFactoryProviderSet = wire.NewSet(
		controller_factories.NewMeshGroupCertificateSigningRequestControllerFactory,
		NewControllerFactories,
	)
)

type ControllerFactories struct {
	CSRControllerFactory controller_factories.MeshGroupCertificateSigningRequestControllerFactory
}

func NewControllerFactories(CSRControllerFactory controller_factories.MeshGroupCertificateSigningRequestControllerFactory) *ControllerFactories {
	return &ControllerFactories{CSRControllerFactory: CSRControllerFactory}
}
