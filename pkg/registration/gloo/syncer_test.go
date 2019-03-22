package gloo

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
)

var _ = Describe("gloo config syncers", func() {
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
	var createDeployment = func(volumes VolumeList, mounts VolumeMountList) *v1beta1.Deployment {
		return &v1beta1.Deployment{
			Spec: v1beta1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Volumes: volumes,
						Containers: []corev1.Container{
							{
								VolumeMounts: mounts,
							},
						},
					},
				},
			},
		}
	}
	Context("should update", func() {
		It("Should not update if nothing changed", func() {
			deployment := createDeployment(volumeList, mountList)
			update, err := ShouldUpdateDeployment(deployment, []*core.ResourceRef{meshResource}, v1.MeshList{istioMesh})
			Expect(err).NotTo(HaveOccurred())
			Expect(update).To(BeFalse())
			Expect(deployment).To(Equal(createDeployment(volumeList, mountList)))
		})

		It("should update if one is removed", func() {
			deployment := createDeployment(volumeList.remove(0), mountList.remove(0))
			update, err := ShouldUpdateDeployment(deployment, []*core.ResourceRef{meshResource}, v1.MeshList{istioMesh})
			Expect(err).NotTo(HaveOccurred())
			Expect(update).To(BeTrue())
			Expect(deployment).NotTo(Equal(createDeployment(volumeList, mountList)))
		})

		It("should update if on is added", func() {
			deployment := createDeployment(volumeList, mountList)
			update, err := ShouldUpdateDeployment(deployment, []*core.ResourceRef{}, v1.MeshList{})
			Expect(err).NotTo(HaveOccurred())
			Expect(update).To(BeTrue())
			Expect(deployment).NotTo(Equal(createDeployment(volumeList, mountList)))
		})
	})

	Context("")

})
