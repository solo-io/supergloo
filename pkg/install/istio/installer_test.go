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

		istioConfig := &v1.MeshInstall_IstioMesh{
			IstioMesh: &v1.IstioInstall{
				IstioVersion: IstioVersion106,
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

		installedMesh, err := installer.EnsureIstioInstall(context.TODO(), install, nil)
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

		// expect an error if installed mesh is not present in the mesh list
		_, err = installer.EnsureIstioInstall(context.TODO(), install, nil)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("installed mesh not found"))

		// enable prometheus
		istioConfig.IstioMesh.InstallPrometheus = true
		installConfig.Mesh.InstallType = istioConfig
		installedMesh, err = installer.EnsureIstioInstall(context.TODO(), install, v1.MeshList{installedMesh})
		Expect(err).NotTo(HaveOccurred())

		// update should propogate thru
		Expect(install.InstalledManifest).NotTo(HaveLen(0))
		installedManifests, err = helm.NewManifestsFromGzippedString(install.InstalledManifest)
		Expect(err).NotTo(HaveOccurred())
		Expect(installedManifests).To(Equal(updatedManifests))

		// uninstall should work
		install.Disabled = true
		installedMesh, err = installer.EnsureIstioInstall(context.TODO(), install, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(installedMesh).To(BeNil())
		Expect(install.InstalledManifest).To(HaveLen(0))

		Expect(deletedManifests).To(Equal(updatedManifests))
	})

	Context("self signed cert option", func() {
		It("sets self-signed cert to be false when the input install has a custom root cert defined", func() {

			istioConfig := &v1.MeshInstall_IstioMesh{
				IstioMesh: &v1.IstioInstall{
					IstioVersion:   IstioVersion106,
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

			_, err := installer.EnsureIstioInstall(context.TODO(), install, nil)
			Expect(err).NotTo(HaveOccurred())

			man := createdManifests.Find("istio/charts/security/templates/deployment.yaml")
			Expect(man).NotTo(BeNil())
			Expect(man.Content).To(ContainSubstring("--self-signed-ca=false"))
		})
		It("sets self-signed cert to be true when the input install has no custom root cert defined", func() {

			istioConfig := &v1.MeshInstall_IstioMesh{
				IstioMesh: &v1.IstioInstall{
					IstioVersion: IstioVersion106,
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

			_, err := installer.EnsureIstioInstall(context.TODO(), install, nil)
			Expect(err).NotTo(HaveOccurred())

			man := createdManifests.Find("istio/charts/security/templates/deployment.yaml")
			Expect(man).NotTo(BeNil())
			Expect(man.Content).To(ContainSubstring("--self-signed-ca=true"))
		})
		It("sets self-signed cert to be false when the input mesh has a custom root cert defined", func() {

			istioConfig := &v1.MeshInstall_IstioMesh{
				IstioMesh: &v1.Istio{
					IstioVersion:   IstioVersion106,
					CustomRootCert: &core.ResourceRef{"foo", "bar"},
				},
			}
			mesh := &v1.Mesh{
				Metadata:   core.Metadata{Name: "mymesh", Namespace: "myns"},
				MtlsConfig: &v1.MtlsConfig{MtlsEnabled: true, RootCertificate: &core.ResourceRef{"root", "cert"}},
			}
			ref := mesh.Metadata.Ref()
			installConfig := &v1.Install_Mesh{
				Mesh: &v1.MeshInstall{
					InstallType:   istioConfig,
					InstalledMesh: &ref,
				},
			}

			install := &v1.Install{
				Metadata:              core.Metadata{Name: "myinstall", Namespace: "myns"},
				Disabled:              false,
				InstallationNamespace: ns,
				InstallType:           installConfig,
			}
			_, err := installer.EnsureIstioInstall(context.TODO(), install, v1.MeshList{mesh})
			Expect(err).NotTo(HaveOccurred())

			man := createdManifests.Find("istio/charts/security/templates/deployment.yaml")
			Expect(man).NotTo(BeNil())
			Expect(man.Content).To(ContainSubstring("--self-signed-ca=false"))
		})
	})
})
