package cert_signer

import (
	"context"
	"fmt"

	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	k8s_core "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	smh_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/types"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/container-runtime"
	"github.com/solo-io/service-mesh-hub/pkg/csr/certgen"
	cert_secrets "github.com/solo-io/service-mesh-hub/pkg/csr/certgen/secrets"
	"github.com/solo-io/service-mesh-hub/pkg/kube/selection"
	k8s_core_types "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func DefaultRootCaName(vm *smh_networking.VirtualMesh) string {
	return fmt.Sprintf("%s-ca-certs", vm.GetName())
}

type virtualMeshCertClient struct {
	localSecretClient      k8s_core.SecretClient
	localVirtualMeshClient smh_networking.VirtualMeshClient
	rootCertGenerator      certgen.RootCertGenerator
}

func NewVirtualMeshCertClient(
	localSecretClient k8s_core.SecretClient,
	localVirtualMeshClient smh_networking.VirtualMeshClient,
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
	meshRef *smh_core_types.ResourceRef,
) (*cert_secrets.RootCAData, error) {
	vm, err := v.localVirtualMeshClient.GetVirtualMesh(ctx, selection.ResourceRefToObjectKey(meshRef))
	if err != nil {
		return nil, err
	}
	var caSecret *k8s_core_types.Secret
	var trustBundleSecretRef *smh_core_types.ResourceRef
	switch vm.Spec.GetCertificateAuthority().GetType().(type) {
	case *smh_networking_types.VirtualMeshSpec_CertificateAuthority_Provided_:
		trustBundleSecretRef = vm.Spec.GetCertificateAuthority().GetProvided().GetCertificate()
		caSecret, err = v.localSecretClient.GetSecret(
			ctx, client.ObjectKey{Name: trustBundleSecretRef.GetName(), Namespace: trustBundleSecretRef.GetNamespace()})
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
	vm *smh_networking.VirtualMesh,
) (*k8s_core_types.Secret, error) {
	// auto-generated cert lives in fixed location
	rootCaName := DefaultRootCaName(vm)
	rootCaNamespace := container_runtime.GetWriteNamespace()
	caSecret, err := v.localSecretClient.GetSecret(ctx, client.ObjectKey{Name: rootCaName, Namespace: rootCaNamespace})
	if errors.IsNotFound(err) {
		rootCaData, err := v.rootCertGenerator.GenRootCertAndKey(vm.Spec.GetCertificateAuthority().GetBuiltin())
		if err != nil {
			return nil, err
		}
		caSecret := rootCaData.BuildSecret(rootCaName, rootCaNamespace)
		if err = v.localSecretClient.CreateSecret(ctx, caSecret); err != nil {
			return nil, err
		}
		return caSecret, nil
	} else if err != nil {
		return nil, err
	} else {
		return caSecret, nil
	}
}
