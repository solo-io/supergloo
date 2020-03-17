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

func DefaultRootCaName(vm *networking_v1alpha1.VirtualMesh) string {
	return fmt.Sprintf("%s-ca-certs", vm.GetName())
}

type virtualMeshCertClient struct {
	localSecretClient      kubernetes_core.SecretsClient
	localVirtualMeshClient zephyr_networking.VirtualMeshClient
}

func NewVirtualMeshCertClient(
	localSecretClient kubernetes_core.SecretsClient,
	localVirtualMeshClient zephyr_networking.VirtualMeshClient) VirtualMeshCertClient {
	return &virtualMeshCertClient{
		localSecretClient:      localSecretClient,
		localVirtualMeshClient: localVirtualMeshClient,
	}
}

func (m *virtualMeshCertClient) GetRootCaBundle(
	ctx context.Context,
	meshRef *core_types.ResourceRef,
) (*cert_secrets.RootCaData, error) {
	vm, err := m.localVirtualMeshClient.Get(ctx, meshRef.GetName(), meshRef.GetNamespace())
	if err != nil {
		return nil, err
	}
	trustBundleSecretRef := &core_types.ResourceRef{
		Name:      DefaultRootCaName(vm),
		Namespace: env.GetWriteNamespace(),
	}
	if vm.Spec.GetTrustBundleRef() != nil {
		trustBundleSecretRef = vm.Spec.GetTrustBundleRef()
	}
	caSecret, err := m.localSecretClient.Get(ctx, trustBundleSecretRef.GetName(), trustBundleSecretRef.GetNamespace())
	if err != nil {
		return nil, err
	}
	return cert_secrets.RootCaDataFromSecret(caSecret)
}
