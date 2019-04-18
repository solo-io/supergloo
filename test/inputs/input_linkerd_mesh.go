package inputs

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

func LinkerdMesh(namespace string, secretRef *core.ResourceRef) *v1.Mesh {
	if secretRef != nil {
		panic("linkerd mesh does not currently support secret refs, so don't test it!")
	}
	return LinkerdMeshWithInstallNs(namespace, "linkerd-was-installed-herr", secretRef)
}

func LinkerdMeshWithInstallNs(namespace, installNs string, secretRef *core.ResourceRef) *v1.Mesh {
	return &v1.Mesh{
		Metadata: core.Metadata{
			Namespace: namespace,
			Name:      "fancy-linkerd",
		},
		MeshType: &v1.Mesh_LinkerdMesh{
			LinkerdMesh: &v1.LinkerdMesh{
				InstallationNamespace: installNs,
			},
		},
		MtlsConfig: &v1.MtlsConfig{
			MtlsEnabled:     true,
			RootCertificate: secretRef,
		},
	}
}

func LinkerdMeshWithInstallNsPrometheus(namespace, installNs string, secretRef *core.ResourceRef, promCfgRefs []core.ResourceRef) *v1.Mesh {
	var monit *v1.MonitoringConfig
	if len(promCfgRefs) > 0 {
		monit = &v1.MonitoringConfig{
			PrometheusConfigmaps: promCfgRefs,
		}
	}
	return &v1.Mesh{
		Metadata: core.Metadata{
			Namespace: namespace,
			Name:      "fancy-linkerd",
		},
		MeshType: &v1.Mesh_LinkerdMesh{
			LinkerdMesh: &v1.LinkerdMesh{
				InstallationNamespace: installNs,
			},
		},
		MtlsConfig: &v1.MtlsConfig{
			MtlsEnabled:     true,
			RootCertificate: secretRef,
		},
		MonitoringConfig: monit,
	}
}
