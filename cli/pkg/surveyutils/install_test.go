package surveyutils_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/gloo/pkg/cliutil/testutil"
	"github.com/solo-io/supergloo/cli/pkg/helpers"
	"github.com/solo-io/supergloo/cli/pkg/options"
	. "github.com/solo-io/supergloo/cli/pkg/surveyutils"
	"github.com/solo-io/supergloo/pkg/install/istio"
)

var _ = Describe("Metadata", func() {
	It("should create the expected istio install ", func() {
		namespace := helpers.MustGetNamespaces()[1]

		testutil.ExpectInteractive(func(c *testutil.Console) {
			c.ExpectString("which namespace to install to? ")
			c.PressDown()
			c.SendLine("")
			c.ExpectString("which version of Istio to install? ")
			c.PressDown()
			c.SendLine("")
			c.ExpectString("enable mtls? ")
			c.SendLine("y")
			c.ExpectString("enable auto-injection? ")
			c.SendLine("y")
			c.ExpectString("add grafana to the install? ")
			c.SendLine("y")
			c.ExpectString("add prometheus to the install? ")
			c.SendLine("y")
			c.ExpectString("add jaeger to the install? ")
			c.SendLine("y")
			c.ExpectString("update an existing install? ")
			c.SendLine("y")
			c.ExpectEOF()
		}, func() {
			var in options.Install
			err := SurveyIstioInstall(&in)
			Expect(err).NotTo(HaveOccurred())
			Expect(in.InstallationNamespace.Istio).To(Equal(namespace))
			Expect(in.IstioInstall.IstioVersion).To(Equal(istio.IstioVersion105))
			Expect(in.IstioInstall.EnableMtls).To(Equal(true))
			Expect(in.IstioInstall.EnableAutoInject).To(Equal(true))
			Expect(in.IstioInstall.InstallGrafana).To(Equal(true))
			Expect(in.IstioInstall.InstallPrometheus).To(Equal(true))
			Expect(in.IstioInstall.InstallJaeger).To(Equal(true))
			Expect(in.Update).To(Equal(true))
		})
	})

	It("should create the expected gloo install ", func() {
		namespace := helpers.MustGetNamespaces()[1]

		testutil.ExpectInteractive(func(c *testutil.Console) {
			c.ExpectString("which namespace to install to? ")
			c.PressDown()
			c.SendLine("")
			c.ExpectString("which version of Gloo to install? ")
			c.SendLine("")
			c.ExpectString("update an existing install? ")
			c.SendLine("n")
			c.ExpectEOF()
		}, func() {
			var in options.Install
			err := SurveyGlooInstall(&in)
			Expect(err).NotTo(HaveOccurred())
			Expect(in.InstallationNamespace.Gloo).To(Equal(namespace))
			Expect(in.GlooIngressInstall.GlooVersion).To(Equal("latest"))
			Expect(in.Update).To(Equal(false))
		})
	})
})
