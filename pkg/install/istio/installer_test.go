package istio

import (
	"context"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/pkg/util"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/install/utils/kuberesource"
	"github.com/solo-io/supergloo/test/inputs"
)

type mockKubeInstaller struct {
	reconcileCalledWith reconcileParams
	purgeCalledWith     purgeParams
	returnErr           error
}

type reconcileParams struct {
	installNamespace string
	resources        kuberesource.UnstructuredResources
	installLabels    map[string]string
}

type purgeParams struct {
	installLabels map[string]string
}

func (i *mockKubeInstaller) ReconcilleResources(ctx context.Context, installNamespace string, resources kuberesource.UnstructuredResources, installLabels map[string]string) error {
	i.reconcileCalledWith = reconcileParams{installNamespace, resources, installLabels}
	return i.returnErr
}

func (i *mockKubeInstaller) PurgeResources(ctx context.Context, withLabels map[string]string) error {
	i.purgeCalledWith = purgeParams{withLabels}
	return i.returnErr
}

var _ = Describe("makeManifestsForInstall", func() {
	type testCase struct {
		installNs       string
		version         string
		disabled        bool
		existingInstall *v1.Mesh
		istioPrefs      *v1.IstioInstall
	}
	testInputs := func(c testCase) (*v1.Install, *v1.Mesh, *v1.IstioInstall) {
		install := inputs.IstioInstall("test", "mesh", c.installNs, c.version, c.disabled)
		if c.istioPrefs != nil {
			c.istioPrefs.IstioVersion = c.version
			install.GetMesh().MeshInstallType = &v1.MeshInstall_IstioMesh{IstioMesh: c.istioPrefs}
		}
		if c.existingInstall != nil {
			ref := c.existingInstall.Metadata.Ref()
			install.GetMesh().InstalledMesh = &ref
		}
		return install, c.existingInstall, install.GetMesh().GetIstioMesh()
	}
	Context("invalid opts", func() {
		It("errors on missing mesh", func() {
			install, _, _ := testInputs(testCase{
				installNs:       "ok",
				version:         IstioVersion106,
				existingInstall: inputs.IstioMesh("ok", &core.ResourceRef{"some", "secret"}),
			})
			kubeInstaller := &mockKubeInstaller{}
			installer := newIstioInstaller(kubeInstaller)
			_, err := installer.EnsureIstioInstall(context.TODO(), install, nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("installed mesh not found"))
		})
	})
	Context("install disabled", func() {
		It("calls purge with the expected labels, sets installed mesh to nil", func() {
			install, mesh, _ := testInputs(testCase{
				installNs:       "ok",
				version:         IstioVersion106,
				disabled:        true,
				existingInstall: inputs.IstioMesh("ok", &core.ResourceRef{"some", "secret"}),
			})
			kubeInstaller := &mockKubeInstaller{}
			installer := newIstioInstaller(kubeInstaller)
			mesh, err := installer.EnsureIstioInstall(context.TODO(), install, v1.MeshList{mesh})
			Expect(err).NotTo(HaveOccurred())
			Expect(mesh).To(BeNil())
			Expect(install.GetMesh().InstalledMesh).To(BeNil())
			Expect(kubeInstaller.purgeCalledWith).To(Equal(purgeParams{
				installLabels: util.LabelsForResource(install),
			}))
		})
	})
	Context("install enabled, no preexisting install", func() {
		It("calls reconcile resources with the expected resources and labels, sets installed mesh", func() {
			install, mesh, istio := testInputs(testCase{
				installNs: "ok",
				version:   IstioVersion106,
			})
			kubeInstaller := &mockKubeInstaller{}
			installer := newIstioInstaller(kubeInstaller)
			installedMesh, err := installer.EnsureIstioInstall(context.TODO(), install, v1.MeshList{mesh})
			Expect(err).NotTo(HaveOccurred())
			Expect(*installedMesh).To(Equal(v1.Mesh{
				Metadata:   install.Metadata,
				MeshType:   &v1.Mesh_Istio{Istio: &v1.IstioMesh{InstallationNamespace: "ok"}},
				MtlsConfig: &v1.MtlsConfig{},
			}))
			Expect(*install.GetMesh().InstalledMesh).To(Equal(installedMesh.Metadata.Ref()))

			manifests, err := makeManifestsForInstall(context.TODO(), install, mesh, istio)
			Expect(err).NotTo(HaveOccurred())

			resources, err := manifests.ResourceList()
			Expect(err).NotTo(HaveOccurred())

			Expect(kubeInstaller.reconcileCalledWith).To(Equal(reconcileParams{
				installNamespace: "ok",
				resources:        resources,
				installLabels:    util.LabelsForResource(install),
			}))
		})
	})
	Context("install enabled with preexisting install", func() {
		It("calls reconcile resources with the expected resources and labels, updates installed mesh", func() {
			originalMesh := inputs.IstioMesh("ok", &core.ResourceRef{"some", "seret"})
			install, mesh, istio := testInputs(testCase{
				installNs:       "ok",
				version:         IstioVersion106,
				existingInstall: originalMesh,
			})
			kubeInstaller := &mockKubeInstaller{}
			installer := newIstioInstaller(kubeInstaller)
			installedMesh, err := installer.EnsureIstioInstall(context.TODO(), install, v1.MeshList{mesh})
			Expect(err).NotTo(HaveOccurred())
			Expect(*installedMesh).To(Equal(v1.Mesh{
				Metadata:   originalMesh.Metadata,
				MeshType:   &v1.Mesh_Istio{Istio: &v1.IstioMesh{InstallationNamespace: "ok"}},
				MtlsConfig: mesh.GetMtlsConfig(),
			}))
			Expect(*install.GetMesh().InstalledMesh).To(Equal(installedMesh.Metadata.Ref()))

			manifests, err := makeManifestsForInstall(context.TODO(), install, mesh, istio)
			Expect(err).NotTo(HaveOccurred())

			resources, err := manifests.ResourceList()
			Expect(err).NotTo(HaveOccurred())

			Expect(kubeInstaller.reconcileCalledWith).To(Equal(reconcileParams{
				installNamespace: "ok",
				resources:        resources,
				installLabels:    util.LabelsForResource(install),
			}))
		})
	})
})
