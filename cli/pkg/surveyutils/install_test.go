package surveyutils_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/gloo/pkg/cliutil/testutil"
	skclients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"
	"github.com/solo-io/supergloo/cli/pkg/options"
	. "github.com/solo-io/supergloo/cli/pkg/surveyutils"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/install/istio"
)

var _ = Describe("SurveyIstioInstall", func() {
	It("should create the expected istio install ", func() {
		namespace := clients.MustGetNamespaces()[1]

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
})

var _ = Describe("SurveyGlooInstall", func() {
	BeforeEach(func() {
		clients.UseMemoryClients()
		_, _ = clients.MustMeshClient().Write(&v1.Mesh{Metadata: core.Metadata{Namespace: "my", Name: "mesh"}}, skclients.WriteOpts{})
	})
	It("should create the expected gloo install ", func() {
		namespace := clients.MustGetNamespaces()[1]

		testutil.ExpectInteractive(func(c *testutil.Console) {
			c.ExpectString("which namespace to install to? ")
			c.PressDown()
			c.SendLine("")
			c.ExpectString("which version of Gloo to install? ")
			c.SendLine("")
			c.ExpectString("add a mesh (choose <done> to finish): ")
			c.PressDown()
			c.SendLine("")
			c.ExpectString("add a mesh (choose <done> to finish): ")
			c.SendLine("")
			c.ExpectString("update an existing install? ")
			c.SendLine("n")
			c.ExpectEOF()
		}, func() {
			var in options.Install
			err := SurveyGlooInstall(context.TODO(), &in)
			Expect(err).NotTo(HaveOccurred())
			Expect(in.InstallationNamespace.Gloo).To(Equal(namespace))
			Expect(in.GlooIngressInstall.GlooVersion).To(Equal("latest"))
			Expect(in.GlooIngressInstall.Meshes).To(HaveLen(1))
			Expect(in.GlooIngressInstall.Meshes[0]).To(Equal(&core.ResourceRef{Namespace: "my", Name: "mesh"}))
			Expect(in.Update).To(Equal(false))
		})
	})
})
