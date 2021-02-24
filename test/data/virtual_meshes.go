package data

import (
	"context"

	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	v1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	skv2corev1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func SelfSignedVirtualMesh(
	dynamicClient client.Client,
	name, namespace string,
	meshes []*skv2corev1.ObjectRef,
	flatNetwork bool,
) (*v1.VirtualMesh, error) {
	hostnameSuffix, err := getTestHostnameSuffix(dynamicClient, meshes)
	if err != nil {
		return nil, err
	}
	return &v1.VirtualMesh{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "VirtualMesh",
			APIVersion: v1.SchemeGroupVersion.String(),
		},
		Spec: v1.VirtualMeshSpec{
			Meshes: meshes,
			MtlsConfig: &v1.VirtualMeshSpec_MTLSConfig{
				TrustModel: &v1.VirtualMeshSpec_MTLSConfig_Shared{Shared: &v1.VirtualMeshSpec_MTLSConfig_SharedTrust{
					RootCertificateAuthority: &v1.VirtualMeshSpec_RootCertificateAuthority{
						CaSource: &v1.VirtualMeshSpec_RootCertificateAuthority_Generated{
							Generated: &v1.VirtualMeshSpec_RootCertificateAuthority_SelfSignedCert{},
						},
					},
				}},
				AutoRestartPods: true,
			},
			Federation: &v1.VirtualMeshSpec_Federation{
				HostnameSuffix: hostnameSuffix,
				FlatNetwork:    flatNetwork,
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
