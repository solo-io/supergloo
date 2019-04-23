package inputs

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

func IstioInstall(name, namespace, installNs, version string, disabled bool) *v1.Install {
	return &v1.Install{
		Metadata:              core.Metadata{Name: name, Namespace: namespace},
		Disabled:              disabled,
		InstallationNamespace: installNs,
		InstallType: &v1.Install_Mesh{
			Mesh: &v1.MeshInstall{
				MeshInstallType: &v1.MeshInstall_Istio{
					Istio: &v1.IstioInstall{
						Version: version,
					},
				},
			},
		},
	}
}

func LinkerdInstall(name, namespace, installNs, version string, disabled bool) *v1.Install {
	return &v1.Install{
		Metadata:              core.Metadata{Name: name, Namespace: namespace},
		Disabled:              disabled,
		InstallationNamespace: installNs,
		InstallType: &v1.Install_Mesh{
			Mesh: &v1.MeshInstall{
				MeshInstallType: &v1.MeshInstall_Linkerd{
					Linkerd: &v1.LinkerdInstall{
						EnableAutoInject: true,
						EnableMtls:       true,
						Version:          version,
					},
				},
			},
		},
	}
}

func GlooIstall(name, namespace, installNs, version string, disabled bool) *v1.Install {
	return &v1.Install{
		Metadata:              core.Metadata{Name: name, Namespace: namespace},
		Disabled:              disabled,
		InstallationNamespace: installNs,
		InstallType: &v1.Install_Ingress{
			Ingress: &v1.MeshIngressInstall{
				IngressInstallType: &v1.MeshIngressInstall_Gloo{
					Gloo: &v1.GlooInstall{
						Version: version,
					},
				},
			},
		},
	}
}

func GlooIstallWithMeshes(name, namespace, installNs, version string, disabled bool, targetMeshes []*core.ResourceRef) *v1.Install {
	return &v1.Install{
		Metadata:              core.Metadata{Name: name, Namespace: namespace},
		Disabled:              disabled,
		InstallationNamespace: installNs,
		InstallType: &v1.Install_Ingress{
			Ingress: &v1.MeshIngressInstall{
				IngressInstallType: &v1.MeshIngressInstall_Gloo{
					Gloo: &v1.GlooInstall{
						Version: version,
						Meshes:  targetMeshes,
					},
				},
			},
		},
	}
}
