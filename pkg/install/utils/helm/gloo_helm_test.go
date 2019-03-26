package helm_test

import (
	"context"

	"github.com/solo-io/supergloo/pkg/install/gloo"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/install/utils/helm"
)

var _ = Describe("gloo helm installer", func() {
	installer := gloo.NewDefaultInstaller(helm.NewHelmInstaller())
	ns := "gloo-system"
	It("installs, upgrades, and uninstalls from an install object", func() {

		istioMesh := []*core.ResourceRef{
			{
				Namespace: "supergloo-system",
				Name:      "istio",
			},
		}
		glooConfig := &v1.MeshIngressInstall_Gloo{
			Gloo: &v1.GlooInstall{
				GlooVersion: "0.11.1",
				Meshes:      istioMesh,
			},
		}
		installConfig := &v1.Install_Ingress{
			Ingress: &v1.MeshIngressInstall{
				IngressInstallType: glooConfig,
			},
		}

		install := &v1.Install{
			Metadata:              core.Metadata{Name: "myinstall", Namespace: "myns"},
			Disabled:              false,
			InstallationNamespace: ns,
			InstallType:           installConfig,
		}

		meshes := v1.MeshList{
			&v1.Mesh{
				Metadata: core.Metadata{
					Name:      "istio",
					Namespace: "supergloo-system",
				},
			},
		}

		meshIngress, err := installer.EnsureGlooInstall(context.TODO(), install, meshes, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(meshIngress.Metadata.Name).To(Equal(install.Metadata.Name))

		// installed manifest should be set
		Expect(install.InstalledManifest).NotTo(HaveLen(0))
		installedManifests, err := helm.NewManifestsFromGzippedString(install.InstalledManifest)
		Expect(err).NotTo(HaveOccurred())
		Expect(installedManifests).NotTo(BeEmpty())

		Expect(install.InstallType).To(BeAssignableToTypeOf(&v1.Install_Ingress{}))
		installedIngress := install.InstallType.(*v1.Install_Ingress)
		// should be set by install
		Expect(installedIngress.Ingress.InstalledIngress).NotTo(BeNil())
		Expect(*installedIngress.Ingress.InstalledIngress).To(Equal(meshIngress.Metadata.Ref()))
		Expect(meshIngress.Metadata.Name).To(Equal(install.Metadata.Name))
		ref := meshes[0].Metadata.Ref()
		Expect(meshIngress.Meshes).To(ContainElement(&ref))

		// uninstall should work
		install.Disabled = true
		uninstalledIngress, err := installer.EnsureGlooInstall(context.TODO(), install, nil, v1.MeshIngressList{meshIngress})
		Expect(err).NotTo(HaveOccurred())
		Expect(uninstalledIngress).To(BeNil())
		Expect(install.InstalledManifest).To(HaveLen(0))

		_, err = helm.NewManifestsFromGzippedString(install.InstalledManifest)
		Expect(err).NotTo(HaveOccurred())

	})

})
