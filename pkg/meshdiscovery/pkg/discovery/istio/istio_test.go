package istio

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

var _ = Describe("istio mesh discovery unit tests", func() {
	var (
		snap *v1.DiscoverySnapshot
	)

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
	})
	It("can properly filter istio meshes", func() {
		Expect(filterIstioMeshes(snap.Meshes.List())).To(HaveLen(1))
	})
})
