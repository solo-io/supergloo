package surveyutils_test

import (
	"context"

	"github.com/solo-io/supergloo/api/external/prometheus"

	promv1 "github.com/solo-io/supergloo/pkg/api/external/prometheus/v1"

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

var _ = Describe("SetStats", func() {
	BeforeEach(func() {
		clients.UseMemoryClients()
		_, _ = clients.MustMeshClient().Write(&v1.Mesh{Metadata: core.Metadata{Namespace: "my", Name: "mesh"}}, skclients.WriteOpts{})
		_, _ = clients.MustMeshClient().Write(&v1.Mesh{Metadata: core.Metadata{Namespace: "your", Name: "mesh"}}, skclients.WriteOpts{})
		_, _ = clients.MustPrometheusConfigClient().Write(&promv1.PrometheusConfig{
			PrometheusConfig: prometheus.PrometheusConfig{Metadata: core.Metadata{Namespace: "my", Name: "cfg"}}}, skclients.WriteOpts{})
		_, _ = clients.MustPrometheusConfigClient().Write(&promv1.PrometheusConfig{
			PrometheusConfig: prometheus.PrometheusConfig{Metadata: core.Metadata{Namespace: "your", Name: "cfg"}}}, skclients.WriteOpts{})
	})
	It("sets the target mesh and prometheus configs from lists", func() {
		testutil.ExpectInteractive(func(c *testutil.Console) {
			c.ExpectString("select the mesh for which you wish to propagate metrics")
			c.PressDown()
			c.SendLine("")
			c.ExpectString("add a prometheus configmap (choose <done> to finish): ")
			c.PressDown()
			c.SendLine("")
			c.ExpectString("add a prometheus configmap (choose <done> to finish): ")
			c.PressDown()
			c.PressDown()
			c.SendLine("")
			c.ExpectString("add a prometheus configmap (choose <done> to finish): ")
			c.SendLine("")
			c.ExpectEOF()
		}, func() {
			var input options.SetStats
			err := SurveySetStats(context.TODO(), &input)
			Expect(err).NotTo(HaveOccurred())

			Expect(core.ResourceRef(input.TargetMesh)).To(Equal(
				core.ResourceRef{Namespace: "your", Name: "mesh"}))
			Expect([]core.ResourceRef(input.PrometheusConfigMaps)).To(BeEquivalentTo(
				[]core.ResourceRef{{Namespace: "my", Name: "cfg"}, {Namespace: "your", Name: "cfg"}}))
		})
	})
})
