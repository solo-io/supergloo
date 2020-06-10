package wire

import (
	"github.com/google/wire"
	smh_security "github.com/solo-io/service-mesh-hub/pkg/api/security.smh.solo.io/v1alpha1"
	networking_multicluster "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/compute-target"
	controller_factories "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/compute-target/controllers"
)

var (
	ClientFactoryProviderSet = wire.NewSet(
		smh_security.VirtualMeshCertificateSigningRequestClientFactoryProvider,
		NewClientFactories,
	)

	ControllerFactoryProviderSet = wire.NewSet(
		controller_factories.NewVirtualMeshCSRControllerFactory,
		NewControllerFactories,
	)
)

func NewClientFactories(
	VirtualMeshCertificateSigningRequestClientFactory smh_security.VirtualMeshCertificateSigningRequestClientFactory,
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
