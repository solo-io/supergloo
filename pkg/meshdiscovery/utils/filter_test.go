package utils

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

var _ = Describe("istio mesh discovery unit tests", func() {
	var (
		snap               *v1.DiscoverySnapshot
		superglooNamespace = "supergloo-system"
	)

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
		filteredMeshes := GetMeshes(snap.Meshes.List(), IstioMeshFilterFunc)
		Expect(filteredMeshes).To(HaveLen(1))
		Expect(filteredMeshes[0].Metadata.Name).To(Equal("one"))
	})
})
