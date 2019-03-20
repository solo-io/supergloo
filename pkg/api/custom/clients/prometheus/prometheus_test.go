package prometheus_test

import (
	"context"
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	"github.com/solo-io/solo-kit/test/tests/typed"
	"github.com/solo-io/supergloo/pkg/api/custom/clients/prometheus"
	. "github.com/solo-io/supergloo/pkg/api/external/prometheus/v1"
	"github.com/solo-io/supergloo/test/inputs"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

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
				kube   kubernetes.Interface
			)
			BeforeEach(func() {
				namespace = "some-namespace"
				kube = fake.NewSimpleClientset()
				kubeCache, err := cache.NewKubeCoreCache(context.TODO(), kube)
				Expect(err).NotTo(HaveOccurred())

				fact := &factory.KubeConfigMapClientFactory{
					Clientset: kube,
					Cache:     kubeCache,
				}
				client, err = NewPrometheusConfigClient(fact)
				Expect(err).NotTo(HaveOccurred())
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
	err := CreatePrometheusConfigmap(namespace, name, kube)
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
	err := CreatePrometheusConfigmap(namespace, name, kube)
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
}

func CreatePrometheusConfigmap(namespace, name string, kube kubernetes.Interface) error {
	_, err := kube.CoreV1().ConfigMaps(namespace).Create(&v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Annotations: map[string]string{"resource_kind": reflect.TypeOf(&PrometheusConfig{}).String()},
		},
		Data: map[string]string{"prometheus.yml": inputs.BasicPrometheusConfig},
	})
	return err
}
