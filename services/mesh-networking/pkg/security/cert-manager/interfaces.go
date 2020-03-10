package cert_manager

import (
	"context"

	discovery_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	networking_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	networking_types "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	security_types "github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1/types"
)

//go:generate mockgen -destination ./mocks/mock_interfaces.go -source ./interfaces.go

/*
	Plugins to configure the certificate info on a MeshGroupCertificateSigningRequest.
	At the moment the producers for each mesh (currently istio) are hardcoded into the handler itself.
*/
type CertConfigProducer interface {
	ConfigureCertificateInfo(
		mg *networking_v1alpha1.MeshGroup,
		mesh *discovery_v1alpha1.Mesh,
	) (*security_types.CertConfig, error)
}

// MeshGroupCertificateManager is the higher level event handler interface for MeshGroups
type MeshGroupCertificateManager interface {
	InitializeCertificateForMeshGroup(
		ctx context.Context,
		new *networking_v1alpha1.MeshGroup,
	) networking_types.MeshGroupStatus
}
