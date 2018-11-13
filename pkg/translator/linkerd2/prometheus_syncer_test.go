package linkerd2_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/test/helpers"
	"github.com/solo-io/solo-kit/test/tests/typed"
	prometheusv1 "github.com/solo-io/supergloo/pkg/api/external/prometheus/v1"
	"github.com/solo-io/supergloo/pkg/api/v1"
	. "github.com/solo-io/supergloo/pkg/translator/linkerd2"
	"github.com/solo-io/supergloo/test/utils"
	"k8s.io/client-go/kubernetes"
)

var _ = Describe("PrometheusSyncer", func() {
	test := &typed.KubeConfigMapRcTester{}
	var (
		namespace            string
		kube                 kubernetes.Interface
		prometheusConfigName = "prometheus-config"
	)
	BeforeEach(func() {
		namespace = helpers.RandString(6)
		fact := test.Setup(namespace)
		kube = fact.(*factory.KubeConfigMapClientFactory).Clientset
	})
	AfterEach(func() {
		test.Teardown(namespace)
	})
	It("works", func() {
		err := utils.DeployPrometheusConfigmap(namespace, prometheusConfigName, kube)
		Expect(err).NotTo(HaveOccurred())
		prometheusClient, err := prometheusv1.NewConfigClient(&factory.KubeConfigMapClientFactory{
			Clientset: kube,
		})
		Expect(err).NotTo(HaveOccurred())
		err = prometheusClient.Register()
		Expect(err).NotTo(HaveOccurred())
		s := &PrometheusSyncer{
			PrometheusClient: prometheusClient,
		}
		err = s.Sync(context.TODO(), &v1.TranslatorSnapshot{
			Meshes: map[string]v1.MeshList{
				"ignored-at-this-point": {{
					Observability: &v1.Observability{
						Prometheus: &v1.Prometheus{
							PrometheusConfigMap: &core.ResourceRef{
								Namespace: namespace,
								Name:      prometheusConfigName,
							},
						},
					},
				}},
			},
		})
		Expect(err).NotTo(HaveOccurred())
	})
})
