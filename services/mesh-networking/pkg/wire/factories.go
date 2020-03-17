package wire

import (
	"github.com/google/wire"
	zephyr_security "github.com/solo-io/mesh-projects/pkg/clients/zephyr/security"
	networking_multicluster "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/multicluster"
	controller_factories "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/multicluster/controllers"
)

var (
	ClientFactoryProviderSet = wire.NewSet(
		zephyr_security.VirtualMeshCSRClientFactoryProvider,
		NewClientFactories,
	)

	ControllerFactoryProviderSet = wire.NewSet(
		controller_factories.NewVirtualMeshCSRControllerFactory,
		NewControllerFactories,
	)
)

func NewClientFactories(
	VirtualMeshCertificateSigningRequestClientFactory zephyr_security.VirtualMeshCSRClientFactory,
) *networking_multicluster.ClientFactories {
	return &networking_multicluster.ClientFactories{
		VirtualMeshCSRClientFactory: VirtualMeshCertificateSigningRequestClientFactory,
	}
}

func NewControllerFactories(
	CSRControllerFactory controller_factories.VirtualMeshCSRControllerFactory,
) *networking_multicluster.ControllerFactories {
	return &networking_multicluster.ControllerFactories{VirtualMeshCSRControllerFactory: CSRControllerFactory}
}
