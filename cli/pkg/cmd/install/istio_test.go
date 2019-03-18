package install_test

import (
	"fmt"

	"github.com/solo-io/supergloo/pkg/install/mesh/istio"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/cli/pkg/helpers"
	"github.com/solo-io/supergloo/cli/test/utils"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/test/inputs"
)

var _ = Describe("Install", func() {

	BeforeEach(func() {
		helpers.UseMemoryClients()
	})

	getInstall := func(name string) *v1.Install {
		in, err := helpers.MustInstallClient().Read("supergloo-system", name, clients.ReadOpts{})
		ExpectWithOffset(1, err).NotTo(HaveOccurred())
		return in
	}

	Describe("non-interactive", func() {
		It("should create the expected install ", func() {
			installAndVerifyIstio := func(
				name,
				namespace,
				version string,
				mtls,
				autoInject,
				prometheus,
				jaeger,
				grafana bool) {

				err := utils.Supergloo("install istio " +
					fmt.Sprintf("--name=%v ", name) +
					fmt.Sprintf("--installation-namespace istio ") +
					fmt.Sprintf("--version=%v ", version) +
					fmt.Sprintf("--mtls=%v ", mtls) +
					fmt.Sprintf("--auto-inject=%v ", autoInject) +
					fmt.Sprintf("--grafana=%v ", grafana) +
					fmt.Sprintf("--prometheus=%v ", prometheus) +
					fmt.Sprintf("--jaeger=%v", jaeger))
				if version == "badver" {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("is not a suppported istio version"))
					return
				}

				Expect(err).NotTo(HaveOccurred())
				install := getInstall(name)
				istio := MustIstioInstallType(install)
				Expect(istio.IstioMesh.IstioVersion).To(Equal(version))
				Expect(istio.IstioMesh.EnableMtls).To(Equal(mtls))
				Expect(istio.IstioMesh.EnableAutoInject).To(Equal(autoInject))
				Expect(istio.IstioMesh.InstallPrometheus).To(Equal(prometheus))
				Expect(istio.IstioMesh.InstallJaeger).To(Equal(jaeger))
				Expect(istio.IstioMesh.InstallGrafana).To(Equal(grafana))
			}

			installAndVerifyIstio("a1a", "ns", "1.0.3", true, true, true, true, true)
			installAndVerifyIstio("b1a", "ns", "1.0.5", false, false, false, false, false)
			installAndVerifyIstio("c1a", "ns", "badver", false, false, false, false, false)
		})
		It("should enable an existing + disabled install", func() {
			name := "input"
			namespace := "ns"
			inst := inputs.IstioInstall(name, namespace, "any", "1.0.5", true)
			ic := helpers.MustInstallClient()
			_, err := ic.Write(inst, clients.WriteOpts{})
			Expect(err).NotTo(HaveOccurred())

			err = utils.Supergloo("install istio " +
				fmt.Sprintf("--name=%v ", name) +
				fmt.Sprintf("--namespace=%v ", namespace))
			Expect(err).NotTo(HaveOccurred())

			updatedInstall, err := ic.Read(namespace, name, clients.ReadOpts{})
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedInstall.Disabled).To(BeFalse())

		})
		It("should error enable on existing enabled install", func() {
			name := "input"
			namespace := "ns"
			inst := inputs.IstioInstall(name, namespace, "any", "1.0.5", false)
			ic := helpers.MustInstallClient()
			_, err := ic.Write(inst, clients.WriteOpts{})
			Expect(err).NotTo(HaveOccurred())

			err = utils.Supergloo("install istio " +
				fmt.Sprintf("--name=%v ", name) +
				fmt.Sprintf("--namespace=%v ", namespace))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("already installed and enabled"))
		})
		It("should update existing enabled install if --update set to true", func() {
			name := "input"
			namespace := "ns"
			inst := inputs.IstioInstall(name, namespace, "istio-system", "1.0.5", false)
			Expect(inst.InstallType).To(BeAssignableToTypeOf(&v1.Install_Mesh{}))
			istioMesh := inst.InstallType.(*v1.Install_Mesh)
			inst.InstalledManifest = "a previously installed manifest"
			istioMesh.Mesh.InstalledMesh = &core.ResourceRef{"installed", "mesh"}
			ic := helpers.MustInstallClient()
			_, err := ic.Write(inst, clients.WriteOpts{})
			Expect(err).NotTo(HaveOccurred())

			err = utils.Supergloo("install istio " +
				fmt.Sprintf("--name=%v ", name) +
				fmt.Sprintf("--namespace=%v ", namespace) +
				"--mtls=true " +
				"--update=true ")
			Expect(err).NotTo(HaveOccurred())

			updatedInstall, err := ic.Read(namespace, name, clients.ReadOpts{})
			Expect(err).NotTo(HaveOccurred())
			Expect(*updatedInstall).To(Equal(v1.Install{
				Metadata: core.Metadata{
					Name:            name,
					Namespace:       namespace,
					ResourceVersion: updatedInstall.Metadata.ResourceVersion,
				},
				InstalledManifest:     "a previously installed manifest",
				InstallationNamespace: "istio-system",
				Disabled:              false,
				InstallType: &v1.Install_Mesh{
					Mesh: &v1.MeshInstall{
						InstalledMesh: &core.ResourceRef{
							Name:      "installed",
							Namespace: "mesh",
						},
						InstallType: &v1.MeshInstall_IstioMesh{
							IstioMesh: &v1.IstioInstall{
								IstioVersion:      istio.IstioVersion106,
								EnableAutoInject:  true,
								EnableMtls:        true,
								InstallGrafana:    true,
								InstallPrometheus: true,
								InstallJaeger:     true,
							},
						},
					},
				},
			}))
		})
	})
})

func MustIstioInstallType(install *v1.Install) *v1.MeshInstall_IstioMesh {
	Expect(install.InstallType).To(BeAssignableToTypeOf(&v1.Install_Mesh{}))
	mesh := install.InstallType.(*v1.Install_Mesh)
	Expect(mesh.Mesh.InstallType).To(BeAssignableToTypeOf(&v1.MeshInstall_IstioMesh{}))
	istio := mesh.Mesh.InstallType.(*v1.MeshInstall_IstioMesh)
	return istio
}
