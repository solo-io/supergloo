package prometheus_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/configmap"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/test/inputs"

	. "github.com/solo-io/supergloo/pkg/api/custom/clients/prometheus"
	v1 "github.com/solo-io/supergloo/pkg/api/external/prometheus/v1"
)

var _ = Describe("ConfigmapConverter", func() {
	c := NewPrometheusConfigmapConverter()
	fakeResourceClient, _ := configmap.NewResourceClient(nil, &v1.PrometheusConfig{}, nil, false)
	It("converts a prometheus configmap to a v1.PrometheusConfig", func() {
		configMap := inputs.PrometheusConfigMap("myname", "mynamespace")
		promCfg, err := c.FromKubeConfigMap(context.TODO(), fakeResourceClient, configMap)
		Expect(err).NotTo(HaveOccurred())
		Expect(promCfg).To(Equal(&v1.PrometheusConfig{
			Metadata:   core.Metadata{Name: "myname", Namespace: "mynamespace"},
			Prometheus: inputs.BasicPrometheusConfig,
		}))
	})
	It("returns nil/nil for a non-prometheus configmap", func() {

	})
	It("returns err for anything other than a v1.PrometheusConfig", func() {

	})
	It("converts a prometheus config resource to a v1.PrometheusConfig", func() {

	})
})
