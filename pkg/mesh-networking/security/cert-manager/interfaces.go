package cert_manager

import (
	"context"

	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	smh_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/types"
	smh_security_types "github.com/solo-io/service-mesh-hub/pkg/api/security.smh.solo.io/v1alpha1/types"
)

//go:generate mockgen -destination ./mocks/mock_interfaces.go -source ./interfaces.go

/*
	Plugins to configure the certificate info on a VirtualMeshCertificateSigningRequest.
	At the moment the producers for each mesh (currently istio) are hardcoded into the handler itself.
*/
type CertConfigProducer interface {
	ConfigureCertificateInfo(
		vm *smh_networking.VirtualMesh,
		mesh *smh_discovery.Mesh,
	) (*smh_security_types.VirtualMeshCertificateSigningRequestSpec_CertConfig, error)
}

// VirtualMeshCertificateManager is the higher level event handler interface for VirtualMeshes
type VirtualMeshCertificateManager interface {
	InitializeCertificateForVirtualMesh(
		ctx context.Context,
		new *smh_networking.VirtualMesh,
	) smh_networking_types.VirtualMeshStatus
}
