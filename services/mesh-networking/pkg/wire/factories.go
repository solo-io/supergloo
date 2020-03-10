package wire

import (
	"github.com/google/wire"
	zephyr_security "github.com/solo-io/mesh-projects/pkg/clients/zephyr/security"
	networking_multicluster "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/multicluster"
	controller_factories "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/multicluster/controllers"
)

var (
	ClientFactoryProviderSet = wire.NewSet(
		zephyr_security.MeshGroupCSRClientFactoryProvider,
		NewClientFactories,
	)

	ControllerFactoryProviderSet = wire.NewSet(
		controller_factories.NewMeshGroupCSRControllerFactory,
		NewControllerFactories,
	)
)

func NewClientFactories(
	MeshGroupCertificateSigningRequestClientFactory zephyr_security.MeshGroupCSRClientFactory,
) *networking_multicluster.ClientFactories {
	return &networking_multicluster.ClientFactories{
		MeshGroupCSRClientFactory: MeshGroupCertificateSigningRequestClientFactory,
	}
}

func NewControllerFactories(
	CSRControllerFactory controller_factories.MeshGroupCSRControllerFactory,
) *networking_multicluster.ControllerFactories {
	return &networking_multicluster.ControllerFactories{MeshGroupCSRControllerFactory: CSRControllerFactory}
}
