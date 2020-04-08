package cert_manager

import (
	"context"

	discovery_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	networking_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	security_types "github.com/solo-io/service-mesh-hub/pkg/api/security.zephyr.solo.io/v1alpha1/types"
)

//go:generate mockgen -destination ./mocks/mock_interfaces.go -source ./interfaces.go

/*
	Plugins to configure the certificate info on a VirtualMeshCertificateSigningRequest.
	At the moment the producers for each mesh (currently istio) are hardcoded into the handler itself.
*/
type CertConfigProducer interface {
	ConfigureCertificateInfo(
		vm *networking_v1alpha1.VirtualMesh,
		mesh *discovery_v1alpha1.Mesh,
	) (*security_types.VirtualMeshCertificateSigningRequestSpec_CertConfig, error)
}

// VirtualMeshCertificateManager is the higher level event handler interface for VirtualMeshes
type VirtualMeshCertificateManager interface {
	InitializeCertificateForVirtualMesh(
		ctx context.Context,
		new *networking_v1alpha1.VirtualMesh,
	) networking_types.VirtualMeshStatus
}
