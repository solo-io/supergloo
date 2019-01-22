package prometheus_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/test/helpers"
	"github.com/solo-io/solo-kit/test/tests/typed"
	"github.com/solo-io/supergloo/pkg/api/external/prometheus"
	. "github.com/solo-io/supergloo/pkg/api/external/prometheus/v1"
	"github.com/solo-io/supergloo/test/utils"
	"gopkg.in/yaml.v2"
	"k8s.io/client-go/kubernetes"
)

// TODO(ilackarms): this is a partially complete test
// to run it, it currently requires deploying the istio protheus.yml configmap to istio-system in kubernetes

var _ = Describe("Prometheus Config Conversion", func() {
	var (
		namespace string
	)
	for _, test := range []typed.ResourceClientTester{
		&typed.KubeConfigMapRcTester{},
	} {
		Context("resource client backed by "+test.Description(), func() {
			var (
				client PrometheusConfigClient
				err    error
				kube   kubernetes.Interface
			)
			BeforeEach(func() {
				namespace = helpers.RandString(6)
				fact := test.Setup(namespace)
				client, err = NewPrometheusConfigClient(fact)
				Expect(err).NotTo(HaveOccurred())
				kube = fact.(*factory.KubeConfigMapClientFactory).Clientset
			})
			AfterEach(func() {
				test.Teardown(namespace)
			})
			It("converts prometheus configs from proto type to go struct type", func() {
				testConvert(namespace, kube, client)
			})
			It("CRUDs Configs", func() {
				testPrometheusSerializer(namespace, kube, client)
			})
		})
	}
})

func testConvert(namespace string, kube kubernetes.Interface, client PrometheusConfigClient) {
	name := "prometheus-config"
	err := utils.DeployPrometheusConfigmap(namespace, name, kube)
	Expect(err).NotTo(HaveOccurred())
	original, err := client.Read(namespace, name, clients.ReadOpts{})
	Expect(err).NotTo(HaveOccurred())
	cfg, err := prometheus.ConfigFromResource(original)
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg.ScrapeConfigs).To(HaveLen(7))
	Expect(cfg.ScrapeConfigs[0].JobName).To(Equal("kubernetes-apiservers"))
}

func testPrometheusSerializer(namespace string, kube kubernetes.Interface, client PrometheusConfigClient) {
	name := "prometheus-config"
	err := utils.DeployPrometheusConfigmap(namespace, name, kube)
	Expect(err).NotTo(HaveOccurred())
	original, err := client.Read(namespace, name, clients.ReadOpts{})
	Expect(err).NotTo(HaveOccurred())
	cfg, err := prometheus.ConfigFromResource(original)
	Expect(err).NotTo(HaveOccurred())
	converted, err := prometheus.ConfigToResource(cfg)
	Expect(err).NotTo(HaveOccurred())
	cfg2, err := prometheus.ConfigFromResource(converted)
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).To(Equal(cfg2))

	yam1, err := yaml.Marshal(cfg)
	Expect(err).NotTo(HaveOccurred())
	yam2, err := yaml.Marshal(cfg2)
	Expect(err).NotTo(HaveOccurred())
	str1 := string(yam1)
	str2 := string(yam2)
	Expect(str1).To(Equal(str2))

	// TODO(ilackarms): test to preserve comments
	//Expect(str1).To(Equal(utils.BasicPrometheusConfig))
}
