package utils

import (
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

type InstallFilterFunc func(install *v1.Install) bool

var IstioInstallFilterFunc InstallFilterFunc = func(install *v1.Install) bool {
	meshInstall := install.GetMesh()
	if meshInstall == nil {
		return false
	}
	if istioMeshInstall := meshInstall.GetIstioMesh(); istioMeshInstall != nil {
		return true
	}
	return false
}

var LinkerdInstallFilterFunc InstallFilterFunc = func(install *v1.Install) bool {
	meshInstall := install.GetMesh()
	if meshInstall == nil {
		return false
	}
	if linkerdMeshInstall := meshInstall.GetLinkerdMesh(); linkerdMeshInstall != nil {
		return true
	}
	return false
}

type MeshFilterFunc func(mesh *v1.Mesh) bool

var IstioMeshFilterFunc MeshFilterFunc = func(mesh *v1.Mesh) bool {
	if istioMesh := mesh.GetIstio(); istioMesh != nil {
		return true
	}
	return false
}

var LinkerdMeshFilterFunc MeshFilterFunc = func(mesh *v1.Mesh) bool {
	if linkerdMesh := mesh.GetLinkerdMesh(); linkerdMesh != nil {
		return true
	}
	return false
}

func GetInstalls(installs v1.InstallList, filterFunc InstallFilterFunc) v1.InstallList {
	var result v1.InstallList
	for _, install := range installs {
		if filterFunc(install) {
			result = append(result, install)
		}
	}
	return result
}

func GetMeshes(meshes v1.MeshList, filterFunc MeshFilterFunc) v1.MeshList {
	var result v1.MeshList
	for _, mesh := range meshes {
		if filterFunc(mesh) {
			result = append(result, mesh)
		}
	}
	return result
}
