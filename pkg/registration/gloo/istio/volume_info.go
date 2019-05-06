package istio

import (
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	corev1 "k8s.io/api/core/v1"
)

const certSuffix = "certs"

var (
	defaultMode int32 = 420
	optional          = true
)

type VolumeList []corev1.Volume

func (s VolumeList) Remove(i int) VolumeList {
	tmp := make(VolumeList, len(s))
	copy(tmp, s)
	return append(tmp[:i], tmp[i+1:]...)
}

type VolumeMountList []corev1.VolumeMount

func (s VolumeMountList) Remove(i int) VolumeMountList {
	tmp := make(VolumeMountList, len(s))
	copy(tmp, s)
	return append(tmp[:i], tmp[i+1:]...)
}

type DeploymentVolumeInfoList []DeploymentVolumeInfo

type DeploymentVolumeInfo struct {
	Volume      corev1.Volume
	VolumeMount corev1.VolumeMount
}

func NewDeploymentVolumeInfo(mesh *v1.Mesh, tlsSecretName string) *DeploymentVolumeInfo {
	certVolumeName := certVolumeName(mesh.Metadata.Ref())
	volume := corev1.Volume{
		Name: certVolumeName,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				Optional:    &optional,
				DefaultMode: &defaultMode,
				SecretName:  tlsSecretName,
			},
		},
	}
	volumeMount := corev1.VolumeMount{
		Name:      certVolumeName,
		ReadOnly:  true,
		MountPath: certVolumePathName(mesh.Metadata.Ref()),
	}
	return &DeploymentVolumeInfo{
		VolumeMount: volumeMount,
		Volume:      volume,
	}
}

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
			deleted = append(deleted, old)
		}
	}
	return added, deleted
}

func volumesToDeploymentInfo(volumes VolumeList, mounts VolumeMountList) DeploymentVolumeInfoList {
	var result DeploymentVolumeInfoList
	for _, volume := range volumes {
		if strings.Contains(volume.Name, certSuffix) {
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
	return result
}

func makeSecretVolumesForMeshes(meshes v1.MeshList) (DeploymentVolumeInfoList, error) {
	result := make(DeploymentVolumeInfoList, len(meshes))
	for i, mesh := range meshes {
		var deploymentVolumeInfo *DeploymentVolumeInfo
		switch mesh.MeshType.(type) {
		case *v1.Mesh_Istio:
			if mesh.MtlsConfig != nil && mesh.MtlsConfig.MtlsEnabled {
				deploymentVolumeInfo = NewDeploymentVolumeInfo(mesh, "istio.default")
			}
		default:
			return nil, errors.Errorf("unsupported mesh type found for mesh ingress "+
				"target mesh, %s.%s", mesh.Metadata.Namespace, mesh.Metadata.Name)
		}

		if deploymentVolumeInfo != nil {
			result[i] = *deploymentVolumeInfo
		}

	}
	return result, nil
}

func certVolumeName(mesh core.ResourceRef) string {
	return strings.Join([]string{mesh.Name, certSuffix}, "-")
}

func certVolumePathName(mesh core.ResourceRef) string {
	return filepath.Join("/etc", "certs", mesh.Namespace, mesh.Name)
}
