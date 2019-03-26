package install_test

import (
	"fmt"
	"strings"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"

	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	skclients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/supergloo/cli/test/utils"
	"github.com/solo-io/supergloo/test/inputs"
)

var _ = Describe("gloo install", func() {

	BeforeEach(func() {
		clients.UseMemoryClients()
	})

	getInstall := func(name string) *v1.Install {
		in, err := clients.MustInstallClient().Read("supergloo-system", name, skclients.ReadOpts{})
		ExpectWithOffset(1, err).NotTo(HaveOccurred())
		return in
	}

	FContext("non-interactive", func() {
		It("should create the expected install ", func() {
			installAndVerifyGloo := func(
				name,
				namespace,
				version string) {

				err := utils.Supergloo("install gloo " +
					fmt.Sprintf("--name=%v ", name) +
					fmt.Sprintf("--installation-namespace %s ", namespace) +
					fmt.Sprintf("--version=%v ", version))
				if version == "badver" {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("is not a supported gloo version"))
					return
				}

				Expect(err).NotTo(HaveOccurred())
				install := getInstall(name)
				glooIngress := MustGlooInstallType(install)
				if version != "latest" {
					Expect(glooIngress.Gloo.GlooVersion).To(Equal(strings.TrimPrefix(version, "v")))
				}
			}

			installAndVerifyGloo("a1a", "ns", "latest")
			installAndVerifyGloo("b1a", "ns", "v0.13.0")
			installAndVerifyGloo("c1a", "ns", "badver")
		})
		It("should enable an existing + disabled install", func() {
			name := "input"
			namespace := "ns"
			inst := inputs.GlooIstall(name, namespace, "any", "v0.13.0", true)
			ic := clients.MustInstallClient()
			_, err := ic.Write(inst, skclients.WriteOpts{})
			Expect(err).NotTo(HaveOccurred())

			err = utils.Supergloo("install gloo " +
				fmt.Sprintf("--name=%v ", name) +
				fmt.Sprintf("--installation-namespace %s ", namespace) +
				fmt.Sprintf("--namespace=%v ", namespace))
			Expect(err).NotTo(HaveOccurred())

			updatedInstall, err := ic.Read(namespace, name, skclients.ReadOpts{})
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedInstall.Disabled).To(BeFalse())

		})
		It("should error enable on existing enabled install", func() {
			name := "input"
			namespace := "ns"
			inst := inputs.GlooIstall(name, namespace, "any", "v0.13.0", false)
			ic := clients.MustInstallClient()
			_, err := ic.Write(inst, skclients.WriteOpts{})
			Expect(err).NotTo(HaveOccurred())

			err = utils.Supergloo("install gloo " +
				fmt.Sprintf("--name=%v ", name) +
				fmt.Sprintf("--installation-namespace %s ", namespace) +
				fmt.Sprintf("--namespace=%v ", namespace))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("already installed and enabled"))
		})

		It("can install with a mesh reference", func() {
			name := "input"
			namespace := "ns"
			err := utils.Supergloo("install gloo " +
				fmt.Sprintf("--name=%v ", name) +
				fmt.Sprintf("--installation-namespace %s ", namespace) +
				fmt.Sprintf("--namespace=%v ", namespace) +
				fmt.Sprintf("--meshes one.two"))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("is not a valid mesh"))
			err = utils.Supergloo("install gloo " +
				fmt.Sprintf("--name=%v ", name) +
				fmt.Sprintf("--installation-namespace %s ", namespace) +
				fmt.Sprintf("--namespace=%v ", namespace) +
				fmt.Sprintf("--meshes one"))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("is of the incorrect format"))
			meshClient := clients.MustMeshClient()
			_, err = meshClient.Write(&v1.Mesh{Metadata: core.Metadata{Name: "one", Namespace: "two"}}, skclients.WriteOpts{})
			Expect(err).NotTo(HaveOccurred())
			err = utils.Supergloo("install gloo " +
				fmt.Sprintf("--name=%v ", name) +
				fmt.Sprintf("--installation-namespace %s ", namespace) +
				fmt.Sprintf("--namespace=%v ", namespace) +
				fmt.Sprintf("--meshes two.one"))
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

func MustGlooInstallType(install *v1.Install) *v1.MeshIngressInstall_Gloo {
	Expect(install.InstallType).To(BeAssignableToTypeOf(&v1.Install_Ingress{}))
	ingress := install.InstallType.(*v1.Install_Ingress)
	Expect(ingress.Ingress.IngressInstallType).To(BeAssignableToTypeOf(&v1.MeshIngressInstall_Gloo{}))
	glooIngress := ingress.Ingress.IngressInstallType.(*v1.MeshIngressInstall_Gloo)
	return glooIngress
}
