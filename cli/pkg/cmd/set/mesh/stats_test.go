package mesh_test

import (
	"fmt"

	"github.com/solo-io/supergloo/api/external/prometheus"

	promv1 "github.com/solo-io/supergloo/pkg/api/external/prometheus/v1"

	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	skclients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/cli/test/utils"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

var _ = Describe("Stats", func() {
	mesh := core.Metadata{Namespace: "my", Name: "mesh"}
	promCfg1 := core.Metadata{Namespace: "my", Name: "cfg1"}
	promCfg2 := core.Metadata{Namespace: "my", Name: "cfg2"}
	BeforeEach(func() {
		clients.UseMemoryClients()
		_, err := clients.MustMeshClient().Write(&v1.Mesh{
			Metadata: mesh,
			MeshType: &v1.Mesh_Istio{Istio: &v1.IstioMesh{}},
		}, skclients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())
		_, err = clients.MustPrometheusConfigClient().Write(&promv1.PrometheusConfig{
			PrometheusConfig: prometheus.PrometheusConfig{Metadata: promCfg1}}, skclients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())
		_, err = clients.MustPrometheusConfigClient().Write(&promv1.PrometheusConfig{
			PrometheusConfig: prometheus.PrometheusConfig{Metadata: promCfg2}}, skclients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())
	})

	It("updates the prometheus configmap refs on an existing mesh", func() {
		err := utils.Supergloo(fmt.Sprintf("set mesh stats --target-mesh %v "+
			" --prometheus-configmap %v --prometheus-configmap %v", mesh.Ref().Key(), promCfg1.Ref().Key(), promCfg2.Ref().Key()))
		Expect(err).NotTo(HaveOccurred())
		meshWithCert, err := clients.MustMeshClient().Read(mesh.Namespace, mesh.Name, skclients.ReadOpts{})
		Expect(err).NotTo(HaveOccurred())
		Expect(meshWithCert.MonitoringConfig).NotTo(BeNil())
		Expect(meshWithCert.MonitoringConfig.PrometheusConfigmaps).To(Equal([]core.ResourceRef{
			promCfg1.Ref(),
			promCfg2.Ref(),
		}))
	})

	It("errors if no target mesh provided the prometheus configmap refs on an existing mesh", func() {
		err := utils.Supergloo(fmt.Sprintf("set mesh stats "+
			"--prometheus-configmap %v --prometheus-configmap %v", promCfg1.Ref().Key(), promCfg2.Ref().Key()))
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("must provide --target-mesh"))
	})
})
