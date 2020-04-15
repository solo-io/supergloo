package cert_manager

import (
	"context"

	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
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
		vm *zephyr_networking.VirtualMesh,
		mesh *zephyr_discovery.Mesh,
	) (*security_types.VirtualMeshCertificateSigningRequestSpec_CertConfig, error)
}

// VirtualMeshCertificateManager is the higher level event handler interface for VirtualMeshes
type VirtualMeshCertificateManager interface {
	InitializeCertificateForVirtualMesh(
		ctx context.Context,
		new *zephyr_networking.VirtualMesh,
	) networking_types.VirtualMeshStatus
}
