package data

import (
	"context"

	discoveryv1alpha2 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1alpha2"
	skv2core "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func SelfSignedVirtualMesh(dynamicClient client.Client, name, namespace string, meshes []*skv2core.ObjectRef) (*v1alpha2.VirtualMesh, error) {
	hostnameSuffix, err := getTestHostnameSuffix(dynamicClient, meshes)
	if err != nil {
		return nil, err
	}
	return &v1alpha2.VirtualMesh{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "VirtualMesh",
			APIVersion: v1alpha2.SchemeGroupVersion.String(),
		},
		Spec: v1alpha2.VirtualMeshSpec{
			Meshes: meshes,
			MtlsConfig: &v1alpha2.VirtualMeshSpec_MTLSConfig{
				TrustModel: &v1alpha2.VirtualMeshSpec_MTLSConfig_Shared{Shared: &v1alpha2.VirtualMeshSpec_MTLSConfig_SharedTrust{
					RootCertificateAuthority: &v1alpha2.VirtualMeshSpec_RootCertificateAuthority{
						CaSource: &v1alpha2.VirtualMeshSpec_RootCertificateAuthority_Generated{
							Generated: &v1alpha2.VirtualMeshSpec_RootCertificateAuthority_SelfSignedCert{},
						},
					},
				}},
				AutoRestartPods: true,
			},
			Federation: &v1alpha2.VirtualMeshSpec_Federation{
				HostnameSuffix: hostnameSuffix,
			},
		},
	}, nil
}

// use a custom hostname suffix when testing against istio >= 1.8
func getTestHostnameSuffix(dynamicClient client.Client, meshes []*skv2core.ObjectRef) (string, error) {
	meshClient := discoveryv1alpha2.NewMeshClient(dynamicClient)
	// assume that all meshes are using the same istio version
	mesh, err := meshClient.GetMesh(context.TODO(), client.ObjectKey{Name: meshes[0].Name, Namespace: meshes[0].Namespace})
	if err != nil {
		return "", err
	}
	if mesh.Spec.GetIstio().GetSmartDnsProxyingEnabled() {
		return "soloio", nil
	}
	return "", nil
}
