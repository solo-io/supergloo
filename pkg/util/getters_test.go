package util_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/install/istio"
	"github.com/solo-io/supergloo/pkg/util"
	"github.com/solo-io/supergloo/test/inputs"
)

var _ = Describe("getters unit tests", func() {
	var (
		install *v1.Install
		meshes  v1.MeshList

		namespace = "supergloo-system"
	)
	Context("istio", func() {
		It("can find the mesh for a given install", func() {
			meshes = v1.MeshList{
				inputs.IstioMeshWithVersion(namespace, istio.IstioVersion106, &core.ResourceRef{}),
				inputs.LinkerdMeshWithVersion(namespace, inputs.LinkerdTestInstallNs, inputs.LinkerdTestVersion),
			}
			install = inputs.IstioInstall("one", namespace, inputs.IstioTestInstallNs, inputs.IstioTestVersion, false)
			foundMesh := util.GetMeshForInstall(install, meshes)
			Expect(foundMesh).NotTo(BeNil())
			Expect(foundMesh).To(BeEquivalentTo(meshes[0]))
		})
		It("returns nil if none are found", func() {
			meshes = v1.MeshList{}
			install = inputs.IstioInstall("one", namespace, inputs.IstioTestInstallNs, inputs.IstioTestVersion, false)
			foundMesh := util.GetMeshForInstall(install, meshes)
			Expect(foundMesh).To(BeNil())
		})
	})
	Context("linkerd", func() {
		It("can find the mesh for a given install", func() {
			meshes = v1.MeshList{
				inputs.IstioMeshWithVersion(namespace, istio.IstioVersion106, &core.ResourceRef{}),
				inputs.LinkerdMeshWithVersion(namespace, inputs.LinkerdTestInstallNs, inputs.LinkerdTestVersion),
			}
			install = inputs.LinkerdInstall("one", namespace, inputs.LinkerdTestInstallNs, inputs.LinkerdTestVersion, false)
			foundMesh := util.GetMeshForInstall(install, meshes)
			Expect(foundMesh).NotTo(BeNil())
			Expect(foundMesh).To(BeEquivalentTo(meshes[1]))
		})
		It("returns nil if none are found", func() {
			meshes = v1.MeshList{}
			install = inputs.LinkerdInstall("one", namespace, inputs.LinkerdTestInstallNs, inputs.LinkerdTestVersion, false)
			foundMesh := util.GetMeshForInstall(install, meshes)
			Expect(foundMesh).To(BeNil())
		})
	})
})
