package istio

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"
	"github.com/solo-io/supergloo/pkg/api/custom/clients/kubernetes"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	kubev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("istio mesh discovery unit tests", func() {
	var (
		snap *v1.DiscoverySnapshot

		istioNamespace     = "istio-system"
		superglooNamespace = "supergloo-system"
	)

	var constructPod = func(container kubev1.Container) *v1.Pod {

		pod := &kubev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: istioNamespace,
			},
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
				superglooNamespace: v1.MeshList{
					{
						Metadata: core.Metadata{Name: "one"},
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
		filteredMeshes := filterIstioMeshes(snap.Meshes.List())
		Expect(filteredMeshes).To(HaveLen(1))
		Expect(filteredMeshes[0].Metadata.Name).To(Equal("one"))
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
		It("fails when image is the incorrect format", func() {
			container := kubev1.Container{
				Image: "istio-pilot:10.6",
			}
			pod := constructPod(container)
			_, err := getVersionFromPod(pod)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unable to find image version for image"))
		})
		It("errors when no version is found in image name", func() {
			container := kubev1.Container{
				Image: "istio-pilot:1.0.6",
			}
			pod := constructPod(container)
			version, err := getVersionFromPod(pod)
			Expect(err).NotTo(HaveOccurred())
			Expect(version).To(Equal("1.0.6"))
		})
	})
	Context("discovery data", func() {
		It("can properly construct the discovery data", func() {
			container := kubev1.Container{
				Image: "istio-pilot:1.0.6",
			}
			pod := constructPod(container)
			mesh, err := constructDiscoveryData(pod)
			Expect(err).NotTo(HaveOccurred())
			Expect(mesh.DiscoveryMetadata).To(BeEquivalentTo(&v1.DiscoveryMetadata{
				InstallationNamespace: istioNamespace,
				MeshVersion:           "1.0.6",
			}))
		})
	})
})
