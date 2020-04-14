package cert_signer

import (
	"context"
	"fmt"

	core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	networking_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/clients"
	kubernetes_core "github.com/solo-io/service-mesh-hub/pkg/clients/kubernetes/core"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	"github.com/solo-io/service-mesh-hub/pkg/security/certgen"
	cert_secrets "github.com/solo-io/service-mesh-hub/pkg/security/secrets"
	core_v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
)

func DefaultRootCaName(vm *networking_v1alpha1.VirtualMesh) string {
	return fmt.Sprintf("%s-ca-certs", vm.GetName())
}

type virtualMeshCertClient struct {
	localSecretClient      kubernetes_core.SecretClient
	localVirtualMeshClient zephyr_networking.VirtualMeshClient
	rootCertGenerator      certgen.RootCertGenerator
}

func NewVirtualMeshCertClient(
	localSecretClient kubernetes_core.SecretClient,
	localVirtualMeshClient zephyr_networking.VirtualMeshClient,
	rootCertGenerator certgen.RootCertGenerator,
) VirtualMeshCertClient {
	return &virtualMeshCertClient{
		localSecretClient:      localSecretClient,
		localVirtualMeshClient: localVirtualMeshClient,
		rootCertGenerator:      rootCertGenerator,
	}
}

// Fetch the root certificate, which can either be a resource ref to a user-supplied cert or an auto-generated cert.
func (v *virtualMeshCertClient) GetRootCaBundle(
	ctx context.Context,
	meshRef *core_types.ResourceRef,
) (*cert_secrets.RootCAData, error) {
	vm, err := v.localVirtualMeshClient.GetVirtualMesh(ctx, clients.ResourceRefToObjectKey(meshRef))
	if err != nil {
		return nil, err
	}
	var caSecret *core_v1.Secret
	var trustBundleSecretRef *core_types.ResourceRef
	switch vm.Spec.GetCertificateAuthority().GetType().(type) {
	case *types.VirtualMeshSpec_CertificateAuthority_Provided_:
		trustBundleSecretRef = vm.Spec.GetCertificateAuthority().GetProvided().GetCertificate()
		caSecret, err = v.localSecretClient.Get(ctx, trustBundleSecretRef.GetName(), trustBundleSecretRef.GetNamespace())
		if err != nil {
			return nil, err
		}
	default:
		caSecret, err = v.getOrCreateBuiltinRootCert(ctx, vm)
		if err != nil {
			return nil, err
		}
	}
	return cert_secrets.RootCADataFromSecret(caSecret)
}

func (v *virtualMeshCertClient) getOrCreateBuiltinRootCert(
	ctx context.Context,
	vm *networking_v1alpha1.VirtualMesh,
) (*core_v1.Secret, error) {
	// auto-generated cert lives in fixed location
	rootCaName := DefaultRootCaName(vm)
	rootCaNamespace := env.GetWriteNamespace()
	caSecret, err := v.localSecretClient.Get(ctx, rootCaName, rootCaNamespace)
	if errors.IsNotFound(err) {
		rootCaData, err := v.rootCertGenerator.GenRootCertAndKey(vm.Spec.GetCertificateAuthority().GetBuiltin())
		if err != nil {
			return nil, err
		}
		caSecret := rootCaData.BuildSecret(rootCaName, rootCaNamespace)
		if err = v.localSecretClient.Create(ctx, caSecret); err != nil {
			return nil, err
		}
		return caSecret, nil
	} else if err != nil {
		return nil, err
	} else {
		return caSecret, nil
	}
}
