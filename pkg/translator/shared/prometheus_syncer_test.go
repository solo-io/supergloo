package shared_test

import (
	"context"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/supergloo/pkg/api/external/prometheus"
	"github.com/solo-io/supergloo/pkg/translator/istio"
	"github.com/solo-io/supergloo/pkg/translator/linkerd2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/test/helpers"
	"github.com/solo-io/solo-kit/test/tests/typed"
	prometheusv1 "github.com/solo-io/supergloo/pkg/api/external/prometheus/v1"
	"github.com/solo-io/supergloo/pkg/api/v1"
	. "github.com/solo-io/supergloo/pkg/translator/shared"
	"github.com/solo-io/supergloo/test/utils"
	"k8s.io/client-go/kubernetes"
)

var _ = Describe("PrometheusSyncer", func() {
	test := &typed.KubeConfigMapRcTester{}
	var (
		namespace                string
		kube                     kubernetes.Interface
		prometheusConfigName     = "prometheus"
		prometheusDeploymentName = "prometheus"
	)
	BeforeEach(func() {
		namespace = helpers.RandString(6)
		fact := test.Setup(namespace)
		kube = fact.(*factory.KubeConfigMapClientFactory).Clientset
	})
	AfterEach(func() {
		test.Teardown(namespace)
	})
	for _, test := range []struct {
		meshType      v1.MeshType
		scrapeConfigs []prometheus.ScrapeConfig
	}{
		{
			meshType:      v1.MeshType_ISTIO,
			scrapeConfigs: istio.IstioScrapeConfigs,
		},
		{
			meshType:      v1.MeshType_LINKERD2,
			scrapeConfigs: linkerd2.LinkerdScrapeConfigs,
		},
	} {

		It("works", func() {
			err := utils.DeployPrometheus(namespace, prometheusDeploymentName, prometheusConfigName, 31001, kube)
			Expect(err).NotTo(HaveOccurred())
			err = utils.DeployPrometheusConfigmap(namespace, prometheusConfigName, kube)
			Expect(err).NotTo(HaveOccurred())
			prometheusClient, err := prometheusv1.NewConfigClient(&factory.KubeConfigMapClientFactory{
				Clientset: kube,
			})
			Expect(err).NotTo(HaveOccurred())
			err = prometheusClient.Register()
			Expect(err).NotTo(HaveOccurred())
			s := &PrometheusSyncer{
				PrometheusClient:     prometheusClient,
				Kube:                 kube,
				DesiredScrapeConfigs: test.scrapeConfigs,
				MeshType:             test.meshType,
			}
			original := getPrometheusConfig(prometheusClient, namespace, prometheusConfigName)
			for _, sc := range s.DesiredScrapeConfigs {
				Expect(original.ScrapeConfigs).NotTo(ContainElement(sc))
			}

			err = s.Sync(context.TODO(), &v1.TranslatorSnapshot{
				Meshes: map[string]v1.MeshList{
					"ignored-at-this-point": {{
						TargetMesh: &v1.TargetMesh{
							MeshType: v1.MeshType_LINKERD2,
						},
						Observability: &v1.Observability{
							Prometheus: &v1.Prometheus{
								EnableMetrics: true,
								PrometheusConfigMap: &core.ResourceRef{
									Namespace: namespace,
									Name:      prometheusConfigName,
								},
								PodLabels: map[string]string{
									"app": "prometheus-server",
								},
							},
						},
					}},
				},
			})
			Expect(err).NotTo(HaveOccurred())
			updated := getPrometheusConfig(prometheusClient, namespace, prometheusConfigName)
			for _, sc := range s.DesiredScrapeConfigs {
				Expect(updated.ScrapeConfigs).To(ContainElement(sc))
			}
		})
	}
})

func getPrometheusConfig(promClient prometheusv1.ConfigClient, namespace, name string) *prometheus.PrometheusConfig {
	cfg, err := promClient.Read(namespace, name, clients.ReadOpts{})
	Expect(err).NotTo(HaveOccurred())
	promCfg, err := prometheus.ConfigFromResource(cfg)
	Expect(err).NotTo(HaveOccurred())
	return promCfg
}
