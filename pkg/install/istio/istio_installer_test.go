package istio

import (
	"context"

	"github.com/solo-io/go-utils/installutils/kubeinstall/mocks"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/pkg/util"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/test/inputs"
)

var _ = Describe("istio installer", func() {
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
		return install, c.existingInstall, install.GetMesh().GetIstioMesh()
	}
	Context("install disabled", func() {
		It("calls purge with the expected labels, sets installed mesh to nil", func() {
			install, mesh, _ := testInputs(testCase{
				installNs:       "ok",
				version:         IstioVersion106,
				disabled:        true,
				existingInstall: inputs.IstioMesh("ok", &core.ResourceRef{"some", "secret"}),
			})
			kubeInstaller := &mocks.MockKubeInstaller{}
			installer := newIstioInstaller(kubeInstaller)
			err := installer.EnsureIstioInstall(context.TODO(), install, v1.MeshList{mesh})
			Expect(err).NotTo(HaveOccurred())
			Expect(kubeInstaller.PurgeCalledWith).To(Equal(mocks.PurgeParams{
				InstallLabels: util.LabelsForResource(install),
			}))
		})
	})
	Context("install enabled, no preexisting install", func() {
		It("calls reconcile resources with the expected resources and labels, sets installed mesh", func() {
			install, mesh, istio := testInputs(testCase{
				installNs: "ok",
				version:   IstioVersion106,
			})
			kubeInstaller := &mocks.MockKubeInstaller{}
			installer := newIstioInstaller(kubeInstaller)
			err := installer.EnsureIstioInstall(context.TODO(), install, v1.MeshList{mesh})
			Expect(err).NotTo(HaveOccurred())

			manifests, err := makeManifestsForInstall(context.TODO(), install, mesh, istio)
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
			originalMesh := inputs.IstioMeshWithVersion("ok", IstioVersion106,
				&core.ResourceRef{"some", "seret"})
			install, mesh, istio := testInputs(testCase{
				installNs:       "istio-was-installed-herr",
				version:         IstioVersion106,
				existingInstall: originalMesh,
			})
			kubeInstaller := &mocks.MockKubeInstaller{}
			installer := newIstioInstaller(kubeInstaller)
			err := installer.EnsureIstioInstall(context.TODO(), install, v1.MeshList{mesh})
			Expect(err).NotTo(HaveOccurred())

			manifests, err := makeManifestsForInstall(context.TODO(), install, mesh, istio)
			Expect(err).NotTo(HaveOccurred())

			resources, err := manifests.ResourceList()
			Expect(err).NotTo(HaveOccurred())

			Expect(kubeInstaller.ReconcileCalledWith).To(Equal(mocks.ReconcileParams{
				InstallNamespace: "istio-was-installed-herr",
				Resources:        resources,
				InstallLabels:    util.LabelsForResource(install),
			}))
		})
	})
})
