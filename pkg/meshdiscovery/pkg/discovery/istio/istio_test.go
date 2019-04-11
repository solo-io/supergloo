package istio

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"
	"github.com/solo-io/supergloo/pkg/api/custom/clients/kubernetes"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	kubev1 "k8s.io/api/core/v1"
)

var _ = Describe("istio mesh discovery unit tests", func() {
	var (
		snap *v1.DiscoverySnapshot
	)

	var constructPod = func(container kubev1.Container) *v1.Pod {

		pod := &kubev1.Pod{
			Spec: kubev1.PodSpec{
				Containers: []kubev1.Container{
					container,
				},
			},
		}
		return kubernetes.FromKube(pod)
	}

	BeforeEach(func() {
		snap = &v1.DiscoverySnapshot{
			Meshes: v1.MeshesByNamespace{
				"one": v1.MeshList{
					{
						MeshType: &v1.Mesh_Istio{
							Istio: &v1.IstioMesh{},
						},
					},
					{
						MeshType: &v1.Mesh_AwsAppMesh{
							AwsAppMesh: &v1.AwsAppMesh{},
						},
					},
				},
			},
		}

		clients.UseMemoryClients()
	})
	It("can properly filter istio meshes", func() {
		Expect(filterIstioMeshes(snap.Meshes.List())).To(HaveLen(1))
	})
	Context("get version from pod", func() {
		It("errors when no pilot container is found", func() {
			container := kubev1.Container{
				Image: "istio-",
			}
			pod := constructPod(container)
			_, err := getVersionFromPod(pod)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unable to find pilot container from pod"))
		})
		It("errors when no version is found in image name", func() {
			container := kubev1.Container{
				Image: "istio-pilot",
			}
			pod := constructPod(container)
			_, err := getVersionFromPod(pod)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unable to find image version for image"))
		})
	})
})
