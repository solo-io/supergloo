package gloo

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("volume info", func() {
	var (
		meshResource = &core.ResourceRef{
			Name:      "istio",
			Namespace: "istio-system",
		}
		istioMesh = &v1.Mesh{
			MeshType: &v1.Mesh_Istio{
				Istio: &v1.IstioMesh{
					InstallationNamespace: "istio-system",
				},
			},
			Metadata: core.Metadata{
				Name:      "istio",
				Namespace: "istio-system",
			},
		}
		volumeList = VolumeList{
			corev1.Volume{
				Name: certVolumeName(meshResource),
			},
			corev1.Volume{
				Name: "1",
			},
			corev1.Volume{
				Name: "2",
			},
		}
		mountList = VolumeMountList{
			corev1.VolumeMount{
				Name: certVolumeName(meshResource),
			},
			corev1.VolumeMount{
				Name: "1",
			},
			corev1.VolumeMount{
				Name: "2",
			},
		}
	)
	Context("remove from list", func() {
		It("can remove an item from a volume list", func() {
			volumes := volumeList.remove(0)
			Expect(volumes).To(HaveLen(2))
			for _, v := range volumes {
				Expect(v.Name).NotTo(Equal(certVolumeName(meshResource)))
			}
		})
		It("can remove an item from a mount list", func() {
			mounts := mountList.remove(0)
			Expect(mounts).To(HaveLen(2))
			for _, v := range mounts {
				Expect(v.Name).NotTo(Equal(certVolumeName(meshResource)))
			}
		})
	})

	Context("Deployment Info", func() {
		It("can create deployment info from volumes", func() {

			deploymentList := VolumesToDeploymentInfo(volumeList, mountList)
			Expect(deploymentList).To(HaveLen(1))
			Expect(deploymentList[0].Volume.Name).To(Equal(certVolumeName(meshResource)))
		})
		It("cam create deployment info from meshes", func() {
			deploymentList, err := ResourcesToDeploymentInfo([]*core.ResourceRef{meshResource}, v1.MeshList{istioMesh})
			Expect(err).NotTo(HaveOccurred())
			Expect(deploymentList).To(HaveLen(1))
			Expect(deploymentList[0].Volume.Name).To(Equal(certVolumeName(meshResource)))
		})

		It("can diff properly", func() {
			newDeploymentList, err := ResourcesToDeploymentInfo([]*core.ResourceRef{meshResource}, v1.MeshList{istioMesh})
			Expect(err).NotTo(HaveOccurred())
			oldDeploymentList := VolumesToDeploymentInfo(volumeList, mountList)
			added, deleted := diff(newDeploymentList, oldDeploymentList)
			Expect(added).To(HaveLen(0))
			Expect(deleted).To(HaveLen(0))
			added, deleted = diff(newDeploymentList, DeploymentVolumeInfoList{})
			Expect(added).To(HaveLen(1))
			Expect(deleted).To(HaveLen(0))
			added, deleted = diff(DeploymentVolumeInfoList{}, oldDeploymentList)
			Expect(added).To(HaveLen(0))
			Expect(deleted).To(HaveLen(1))
		})
	})
})
