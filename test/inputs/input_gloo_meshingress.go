package inputs

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

func GlooMeshIngress(namespace string, meshes []*core.ResourceRef) *v1.MeshIngress {
	return GlooMeshIngressInstallNs(namespace, "gloo-was-installed-herr", meshes)
}

func GlooMeshIngressInstallNs(namespace, installNs string, meshes []*core.ResourceRef) *v1.MeshIngress {
	return &v1.MeshIngress{
		Metadata: core.Metadata{
			Namespace: namespace,
			Name:      "fancy-gloo",
		},
		MeshIngressType: &v1.MeshIngress_Gloo{
			Gloo: &v1.GlooMeshIngress{},
		},
		InstallationNamespace: installNs,
		Meshes:                meshes,
	}
}
