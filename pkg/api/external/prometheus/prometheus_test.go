package prometheus_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/utils/protoutils"
	"github.com/solo-io/solo-kit/test/helpers"
	"github.com/solo-io/solo-kit/test/tests/typed"
	"github.com/solo-io/supergloo/pkg/api/external/prometheus"
	. "github.com/solo-io/supergloo/pkg/api/external/prometheus/v1"
	"github.com/solo-io/supergloo/test/utils"
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
				client ConfigClient
				err    error
				kube   kubernetes.Interface
			)
			BeforeEach(func() {
				namespace = helpers.RandString(6)
				fact := test.Setup(namespace)
				client, err = NewConfigClient(fact)
				Expect(err).NotTo(HaveOccurred())
				kube = fact.(*factory.KubeConfigMapClientFactory).Clientset
			})
			AfterEach(func() {
				test.Teardown(namespace)
			})
			It("CRUDs Configs", func() {
				testPrometheusSerializer(namespace, kube, client)
			})
		})
	}
})

func testPrometheusSerializer(namespace string, kube kubernetes.Interface, client ConfigClient) {
	name := "prometheus-config"
	err := utils.DeployPrometheusConfigmap(namespace, name, kube)
	Expect(err).NotTo(HaveOccurred())
	read, err := client.Read(namespace, name, clients.ReadOpts{})
	Expect(err).NotTo(HaveOccurred())
	cfg, err := prometheus.ConfigFromResource(read)
	Expect(err).NotTo(HaveOccurred())
	expected, err := prometheus.ConfigToResource(cfg)
	Expect(err).NotTo(HaveOccurred())
	expected.SetMetadata(read.Metadata)
	Expect(expected).To(Equal(read))
	jsn1, err := protoutils.MarshalBytes(read.Prometheus)
	Expect(err).NotTo(HaveOccurred())
	jsn2, err := protoutils.MarshalBytes(expected.Prometheus)
	Expect(err).NotTo(HaveOccurred())
	str1 := string(jsn1)
	str2 := string(jsn2)
	Expect(str1).To(Equal(str2))
}
