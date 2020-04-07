package cert_signer

import (
	"context"

	core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	security_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/security.zephyr.solo.io/v1alpha1"
	security_types "github.com/solo-io/service-mesh-hub/pkg/api/security.zephyr.solo.io/v1alpha1/types"
	cert_secrets "github.com/solo-io/service-mesh-hub/pkg/security/secrets"
)

//go:generate mockgen -destination ./mocks/mock_interfaces.go -source ./interfaces.go

/*
	VirtualMeshCertClient is a higher-level client meant to abstract common certificate related mesh actions
*/
type VirtualMeshCertClient interface {
	GetRootCaBundle(ctx context.Context, meshRef *core_types.ResourceRef) (*cert_secrets.RootCAData, error)
}

type VirtualMeshCSRSigner interface {
	Sign(
		ctx context.Context,
		obj *security_v1alpha1.VirtualMeshCertificateSigningRequest,
	) *security_types.VirtualMeshCertificateSigningRequestStatus
}
