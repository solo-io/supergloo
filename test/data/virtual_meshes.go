package data

import (
	"context"

	certificatesv1 "github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/v1"
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	networkingv1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	skv2corev1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func SelfSignedVirtualMesh(
	dynamicClient client.Client,
	name, namespace string,
	meshes []*skv2corev1.ObjectRef,
	flatNetwork bool,
) (*networkingv1.VirtualMesh, error) {
	hostnameSuffix, err := getTestHostnameSuffix(dynamicClient, meshes)
	if err != nil {
		return nil, err
	}
	return &networkingv1.VirtualMesh{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "VirtualMesh",
			APIVersion: networkingv1.SchemeGroupVersion.String(),
		},
		Spec: networkingv1.VirtualMeshSpec{
			Meshes: meshes,
			MtlsConfig: &networkingv1.VirtualMeshSpec_MTLSConfig{
				TrustModel: &networkingv1.VirtualMeshSpec_MTLSConfig_Shared{
					Shared: &networkingv1.SharedTrust{
						CertificateAuthority: &networkingv1.SharedTrust_RootCertificateAuthority{
							RootCertificateAuthority: &networkingv1.RootCertificateAuthority{
								CaSource: &networkingv1.RootCertificateAuthority_Generated{
									Generated: &certificatesv1.CommonCertOptions{},
								},
							},
						},
					},
				},
				AutoRestartPods: true,
			},
			Federation: &networkingv1.VirtualMeshSpec_Federation{
				Selectors: []*networkingv1.VirtualMeshSpec_Federation_FederationSelector{
					{}, // permissive federation
				},
				FlatNetwork:    flatNetwork,
				HostnameSuffix: hostnameSuffix,
			},
		},
	}, nil
}

// use a custom hostname suffix when testing against istio >= 1.8
func getTestHostnameSuffix(dynamicClient client.Client, meshes []*skv2corev1.ObjectRef) (string, error) {
	meshClient := discoveryv1.NewMeshClient(dynamicClient)
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
