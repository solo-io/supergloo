package utils

import (
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"k8s.io/apimachinery/pkg/labels"
)

type InstallFilterFunc func(install *v1.Install) bool

var IstioInstallFilterFunc InstallFilterFunc = func(install *v1.Install) bool {
	meshInstall := install.GetMesh()
	if meshInstall == nil {
		return false
	}
	if istioMeshInstall := meshInstall.GetIstio(); istioMeshInstall != nil {
		return true
	}
	return false
}

var LinkerdInstallFilterFunc InstallFilterFunc = func(install *v1.Install) bool {
	meshInstall := install.GetMesh()
	if meshInstall == nil {
		return false
	}
	if linkerdMeshInstall := meshInstall.GetLinkerd(); linkerdMeshInstall != nil {
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
	if linkerdMesh := mesh.GetLinkerd(); linkerdMesh != nil {
		return true
	}
	return false
}

var AppmeshFilterFunc MeshFilterFunc = func(mesh *v1.Mesh) bool {
	if appmeshMesh := mesh.GetAwsAppMesh(); appmeshMesh != nil {
		return true
	}
	return false
}

func FilterByLabels(meshLabels map[string]string) MeshFilterFunc {
	return func(mesh *v1.Mesh) bool {
		return labels.SelectorFromSet(meshLabels).Matches(labels.Set(mesh.Metadata.GetLabels()))
	}
}

func GetActiveInstalls(installs v1.InstallList, filterFuncs ...InstallFilterFunc) v1.InstallList {
	var result v1.InstallList
	for _, install := range installs {
		if install.Disabled {
			continue
		}
		for _, filterFunc := range filterFuncs {
			if !filterFunc(install) {
				continue
			}
		}
		result = append(result, install)
	}
	return result
}

func GetMeshes(meshes v1.MeshList, filterFuncs ...MeshFilterFunc) v1.MeshList {
	var result v1.MeshList
	for _, mesh := range meshes {
		validMesh := true
		for _, filterFunc := range filterFuncs {
			if !filterFunc(mesh) {
				validMesh = false
				break
			}
		}
		if validMesh {
			result = append(result, mesh)
		}
	}
	return result
}
