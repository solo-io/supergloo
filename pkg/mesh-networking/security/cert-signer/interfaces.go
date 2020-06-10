package cert_signer

import (
	"context"

	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_security "github.com/solo-io/service-mesh-hub/pkg/api/security.smh.solo.io/v1alpha1"
	smh_security_types "github.com/solo-io/service-mesh-hub/pkg/api/security.smh.solo.io/v1alpha1/types"
	cert_secrets "github.com/solo-io/service-mesh-hub/pkg/common/csr/certgen/secrets"
)

//go:generate mockgen -destination ./mocks/mock_interfaces.go -source ./interfaces.go

/*
	VirtualMeshCertClient is a higher-level client meant to abstract common certificate related mesh actions
*/
type VirtualMeshCertClient interface {
	GetRootCaBundle(ctx context.Context, meshRef *smh_core_types.ResourceRef) (*cert_secrets.RootCAData, error)
}

type VirtualMeshCSRSigner interface {
	Sign(
		ctx context.Context,
		obj *smh_security.VirtualMeshCertificateSigningRequest,
	) *smh_security_types.VirtualMeshCertificateSigningRequestStatus
}
