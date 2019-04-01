package surveyutils_test

import (
	"context"

	"github.com/solo-io/supergloo/cli/pkg/surveyutils"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"

	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/gloo/pkg/cliutil/testutil"
	skclients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
)

var _ = Describe("SelectMeshes", func() {
	BeforeEach(func() {
		clients.UseMemoryClients()
		istioMesh := &v1.Mesh{
			MtlsConfig: &v1.MtlsConfig{
				MtlsEnabled: true,
			},
			MeshType: &v1.Mesh_Istio{
				Istio: &v1.IstioMesh{
					InstallationNamespace: "istio-system",
				},
			},
			Metadata: core.Metadata{
				Name:      "istio",
				Namespace: "supergloo-system",
			},
		}
		glooMeshIngress := &v1.MeshIngress{
			Meshes: []*core.ResourceRef{
				{
					Namespace: "istio-system",
					Name:      "istio",
				},
			},
			Metadata: core.Metadata{
				Name:      "gloo",
				Namespace: "supergloo-system",
			},
			InstallationNamespace: "gloo-system",
			MeshIngressType: &v1.MeshIngress_Gloo{
				Gloo: &v1.GlooMeshIngress{},
			},
		}
		_, err := clients.MustMeshClient().Write(istioMesh, skclients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())
		_, err = clients.MustMeshIngressClient().Write(glooMeshIngress, skclients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())
	})
	It("selects mesh ingress from the list", func() {
		testutil.ExpectInteractive(func(c *testutil.Console) {
			c.ExpectString("select a mesh ingress: ")
			c.PressDown()
			c.SendLine("")
			c.ExpectEOF()
		}, func() {
			meshes, err := surveyutils.SurveyMeshIngress(context.TODO())
			Expect(err).NotTo(HaveOccurred())
			Expect(meshes).To(BeEquivalentTo(core.ResourceRef{
				Name:      "gloo",
				Namespace: "supergloo-system",
			}))
		})
	})

	It("selects mesh ingress from the list until done is selected", func() {
		testutil.ExpectInteractive(func(c *testutil.Console) {
			c.ExpectString("add a mesh ingress (choose <done> to finish): ")
			c.PressDown()
			c.SendLine("")
			c.ExpectString("add a mesh ingress (choose <done> to finish): ")
			c.SendLine("")
			c.ExpectEOF()
		}, func() {
			meshes, err := surveyutils.SurveyMeshIngresses(context.TODO())
			Expect(err).NotTo(HaveOccurred())
			Expect(meshes).To(BeEquivalentTo([]*core.ResourceRef{
				{
					Name:      "gloo",
					Namespace: "supergloo-system",
				},
			}))
		})
	})
})
