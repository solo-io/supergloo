package util

import (
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

func GetMeshInstallatioNamespace(mesh *v1.Mesh) string {
	var result string
	if mesh.DiscoveryMetadata != nil {
		result = mesh.DiscoveryMetadata.InstallationNamespace
	}

	var installOptions *v1.InstallOptions
	switch meshType := mesh.GetMeshType().(type) {
	case *v1.Mesh_Istio:
		installOptions = meshType.Istio.Install.Options
	case *v1.Mesh_Linkerd:
		installOptions = meshType.Linkerd.Install.Options
	}
	if installOptions != nil {
		result = installOptions.InstallationNamespace
	}
	return result
}
