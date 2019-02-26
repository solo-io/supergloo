package install_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
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
					return
				}

				Expect(err).NotTo(HaveOccurred())
				install := getInstall(name)
				Expect(install.InstallType).To(BeAssignableToTypeOf(&v1.Install_Istio_{}))
				istio := install.InstallType.(*v1.Install_Istio_)
				Expect(istio.Istio.IstioVersion).To(Equal(version))
				Expect(istio.Istio.EnableMtls).To(Equal(mtls))
				Expect(istio.Istio.EnableAutoInject).To(Equal(autoInject))
				Expect(istio.Istio.InstallPrometheus).To(Equal(prometheus))
				Expect(istio.Istio.InstallJaeger).To(Equal(jaeger))
				Expect(istio.Istio.InstallGrafana).To(Equal(grafana))
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

		})
	})
})
