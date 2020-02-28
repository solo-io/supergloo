package csr_agent_controller

import (
	"context"

	securityv1alpha1 "github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1"
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
		obj *securityv1alpha1.MeshGroupCertificateSigningRequest,
	) (secret *cert_secrets.RootCaData, err error)
}
