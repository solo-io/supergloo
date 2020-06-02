package cert_signer

import (
	"context"

	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_security "github.com/solo-io/service-mesh-hub/pkg/api/security.zephyr.solo.io/v1alpha1"
	zephyr_security_types "github.com/solo-io/service-mesh-hub/pkg/api/security.zephyr.solo.io/v1alpha1/types"
	cert_secrets "github.com/solo-io/service-mesh-hub/pkg/csr/certgen/secrets"
)

//go:generate mockgen -destination ./mocks/mock_interfaces.go -source ./interfaces.go

/*
	VirtualMeshCertClient is a higher-level client meant to abstract common certificate related mesh actions
*/
type VirtualMeshCertClient interface {
	GetRootCaBundle(ctx context.Context, meshRef *zephyr_core_types.ResourceRef) (*cert_secrets.RootCAData, error)
}

type VirtualMeshCSRSigner interface {
	Sign(
		ctx context.Context,
		obj *zephyr_security.VirtualMeshCertificateSigningRequest,
	) *zephyr_security_types.VirtualMeshCertificateSigningRequestStatus
}
