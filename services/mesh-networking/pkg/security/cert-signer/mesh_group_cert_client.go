package cert_signer

import (
	"context"
	"fmt"

	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	networking_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	kubernetes_core "github.com/solo-io/mesh-projects/pkg/clients/kubernetes/core"
	zephyr_networking "github.com/solo-io/mesh-projects/pkg/clients/zephyr/networking"
	"github.com/solo-io/mesh-projects/pkg/env"
	cert_secrets "github.com/solo-io/mesh-projects/pkg/security/secrets"
)

func DefaultRootCaName(mg *networking_v1alpha1.MeshGroup) string {
	return fmt.Sprintf("%s-ca-certs", mg.GetName())
}

type meshGroupCertClient struct {
	localSecretClient    kubernetes_core.SecretsClient
	localMeshGroupClient zephyr_networking.MeshGroupClient
}

func NewMeshGroupCertClient(
	localSecretClient kubernetes_core.SecretsClient,
	localMeshGroupClient zephyr_networking.MeshGroupClient) MeshGroupCertClient {
	return &meshGroupCertClient{
		localSecretClient:    localSecretClient,
		localMeshGroupClient: localMeshGroupClient,
	}
}

func (m *meshGroupCertClient) GetRootCaBundle(
	ctx context.Context,
	meshRef *core_types.ResourceRef,
) (*cert_secrets.RootCaData, error) {
	mg, err := m.localMeshGroupClient.Get(ctx, meshRef.GetName(), meshRef.GetNamespace())
	if err != nil {
		return nil, err
	}
	trustBundleSecretRef := &core_types.ResourceRef{
		Name:      DefaultRootCaName(mg),
		Namespace: env.GetWriteNamespace(),
	}
	if mg.Spec.GetTrustBundleRef() != nil {
		trustBundleSecretRef = mg.Spec.GetTrustBundleRef()
	}
	caSecret, err := m.localSecretClient.Get(ctx, trustBundleSecretRef.GetName(), trustBundleSecretRef.GetNamespace())
	if err != nil {
		return nil, err
	}
	return cert_secrets.RootCaDataFromSecret(caSecret)
}
