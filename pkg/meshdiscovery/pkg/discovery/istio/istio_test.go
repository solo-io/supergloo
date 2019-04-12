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

	var constructPod = func(container kubev1.Container, namespace string) *v1.Pod {

		pod := &kubev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
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
			pod := constructPod(container, istioNamespace)
			_, err := getVersionFromPod(pod)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unable to find pilot container from pod"))
		})
		It("errors when no version is found in image name", func() {
			container := kubev1.Container{
				Image: "istio-pilot",
			}
			pod := constructPod(container, istioNamespace)
			_, err := getVersionFromPod(pod)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unable to find image version for image"))
		})
		It("fails when image is the incorrect format", func() {
			container := kubev1.Container{
				Image: "istio-pilot:10.6",
			}
			pod := constructPod(container, istioNamespace)
			_, err := getVersionFromPod(pod)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unable to find image version for image"))
		})
		It("errors when no version is found in image name", func() {
			container := kubev1.Container{
				Image: "istio-pilot:1.0.6",
			}
			pod := constructPod(container, istioNamespace)
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
			pod := constructPod(container, istioNamespace)
			mesh, err := constructDiscoveryData(pod)
			Expect(err).NotTo(HaveOccurred())
			Expect(mesh.DiscoveryMetadata).To(BeEquivalentTo(&v1.DiscoveryMetadata{
				InstallationNamespace: istioNamespace,
				MeshVersion:           "1.0.6",
			}))
		})
	})
	Context("merge meshes", func() {
		var getDiscoveredMeshes = func(namespace string) *v1.Mesh {
			container := kubev1.Container{
				Image: "istio-pilot:1.0.6",
			}
			pod := constructPod(container, namespace)
			mesh, err := constructDiscoveryData(pod)
			Expect(err).NotTo(HaveOccurred())
			mesh.MeshType = &v1.Mesh_Istio{Istio: &v1.IstioMesh{
				InstallationNamespace: namespace,
				IstioVersion:          "1.0.6",
			}}
			return mesh
		}
		It("can merge into an empty existing list", func() {
			discoveredMeshes := v1.MeshList{getDiscoveredMeshes(istioNamespace)}
			meshList, err := mergeMeshes(discoveredMeshes, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(meshList).To(HaveLen(1))
		})
		It("can merge into a non empty list, with a mesh in a different Namespace", func() {
			discoveredMeshes := v1.MeshList{getDiscoveredMeshes(istioNamespace)}
			existingMeshes := v1.MeshList{getDiscoveredMeshes("one")}
			meshList, err := mergeMeshes(discoveredMeshes, existingMeshes)
			Expect(err).NotTo(HaveOccurred())
			Expect(meshList).To(HaveLen(2))
		})
		It("can merge into a non empty list with an existing mesh", func() {
			discoverdMeshes := v1.MeshList{getDiscoveredMeshes(istioNamespace)}
			existingMesh := getDiscoveredMeshes(istioNamespace)
			existingMesh.DiscoveryMetadata = nil
			existingMeshes := v1.MeshList{existingMesh}
			meshList, err := mergeMeshes(discoverdMeshes, existingMeshes)
			Expect(err).NotTo(HaveOccurred())
			Expect(meshList).To(HaveLen(1))
			Expect(&meshList[0]).To(Equal(&existingMeshes[0]))
		})

	})
})
