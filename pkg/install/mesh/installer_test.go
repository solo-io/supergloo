package mesh

import (
	"context"

	"github.com/solo-io/supergloo/pkg/install/mesh/istio"

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
	installer := defaultInstaller{helmInstaller: helm.NewMockHelm(
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

		istioConfig := &v1.MeshInstall_IstioMesh{
			IstioMesh: &v1.IstioInstall{
				IstioVersion: istio.IstioVersion106,
			},
		}
		installConfig := &v1.Install_Mesh{
			Mesh: &v1.MeshInstall{
				InstallType: istioConfig,
			},
		}

		install := &v1.Install{
			Metadata:              core.Metadata{Name: "myinstall", Namespace: "myns"},
			Disabled:              false,
			InstallationNamespace: ns,
			InstallType:           installConfig,
		}

		installedMesh, err := installer.EnsureMeshInstall(context.TODO(), install)
		Expect(err).NotTo(HaveOccurred())
		Expect(installedMesh.Metadata.Name).To(Equal(install.Metadata.Name))

		// installed manifest should be set
		Expect(install.InstalledManifest).NotTo(HaveLen(0))
		installedManifests, err := helm.NewManifestsFromGzippedString(install.InstalledManifest)
		Expect(err).NotTo(HaveOccurred())
		Expect(installedManifests).To(Equal(createdManifests))

		Expect(install.InstallType).To(BeAssignableToTypeOf(&v1.Install_Mesh{}))
		mesh := install.InstallType.(*v1.Install_Mesh)
		// should be set by install
		Expect(mesh.Mesh.InstalledMesh).NotTo(BeNil())
		Expect(*mesh.Mesh.InstalledMesh).To(Equal(installedMesh.Metadata.Ref()))

		Expect(installedMesh.Metadata.Name).To(Equal(install.Metadata.Name))

		// enable prometheus
		istioConfig.IstioMesh.InstallPrometheus = true
		installConfig.Mesh.InstallType = istioConfig
		installedMesh, err = installer.EnsureMeshInstall(context.TODO(), install)
		Expect(err).NotTo(HaveOccurred())

		// update should propogate thru
		Expect(install.InstalledManifest).NotTo(HaveLen(0))
		installedManifests, err = helm.NewManifestsFromGzippedString(install.InstalledManifest)
		Expect(err).NotTo(HaveOccurred())
		Expect(installedManifests).To(Equal(updatedManifests))

		// uninstall should work
		install.Disabled = true
		installedMesh, err = installer.EnsureMeshInstall(context.TODO(), install)
		Expect(err).NotTo(HaveOccurred())
		Expect(installedMesh).To(BeNil())
		Expect(install.InstalledManifest).To(HaveLen(0))

		Expect(deletedManifests).To(Equal(updatedManifests))
	})

	Context("self signed cert option", func() {
		It("sets self-signed cert to be false when the input mesh has a custom root cert defined", func() {

			istioConfig := &v1.MeshInstall_IstioMesh{
				IstioMesh: &v1.IstioInstall{
					IstioVersion:   istio.IstioVersion106,
					CustomRootCert: &core.ResourceRef{"foo", "bar"},
				},
			}
			installConfig := &v1.Install_Mesh{
				Mesh: &v1.MeshInstall{
					InstallType: istioConfig,
				},
			}

			install := &v1.Install{
				Metadata:              core.Metadata{Name: "myinstall", Namespace: "myns"},
				Disabled:              false,
				InstallationNamespace: ns,
				InstallType:           installConfig,
			}

			_, err := installer.EnsureMeshInstall(context.TODO(), install)
			Expect(err).NotTo(HaveOccurred())

			man := createdManifests.Find("istio/charts/security/templates/deployment.yaml")
			Expect(man).NotTo(BeNil())
			Expect(man.Content).To(ContainSubstring("--self-signed-ca=false"))
		})
		It("sets self-signed cert to be true when the input mesh has no custom root cert defined", func() {

			istioConfig := &v1.MeshInstall_IstioMesh{
				IstioMesh: &v1.IstioInstall{
					IstioVersion: istio.IstioVersion106,
				},
			}
			installConfig := &v1.Install_Mesh{
				Mesh: &v1.MeshInstall{
					InstallType: istioConfig,
				},
			}

			install := &v1.Install{
				Metadata:              core.Metadata{Name: "myinstall", Namespace: "myns"},
				Disabled:              false,
				InstallationNamespace: ns,
				InstallType:           installConfig,
			}

			_, err := installer.EnsureMeshInstall(context.TODO(), install)
			Expect(err).NotTo(HaveOccurred())

			man := createdManifests.Find("istio/charts/security/templates/deployment.yaml")
			Expect(man).NotTo(BeNil())
			Expect(man.Content).To(ContainSubstring("--self-signed-ca=true"))
		})
	})
})
