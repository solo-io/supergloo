package util

import (
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

func getLinkerdMeshForInstall(install *v1.LinkerdInstall, meshes v1.MeshList, namespace string) *v1.Mesh {
	for _, mesh := range meshes {
		linkerdMesh := mesh.GetLinkerdMesh()
		if linkerdMesh == nil {
			continue
		}

		if linkerdMesh.InstallationNamespace == namespace &&
			linkerdMesh.LinkerdVersion == install.LinkerdVersion {
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
			istioMesh.IstioVersion == install.IstioVersion {
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
	case *v1.MeshInstall_LinkerdMesh:
		return getLinkerdMeshForInstall(meshInstallType.LinkerdMesh, meshes, install.InstallationNamespace)
	case *v1.MeshInstall_IstioMesh:
		return getIstioMeshForInstall(meshInstallType.IstioMesh, meshes, install.InstallationNamespace)
	default:
		return nil
	}
}

type NamespaceListFilterFunc = func(namespace *v1.KubeNamespace) bool

func FilterNamespaces(namespaces v1.KubeNamespaceList, filterFunc NamespaceListFilterFunc) v1.KubeNamespaceList {
	var result v1.KubeNamespaceList
	for _, namespace := range namespaces {
		if filterFunc(namespace) {
			result = append(result, namespace)
		}
	}
	return result
}

type InjetedPodFilterFunc = func(pod *v1.Pod) bool

// TODO(EItanya): figure out a heuristic for when a singular pod has been injected
func GetInjectedPods(namespaces v1.KubeNamespaceList, pods v1.PodList,
	filterFunc InjetedPodFilterFunc) v1.PodsByNamespace {
	result := make(v1.PodsByNamespace)
	for _, namespace := range namespaces {
		result[namespace.Name] = v1.PodList{}
		for _, pod := range pods {
			if pod.Namespace == namespace.Name && filterFunc(pod) {
				result[namespace.Name] = append(result[namespace.Name], pod)
			}
		}

	}
	return result
}
