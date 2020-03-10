package csr_generator

import (
	"context"

	security_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1/controller"
	security_types "github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1/types"
	zephyr_security "github.com/solo-io/mesh-projects/pkg/clients/zephyr/security"
	cert_secrets "github.com/solo-io/mesh-projects/pkg/security/secrets"
)

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_mesh_interfaces.go

// CertClient is a higher level client used for operations involving MeshGroupCertificateSigningRequests.

type CertClient interface {
	/*
		EnsureSecretKey retrieves the private key for a given MeshGroupCertificateSigningRequest. If the key does not
		exist already, it will attempt to create one.
	*/
	EnsureSecretKey(
		ctx context.Context,
		obj *security_v1alpha1.MeshGroupCertificateSigningRequest,
	) (secret *cert_secrets.RootCaData, err error)
}

type IstioCSRGenerator interface {
	GenerateIstioCSR(
		ctx context.Context,
		obj *security_v1alpha1.MeshGroupCertificateSigningRequest,
	) *security_types.MeshGroupCertificateSigningRequestStatus
}

type MeshGroupCSRDataSourceFactory func(
	ctx context.Context,
	csrClient zephyr_security.MeshGroupCSRClient,
	processor MeshGroupCSRProcessor,
) controller.MeshGroupCertificateSigningRequestEventHandler

/*
	MeshGroupCSRProcessor is meant to be an extension to the autopilot handler pattern.

	The status returned by a processor is intended to reflect the status of the object.
	If the returned status is nil than the object was not processed and the status should not be used or updated.
	The contract of how else that data is used by the caller is up to the caller.
*/
type MeshGroupCSRProcessor interface {
	ProcessUpsert(
		ctx context.Context,
		csr *security_v1alpha1.MeshGroupCertificateSigningRequest,
	) *security_types.MeshGroupCertificateSigningRequestStatus
	ProcessDelete(
		ctx context.Context,
		csr *security_v1alpha1.MeshGroupCertificateSigningRequest,
	) *security_types.MeshGroupCertificateSigningRequestStatus
}

type MeshGroupCSRProcessorFuncs struct {
	OnProcessUpsert func(
		ctx context.Context,
		csr *security_v1alpha1.MeshGroupCertificateSigningRequest,
	) *security_types.MeshGroupCertificateSigningRequestStatus
	OnProcessDelete func(
		ctx context.Context,
		csr *security_v1alpha1.MeshGroupCertificateSigningRequest,
	) *security_types.MeshGroupCertificateSigningRequestStatus
}

func (m *MeshGroupCSRProcessorFuncs) ProcessUpsert(
	ctx context.Context,
	csr *security_v1alpha1.MeshGroupCertificateSigningRequest,
) *security_types.MeshGroupCertificateSigningRequestStatus {
	if m.OnProcessUpsert != nil {
		return m.OnProcessUpsert(ctx, csr)
	}
	return nil
}

func (m *MeshGroupCSRProcessorFuncs) ProcessDelete(
	ctx context.Context,
	csr *security_v1alpha1.MeshGroupCertificateSigningRequest,
) *security_types.MeshGroupCertificateSigningRequestStatus {
	if m.OnProcessDelete != nil {
		return m.OnProcessDelete(ctx, csr)
	}
	return nil
}
