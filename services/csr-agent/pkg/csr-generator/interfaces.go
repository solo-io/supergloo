package csr_generator

import (
	"context"

	zephyr_security "github.com/solo-io/service-mesh-hub/pkg/api/security.zephyr.solo.io/v1alpha1"
	zephyr_security_controller "github.com/solo-io/service-mesh-hub/pkg/api/security.zephyr.solo.io/v1alpha1/controller"
	zephyr_security_types "github.com/solo-io/service-mesh-hub/pkg/api/security.zephyr.solo.io/v1alpha1/types"
	cert_secrets "github.com/solo-io/service-mesh-hub/pkg/security/secrets"
)

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_mesh_interfaces.go

// CertClient is a higher level client used for operations involving VirtualMeshCertificateSigningRequests.

type CertClient interface {
	/*
		EnsureSecretKey retrieves the private key for a given VirtualMeshCertificateSigningRequest. If the key does not
		exist already, it will attempt to create one.
	*/
	EnsureSecretKey(
		ctx context.Context,
		obj *zephyr_security.VirtualMeshCertificateSigningRequest,
	) (secret *cert_secrets.IntermediateCAData, err error)
}

type PrivateKeyGenerator interface {
	// Generate an RSA private key and return as pem encoded bytes.
	GenerateRSA(keySize int) ([]byte, error)
}

type IstioCSRGenerator interface {
	GenerateIstioCSR(
		ctx context.Context,
		obj *zephyr_security.VirtualMeshCertificateSigningRequest,
	) *zephyr_security_types.VirtualMeshCertificateSigningRequestStatus
}

type VirtualMeshCSRDataSourceFactory func(
	ctx context.Context,
	csrClient zephyr_security.VirtualMeshCertificateSigningRequestClient,
	processor VirtualMeshCSRProcessor,
) zephyr_security_controller.VirtualMeshCertificateSigningRequestEventHandler

/*
	VirtualMeshCSRProcessor is meant to be an extension to the autopilot handler pattern.

	The status returned by a processor is intended to reflect the status of the object.
	If the returned status is nil than the object was not processed and the status should not be used or updated.
	The contract of how else that data is used by the caller is up to the caller.
*/
type VirtualMeshCSRProcessor interface {
	ProcessUpsert(
		ctx context.Context,
		csr *zephyr_security.VirtualMeshCertificateSigningRequest,
	) *zephyr_security_types.VirtualMeshCertificateSigningRequestStatus
	ProcessDelete(
		ctx context.Context,
		csr *zephyr_security.VirtualMeshCertificateSigningRequest,
	) *zephyr_security_types.VirtualMeshCertificateSigningRequestStatus
}

type VirtualMeshCSRProcessorFuncs struct {
	OnProcessUpsert func(
		ctx context.Context,
		csr *zephyr_security.VirtualMeshCertificateSigningRequest,
	) *zephyr_security_types.VirtualMeshCertificateSigningRequestStatus
	OnProcessDelete func(
		ctx context.Context,
		csr *zephyr_security.VirtualMeshCertificateSigningRequest,
	) *zephyr_security_types.VirtualMeshCertificateSigningRequestStatus
}

func (m *VirtualMeshCSRProcessorFuncs) ProcessUpsert(
	ctx context.Context,
	csr *zephyr_security.VirtualMeshCertificateSigningRequest,
) *zephyr_security_types.VirtualMeshCertificateSigningRequestStatus {
	if m.OnProcessUpsert != nil {
		return m.OnProcessUpsert(ctx, csr)
	}
	return nil
}

func (m *VirtualMeshCSRProcessorFuncs) ProcessDelete(
	ctx context.Context,
	csr *zephyr_security.VirtualMeshCertificateSigningRequest,
) *zephyr_security_types.VirtualMeshCertificateSigningRequestStatus {
	if m.OnProcessDelete != nil {
		return m.OnProcessDelete(ctx, csr)
	}
	return nil
}
