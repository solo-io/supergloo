package security_processors

import (
	"context"

	security_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1"
	security_types "github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1/types"
)

//go:generate mockgen -destination ./mocks/mock_interfaces.go -source ./interfaces.go

/*
	MeshGroupCertificateSigningRequestProcessor is meant to be an extension to the autopilot handler pattern.
	The status returned by a processor will be updated to the cluster upon return.
	The contract of how else that data is used by the caller is up to the caller.
*/
type MeshGroupCertificateSigningRequestProcessor interface {
	ProcessCreate(
		ctx context.Context,
		csr *security_v1alpha1.MeshGroupCertificateSigningRequest,
	) security_types.MeshGroupCertificateSigningRequestStatus
	ProcessUpdate(
		ctx context.Context,
		old, new *security_v1alpha1.MeshGroupCertificateSigningRequest,
	) security_types.MeshGroupCertificateSigningRequestStatus
	ProcessDelete(
		ctx context.Context,
		csr *security_v1alpha1.MeshGroupCertificateSigningRequest,
	) security_types.MeshGroupCertificateSigningRequestStatus
}
