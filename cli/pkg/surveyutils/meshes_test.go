package surveyutils_test

import (
	"context"

	v1 "github.com/solo-io/supergloo/pkg/api/v1"

	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/gloo/pkg/cliutil/testutil"
	skclients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	. "github.com/solo-io/supergloo/cli/pkg/surveyutils"
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
				Namespace: "istio-system",
			},
		}
		_, err := clients.MustMeshClient().Write(istioMesh, skclients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())
	})
	It("selects mesh from the list until skip is selected", func() {
		testutil.ExpectInteractive(func(c *testutil.Console) {
			c.ExpectString("add an upstream (choose <done> to finish): ")
			c.PressDown()
			c.SendLine("")
			c.ExpectEOF()
		}, func() {
			ups, err := SurveyMeshes(context.TODO())
			Expect(err).NotTo(HaveOccurred())
			Expect(ups).To(Equal([]core.ResourceRef{
				{
					Name:      "istio",
					Namespace: "istio-system",
				},
			}))
		})
	})
})
