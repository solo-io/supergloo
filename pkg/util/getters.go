package util

import (
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

func getLinkerdMeshForInstall(install *v1.LinkerdInstall, meshes v1.MeshList, namespace string) *v1.Mesh {
	for _, mesh := range meshes {
		linkerdMesh := mesh.GetLinkerd()
		if linkerdMesh == nil {
			continue
		}

		if linkerdMesh.InstallationNamespace == namespace &&
			linkerdMesh.Version == install.Version {
			return mesh
		}
	}
	return nil
}

func getIstioMeshForInstall(install *v1.IstioInstall, meshes v1.MeshList, namespace string) *v1.Mesh {
	for _, mesh := range meshes {
		istioMesh := mesh.GetIstio()
		if istioMesh == nil {
			continue
		}

		if istioMesh.InstallationNamespace == namespace &&
			istioMesh.Version == install.Version {
			return mesh
		}
	}
	return nil
}

func GetMeshForInstall(install *v1.Install, meshes v1.MeshList) *v1.Mesh {
	meshInstall, ok := install.GetInstallType().(*v1.Install_Mesh)
	if !ok {
		return nil
	}

	switch meshInstallType := meshInstall.Mesh.GetMeshInstallType().(type) {
	case *v1.MeshInstall_Linkerd:
		return getLinkerdMeshForInstall(meshInstallType.Linkerd, meshes, install.InstallationNamespace)
	case *v1.MeshInstall_Istio:
		return getIstioMeshForInstall(meshInstallType.Istio, meshes, install.InstallationNamespace)
	default:
		return nil
	}
}
