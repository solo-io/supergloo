package meshingress_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/install/meshingress"
	"github.com/solo-io/supergloo/pkg/install/utils/helm"
)

var _ = Describe("mesh ingress mock installer", func() {
	var createdManifests, deletedManifests, updatedManifests helm.Manifests
	BeforeEach(func() {
		createdManifests, deletedManifests, updatedManifests = nil, nil, nil
	})
	installer := meshingress.NewDefaultInstaller(helm.NewMockHelm(
		func(ctx context.Context, namespace string, manifests helm.Manifests) error {
			createdManifests = manifests
			return nil
		}, func(ctx context.Context, namespace string, manifests helm.Manifests) error {
			deletedManifests = manifests
			return nil
		}, func(ctx context.Context, namespace string, original, updated helm.Manifests, recreatePods bool) error {
			updatedManifests = updated
			return nil
		}))
	ns := "ns"
	It("installs, upgrades, and uninstalls from an install object", func() {

		glooConfig := &v1.MeshIngressInstall_Gloo{
			Gloo: &v1.GlooInstall{
				GlooVersion: "0.11.1",
			},
		}
		installConfig := &v1.Install_Ingress{
			Ingress: &v1.MeshIngressInstall{
				InstallType: glooConfig,
			},
		}

		install := &v1.Install{
			Metadata:              core.Metadata{Name: "myinstall", Namespace: "myns"},
			Disabled:              false,
			InstallationNamespace: ns,
			InstallType:           installConfig,
		}

		installedIngress, err := installer.EnsureIngressInstall(context.TODO(), install)
		Expect(err).NotTo(HaveOccurred())
		Expect(installedIngress.Metadata.Name).To(Equal(install.Metadata.Name))

		// installed manifest should be set
		Expect(install.InstalledManifest).NotTo(HaveLen(0))
		installedManifests, err := helm.NewManifestsFromGzippedString(install.InstalledManifest)
		Expect(err).NotTo(HaveOccurred())
		Expect(installedManifests).To(Equal(createdManifests))

		Expect(install.InstallType).To(BeAssignableToTypeOf(&v1.Install_Ingress{}))
		mesh := install.InstallType.(*v1.Install_Ingress)
		// should be set by install
		Expect(mesh.Ingress.InstalledIngress).NotTo(BeNil())
		Expect(*mesh.Ingress.InstalledIngress).To(Equal(installedIngress.Metadata.Ref()))

		Expect(installedIngress.Metadata.Name).To(Equal(install.Metadata.Name))

		// uninstall should work
		install.Disabled = true
		installedIngress, err = installer.EnsureIngressInstall(context.TODO(), install)
		Expect(err).NotTo(HaveOccurred())
		Expect(installedIngress).To(BeNil())
		Expect(install.InstalledManifest).To(HaveLen(0))

		Expect(deletedManifests).To(Equal(createdManifests))
		Expect(updatedManifests).To(HaveLen(0))
	})

})
