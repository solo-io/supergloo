package install_test

import (
	"fmt"

	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"

	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/install/linkerd"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	skclients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/cli/test/utils"
	"github.com/solo-io/supergloo/test/inputs"
)

var _ = Describe("Install", func() {

	BeforeEach(func() {
		clients.UseMemoryClients()
	})

	getInstall := func(name string) *v1.Install {
		in, err := clients.MustInstallClient().Read("supergloo-system", name, skclients.ReadOpts{})
		ExpectWithOffset(1, err).NotTo(HaveOccurred())
		return in
	}

	Context("non-interactive", func() {
		It("should create the expected install ", func() {
			installAndVerifyLinkerd := func(
				name,
				namespace,
				version string,
				mtls,
				autoInject bool) {

				err := utils.Supergloo("install linkerd " +
					fmt.Sprintf("--name=%v ", name) +
					fmt.Sprintf("--installation-namespace linkerd ") +
					fmt.Sprintf("--version=%v ", version) +
					fmt.Sprintf("--mtls=%v ", mtls) +
					fmt.Sprintf("--auto-inject=%v ", autoInject))
				if version == "badver" {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("is not a supported linkerd version"))
					return
				}

				Expect(err).NotTo(HaveOccurred())
				install := getInstall(name)
				linkerd := MustLinkerdInstallType(install)
				Expect(linkerd.Linkerd.Version).To(Equal(version))
				Expect(linkerd.Linkerd.EnableMtls).To(Equal(mtls))
				Expect(linkerd.Linkerd.EnableAutoInject).To(Equal(autoInject))
			}

			installAndVerifyLinkerd("a1a", "ns", linkerd.Version_stable230, true, true)
			installAndVerifyLinkerd("b1a", "ns", linkerd.Version_stable230, false, false)
			installAndVerifyLinkerd("c1a", "ns", "badver", false, false)
		})
		It("should enable an existing + disabled install", func() {
			name := "input"
			namespace := "ns"
			inst := inputs.LinkerdInstall(name, namespace, "any", linkerd.Version_stable230, true)
			ic := clients.MustInstallClient()
			_, err := ic.Write(inst, skclients.WriteOpts{})
			Expect(err).NotTo(HaveOccurred())

			err = utils.Supergloo("install linkerd " +
				fmt.Sprintf("--name=%v ", name) +
				fmt.Sprintf("--namespace=%v ", namespace))
			Expect(err).NotTo(HaveOccurred())

			updatedInstall, err := ic.Read(namespace, name, skclients.ReadOpts{})
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedInstall.Disabled).To(BeFalse())

		})
		It("should error enable on existing enabled install", func() {
			name := "input"
			namespace := "ns"
			inst := inputs.LinkerdInstall(name, namespace, "any", linkerd.Version_stable230, false)
			ic := clients.MustInstallClient()
			_, err := ic.Write(inst, skclients.WriteOpts{})
			Expect(err).NotTo(HaveOccurred())

			err = utils.Supergloo("install linkerd " +
				fmt.Sprintf("--name=%v ", name) +
				fmt.Sprintf("--namespace=%v ", namespace))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("already installed and enabled"))
		})
		It("should update existing enabled install if --update set to true", func() {
			name := "input"
			namespace := "ns"
			inst := inputs.LinkerdInstall(name, namespace, "linkerd-system", linkerd.Version_stable230, false)
			Expect(inst.InstallType).To(BeAssignableToTypeOf(&v1.Install_Mesh{}))
			ic := clients.MustInstallClient()
			_, err := ic.Write(inst, skclients.WriteOpts{})
			Expect(err).NotTo(HaveOccurred())

			err = utils.Supergloo("install linkerd " +
				fmt.Sprintf("--name=%v ", name) +
				fmt.Sprintf("--namespace=%v ", namespace) +
				"--mtls=true " +
				"--update=true ")
			Expect(err).NotTo(HaveOccurred())

			updatedInstall, err := ic.Read(namespace, name, skclients.ReadOpts{})
			Expect(err).NotTo(HaveOccurred())
			Expect(*updatedInstall).To(Equal(v1.Install{
				Metadata: core.Metadata{
					Name:            name,
					Namespace:       namespace,
					ResourceVersion: updatedInstall.Metadata.ResourceVersion,
				},
				InstallationNamespace: "linkerd-system",
				Disabled:              false,
				InstallType: &v1.Install_Mesh{
					Mesh: &v1.MeshInstall{
						MeshInstallType: &v1.MeshInstall_Linkerd{
							Linkerd: &v1.LinkerdInstall{
								Version:          linkerd.Version_stable230,
								EnableAutoInject: true,
								EnableMtls:       true,
							},
						},
					},
				},
			}))
		})
	})
})

func MustLinkerdInstallType(install *v1.Install) *v1.MeshInstall_Linkerd {
	Expect(install.InstallType).To(BeAssignableToTypeOf(&v1.Install_Mesh{}))
	mesh := install.InstallType.(*v1.Install_Mesh)
	Expect(mesh.Mesh.MeshInstallType).To(BeAssignableToTypeOf(&v1.MeshInstall_Linkerd{}))
	linkerdMesh := mesh.Mesh.MeshInstallType.(*v1.MeshInstall_Linkerd)
	return linkerdMesh
}
