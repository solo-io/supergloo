package gloo

import (
	"context"

	"github.com/solo-io/supergloo/pkg/install/utils/kubeinstall/mocks"
	"github.com/solo-io/supergloo/pkg/util"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/test/inputs"
)

var _ = Describe("gloo installer", func() {
	type testCase struct {
		installNs       string
		version         string
		disabled        bool
		existingInstall *v1.MeshIngress
		glooPrefs       *v1.GlooInstall
		targetMeshes    []*core.ResourceRef
	}
	testInputs := func(c testCase) (*v1.Install, *v1.MeshIngress, *v1.GlooInstall) {
		install := inputs.GlooIstallWithMeshes("test", "mesh", c.installNs, c.version, c.disabled, c.targetMeshes)
		if c.glooPrefs != nil {
			c.glooPrefs.GlooVersion = c.version
			install.GetIngress().IngressInstallType = &v1.MeshIngressInstall_Gloo{Gloo: c.glooPrefs}
		}
		if c.existingInstall != nil {
			ref := c.existingInstall.Metadata.Ref()
			install.GetIngress().InstalledIngress = &ref
		}
		return install, c.existingInstall, install.GetIngress().GetGloo()
	}
	Context("invalid opts", func() {
		It("errors on missing meshingress", func() {
			install, _, _ := testInputs(testCase{
				installNs:       "ok",
				version:         "0.13.9",
				existingInstall: inputs.GlooMeshIngress("ok", nil),
			})
			kubeInstaller := &mocks.MockKubeInstaller{}
			installer := newGlooInstaller(kubeInstaller)
			_, err := installer.EnsureGlooInstall(context.TODO(), install, nil, nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("installed ingress not found"))
		})
	})
	Context("invalid opts", func() {
		It("errors on missing target meshes", func() {
			install, meshIngress, _ := testInputs(testCase{
				installNs:       "ok",
				version:         "0.13.9",
				existingInstall: inputs.GlooMeshIngress("ok", nil),
				targetMeshes:    []*core.ResourceRef{{"some", "mesh"}},
			})
			kubeInstaller := &mocks.MockKubeInstaller{}
			installer := newGlooInstaller(kubeInstaller)
			_, err := installer.EnsureGlooInstall(context.TODO(), install, nil, v1.MeshIngressList{meshIngress})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("target mesh not found"))
		})
	})
	Context("install disabled", func() {
		It("calls purge with the expected labels, sets installed mesh to nil", func() {
			install, meshIngress, _ := testInputs(testCase{
				installNs:       "ok",
				version:         "0.13.9",
				existingInstall: inputs.GlooMeshIngress("ok", nil),
				targetMeshes:    []*core.ResourceRef{{"some", "mesh"}},
				disabled:        true,
			})
			kubeInstaller := &mocks.MockKubeInstaller{}
			installer := newGlooInstaller(kubeInstaller)
			updatedMeshIngress, err := installer.EnsureGlooInstall(context.TODO(), install, nil, v1.MeshIngressList{meshIngress})
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedMeshIngress).To(BeNil())
			Expect(install.GetIngress().InstalledIngress).To(BeNil())
			Expect(kubeInstaller.PurgeCalledWith).To(Equal(mocks.PurgeParams{
				InstallLabels: util.LabelsForResource(install),
			}))
		})
	})
	Context("install enabled, no preexisting install", func() {
		It("calls reconcile resources with the expected resources and labels, sets installed mesh", func() {
			mesh := inputs.IstioMesh("some-ns", nil)
			meshRef := mesh.Metadata.Ref()
			targetMeshes := []*core.ResourceRef{&meshRef}
			install, meshIngress, gloo := testInputs(testCase{
				installNs:    "ok",
				version:      "0.13.9",
				targetMeshes: targetMeshes,
			})
			kubeInstaller := &mocks.MockKubeInstaller{}
			installer := newGlooInstaller(kubeInstaller)
			updatedMeshIngress, err := installer.EnsureGlooInstall(context.TODO(), install, v1.MeshList{mesh}, v1.MeshIngressList{meshIngress})
			Expect(err).NotTo(HaveOccurred())
			Expect(*updatedMeshIngress).To(Equal(v1.MeshIngress{
				Metadata:              install.Metadata,
				MeshIngressType:       &v1.MeshIngress_Gloo{Gloo: &v1.GlooMeshIngress{}},
				InstallationNamespace: "ok",
				Meshes:                targetMeshes,
			}))
			Expect(*install.GetIngress().InstalledIngress).To(Equal(updatedMeshIngress.Metadata.Ref()))

			manifests, err := makeManifestsForInstall(context.TODO(), install, gloo)
			Expect(err).NotTo(HaveOccurred())

			resources, err := manifests.ResourceList()
			Expect(err).NotTo(HaveOccurred())

			Expect(kubeInstaller.ReconcileCalledWith).To(Equal(mocks.ReconcileParams{
				InstallNamespace: "ok",
				Resources:        resources,
				InstallLabels:    util.LabelsForResource(install),
			}))
		})
	})
	Context("install enabled with preexisting install", func() {
		It("calls reconcile resources with the expected resources and labels, updates installed mesh", func() {
			mesh := inputs.IstioMesh("some-ns", nil)
			meshRef := mesh.Metadata.Ref()
			targetMeshes := []*core.ResourceRef{&meshRef}
			originalIngress := inputs.GlooMeshIngress("ok", nil)
			install, _, gloo := testInputs(testCase{
				installNs:       "ok",
				version:         "0.13.9",
				existingInstall: originalIngress,
				targetMeshes:    targetMeshes,
			})
			kubeInstaller := &mocks.MockKubeInstaller{}
			installer := newGlooInstaller(kubeInstaller)
			updatedMeshIngress, err := installer.EnsureGlooInstall(context.TODO(), install, v1.MeshList{mesh}, v1.MeshIngressList{originalIngress})
			Expect(err).NotTo(HaveOccurred())
			Expect(*updatedMeshIngress).To(Equal(v1.MeshIngress{
				Metadata:              originalIngress.Metadata,
				MeshIngressType:       &v1.MeshIngress_Gloo{Gloo: &v1.GlooMeshIngress{}},
				InstallationNamespace: "ok",
				Meshes:                targetMeshes,
			}))
			Expect(*install.GetIngress().InstalledIngress).To(Equal(updatedMeshIngress.Metadata.Ref()))

			manifests, err := makeManifestsForInstall(context.TODO(), install, gloo)
			Expect(err).NotTo(HaveOccurred())

			resources, err := manifests.ResourceList()
			Expect(err).NotTo(HaveOccurred())

			Expect(kubeInstaller.ReconcileCalledWith).To(Equal(mocks.ReconcileParams{
				InstallNamespace: "ok",
				Resources:        resources,
				InstallLabels:    util.LabelsForResource(install),
			}))
		})
	})
})
