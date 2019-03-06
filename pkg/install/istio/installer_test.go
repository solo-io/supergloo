package istio

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/install/utils/helm"
)

var _ = Describe("Installer", func() {
	var createdManifests, deletedManifests, updatedManifests helm.Manifests
	BeforeEach(func() {
		createdManifests, deletedManifests, updatedManifests = nil, nil, nil
	})
	installer := defaultIstioInstaller{helmInstaller: newMockHelm(
		func(ctx context.Context, namespace string, manifests helm.Manifests) error {
			createdManifests = manifests
			return nil
		}, func(ctx context.Context, namespace string, manifests helm.Manifests) error {
			deletedManifests = manifests
			return nil
		}, func(ctx context.Context, namespace string, original, updated helm.Manifests, recreatePods bool) error {
			updatedManifests = updated
			return nil
		})}
	ns := "ns"
	It("installs, upgrades, and uninstalls from an install object", func() {
		installConfig := &v1.Install_Istio_{
			Istio: &v1.Install_Istio{
				InstallationNamespace: ns,
				IstioVersion:          IstioVersion106,
			},
		}

		install := &v1.Install{
			Metadata:    core.Metadata{Name: "myinstall", Namespace: "myns"},
			Disabled:    false,
			InstallType: installConfig,
		}

		installedMesh, err := installer.EnsureIstioInstall(context.TODO(), install)
		Expect(err).NotTo(HaveOccurred())
		Expect(installedMesh.Metadata.Name).To(Equal(install.Metadata.Name))

		// installed manifest should be set
		Expect(install.InstalledManifest).NotTo(HaveLen(0))
		installedManifests, err := helm.NewManifestsFromGzippedString(install.InstalledManifest)
		Expect(err).NotTo(HaveOccurred())
		Expect(installedManifests).To(Equal(createdManifests))

		// should be set by install
		Expect(install.InstalledMesh).NotTo(BeNil())
		Expect(*install.InstalledMesh).To(Equal(installedMesh.Metadata.Ref()))

		Expect(installedMesh.Metadata.Name).To(Equal(install.Metadata.Name))

		// enable prometheus
		installConfig.Istio.InstallPrometheus = true
		installedMesh, err = installer.EnsureIstioInstall(context.TODO(), install)
		Expect(err).NotTo(HaveOccurred())

		// update should propogate thru
		Expect(install.InstalledManifest).NotTo(HaveLen(0))
		installedManifests, err = helm.NewManifestsFromGzippedString(install.InstalledManifest)
		Expect(err).NotTo(HaveOccurred())
		Expect(installedManifests).To(Equal(updatedManifests))

		// uninstall should work
		install.Disabled = true
		installedMesh, err = installer.EnsureIstioInstall(context.TODO(), install)
		Expect(err).NotTo(HaveOccurred())
		Expect(installedMesh).To(BeNil())
		Expect(install.InstalledManifest).To(HaveLen(0))

		Expect(deletedManifests).To(Equal(updatedManifests))
	})
})
