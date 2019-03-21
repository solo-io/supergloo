package gloo

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	corev1 "k8s.io/api/core/v1"
)

type DeploymentVolumeInfoList []DeploymentVolumeInfo

func (list DeploymentVolumeInfoList) containsVolume(volume corev1.Volume) bool {
	for _, v := range list {
		if v.Volume.Name == volume.Name {
			return true
		}
	}
	return false
}

func diff(newList DeploymentVolumeInfoList, oldList DeploymentVolumeInfoList) (added DeploymentVolumeInfoList, deleted DeploymentVolumeInfoList) {
	for _, new := range newList {
		found := false
		for _, old := range oldList {
			if new.Volume.Name == old.Volume.Name {
				found = true
			}
		}
		if !found {
			added = append(added, new)
		}
	}

	for _, old := range oldList {
		found := false
		for _, new := range newList {
			if old.Volume.Name == new.Volume.Name {
				found = true
			}
		}
		if !found {
			deleted = append(deleted)
		}
	}
	return added, deleted
}

type DeploymentVolumeInfo struct {
	Volume      corev1.Volume
	VolumeMount corev1.VolumeMount
}

func VolumesToDeploymentInfo(volumes []corev1.Volume, mounts []corev1.VolumeMount) DeploymentVolumeInfoList {
	var result DeploymentVolumeInfoList
	for _, volume := range volumes {
		if strings.Contains(volume.Name, "-certs") {
			for _, mount := range mounts {
				if mount.Name == volume.Name {
					result = append(result, DeploymentVolumeInfo{
						VolumeMount: mount,
						Volume:      volume,
					})
				}
			}
		}
	}
}

func ResourcesToDeploymentInfo(resources []*core.ResourceRef, meshes v1.MeshList) (DeploymentVolumeInfoList, error) {
	result := make(DeploymentVolumeInfoList, len(resources))
	for _, resource := range resources {
		mesh, err := meshes.Find(resource.Namespace, resource.Name)
		if err != nil {
			return nil, err
		}
		var tlsSecretName string
		switch mesh.MeshType.(type) {
		case *v1.Mesh_Istio:
			tlsSecretName = "istio.defaut"
		default:
			return nil, errors.Errorf("unsupported mesh type found for mesh ingress "+
				"target mesh, %s.%s", resource.Namespace, resource.Name)
		}
		volume := corev1.Volume{
			Name: certVolumeName(resource),
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					Optional:    &optional,
					DefaultMode: &defaultMode,
					SecretName:  tlsSecretName,
				},
			},
		}
		volumeMount := corev1.VolumeMount{
			Name:      tlsSecretName,
			ReadOnly:  true,
			MountPath: certVolumePathName(resource),
		}
		result = append(result, DeploymentVolumeInfo{
			VolumeMount: volumeMount,
			Volume:      volume,
		})
	}
	return result, nil
}

func certVolumeName(mesh *core.ResourceRef) string {
	return fmt.Sprintf("%s-%s-certs", mesh.Name, mesh.Namespace)
}

func certVolumePathName(mesh *core.ResourceRef) string {
	return filepath.Join("/etc", "certs", "namespace", "name")
}
