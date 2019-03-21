package surveyutils_test

import (
	"context"

	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/gloo/pkg/cliutil/testutil"
	skclients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/cli/pkg/options"
	. "github.com/solo-io/supergloo/cli/pkg/surveyutils"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

var _ = Describe("RootCert", func() {
	BeforeEach(func() {
		clients.UseMemoryClients()
		_, _ = clients.MustMeshClient().Write(&v1.Mesh{Metadata: core.Metadata{Namespace: "my", Name: "mesh"}}, skclients.WriteOpts{})
		_, _ = clients.MustMeshClient().Write(&v1.Mesh{Metadata: core.Metadata{Namespace: "your", Name: "mesh"}}, skclients.WriteOpts{})
		_, _ = clients.MustTlsSecretClient().Write(&v1.TlsSecret{Metadata: core.Metadata{Namespace: "my", Name: "secret"}}, skclients.WriteOpts{})
		_, _ = clients.MustTlsSecretClient().Write(&v1.TlsSecret{Metadata: core.Metadata{Namespace: "your", Name: "secret"}}, skclients.WriteOpts{})
	})
	It("sets the target mesh and root cert from a list", func() {
		testutil.ExpectInteractive(func(c *testutil.Console) {
			c.ExpectString("select the mesh for which you wish to set the root cert")
			c.PressDown()
			c.SendLine("")
			c.ExpectString("select the tls secret to use as the new root cert")
			c.PressDown()
			c.SendLine("")
			c.ExpectEOF()
		}, func() {
			var input options.SetRootCert
			err := SurveySetRootCert(context.TODO(), &input)
			Expect(err).NotTo(HaveOccurred())

			Expect(core.ResourceRef(input.TargetMesh)).To(Equal(core.ResourceRef{Namespace: "your", Name: "mesh"}))
			Expect(core.ResourceRef(input.TlsSecret)).To(Equal(core.ResourceRef{Namespace: "my", Name: "secret"}))
		})
	})
	It("provides an option to leave the root cert empty", func() {
		testutil.ExpectInteractive(func(c *testutil.Console) {
			c.ExpectString("select the mesh for which you wish to set the root cert")
			c.PressDown()
			c.SendLine("")
			c.ExpectString("select the tls secret to use as the new root cert")
			c.SendLine("")
			c.ExpectEOF()
		}, func() {
			var input options.SetRootCert
			err := SurveySetRootCert(context.TODO(), &input)
			Expect(err).NotTo(HaveOccurred())

			Expect(core.ResourceRef(input.TargetMesh)).To(Equal(core.ResourceRef{Namespace: "your", Name: "mesh"}))
			Expect(core.ResourceRef(input.TlsSecret)).To(Equal(core.ResourceRef{Namespace: "", Name: ""}))
		})
	})
})
