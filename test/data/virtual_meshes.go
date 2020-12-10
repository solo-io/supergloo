package data

import (
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1alpha2"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func SelfSignedVirtualMesh(name, namespace string, meshes []*v1.ObjectRef) *v1alpha2.VirtualMesh {
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
			Federation: &v1alpha2.VirtualMeshSpec_Federation{},
		},
	}
}

// LimitedTrustSelfSignedVirtualMesh creates a virtual mesh manifest that uses limited trust model
func LimitedTrustSelfSignedVirtualMesh(name, namespace string, meshes []*v1.ObjectRef) *v1alpha2.VirtualMesh {
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
				TrustModel: &v1alpha2.VirtualMeshSpec_MTLSConfig_Limited{Limited: &v1alpha2.VirtualMeshSpec_MTLSConfig_LimitedTrust{
					RootCertificateAuthority: &v1alpha2.VirtualMeshSpec_RootCertificateAuthority{
						CaSource: &v1alpha2.VirtualMeshSpec_RootCertificateAuthority_Generated{
							Generated: &v1alpha2.VirtualMeshSpec_RootCertificateAuthority_SelfSignedCert{},
						},
					},
				}},
				AutoRestartPods: false,
			},
			Federation: &v1alpha2.VirtualMeshSpec_Federation{},
		},
	}
}
