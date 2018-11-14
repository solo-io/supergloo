package linkerd2_test

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	gloov1 "github.com/solo-io/supergloo/pkg/api/external/gloo/v1"
	"github.com/solo-io/supergloo/pkg/api/external/gloo/v1/plugins/kubernetes"
	prometheusv1 "github.com/solo-io/supergloo/pkg/api/external/prometheus/v1"
	"github.com/solo-io/supergloo/pkg/api/v1"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"

	. "github.com/solo-io/supergloo/pkg/translator/linkerd2"
)

var _ = Describe("PrometheusSyncer", func() {
	It("works", func() {
		kubeconfigPath := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		Expect(err).NotTo(HaveOccurred())
		prometheusClient, err := prometheusv1.NewConfigClient(&factory.KubeResourceClientFactory{
			Crd:         prometheusv1.ConfigCrd,
			Cfg:         cfg,
			SharedCache: kube.NewKubeCache(),
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
								Namespace:,
								Name:,
							},
						},
						DestinationRules: []*v1.DestinationRule{
							{
								Destination: &gloov1.Destination{
									Upstream: core.ResourceRef{
										Name:      "default-reviews-9080",
										Namespace: "gloo-system",
									},
								},
								MeshHttpRules: []*v1.HTTPRule{
									{
										Route: []*v1.HTTPRouteDestination{
											{
												AlternateDestination: &gloov1.Destination{
													Upstream: core.ResourceRef{
														Name:      "default-reviews-9080",
														Namespace: "gloo-system",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				}},
			},
			Upstreams: map[string]gloov1.UpstreamList{
				"also gets ignored": {
					{
						Metadata: core.Metadata{
							Name:      "default-reviews-9080",
							Namespace: "gloo-system",
						},
						UpstreamSpec: &gloov1.UpstreamSpec{
							UpstreamType: &gloov1.UpstreamSpec_Kube{
								Kube: &kubernetes.UpstreamSpec{
									ServiceName:      "reviews",
									ServiceNamespace: "default",
									ServicePort:      9080,
								},
							},
						},
					},
				},
			},
		})
		Expect(err).NotTo(HaveOccurred())
	})
})
