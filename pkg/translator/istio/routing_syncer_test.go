package istio_test

import (
	"context"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	gloov1 "github.com/solo-io/supergloo/pkg/api/external/gloo/v1"
	"github.com/solo-io/supergloo/pkg/api/external/gloo/v1/plugins/kubernetes"
	"github.com/solo-io/supergloo/pkg/api/external/istio/networking/v1alpha3"
	"github.com/solo-io/supergloo/pkg/api/v1"
	. "github.com/solo-io/supergloo/pkg/translator/istio"
	"k8s.io/client-go/tools/clientcmd"
)

var _ = Describe("RoutingSyncer", func() {
	It("works", func() {
		kubeconfigPath := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		Expect(err).NotTo(HaveOccurred())
		vsClient, err := v1alpha3.NewVirtualServiceClient(&factory.KubeResourceClientFactory{
			Crd:         v1alpha3.VirtualServiceCrd,
			Cfg:         cfg,
			SharedCache: kube.NewKubeCache(),
		})
		Expect(err).NotTo(HaveOccurred())
		err = vsClient.Register()
		Expect(err).NotTo(HaveOccurred())
		vsReconciler := v1alpha3.NewVirtualServiceReconciler(vsClient)
		drClient, err := v1alpha3.NewDestinationRuleClient(&factory.KubeResourceClientFactory{
			Crd:         v1alpha3.DestinationRuleCrd,
			Cfg:         cfg,
			SharedCache: kube.NewKubeCache(),
		})
		Expect(err).NotTo(HaveOccurred())
		err = drClient.Register()
		Expect(err).NotTo(HaveOccurred())
		drReconciler := v1alpha3.NewDestinationRuleReconciler(drClient)
		s := &RoutingSyncer{
			WriteSelector:             map[string]string{"creatd_by": "syncer"},
			WriteNamespace:            "gloo-system",
			VirtualServiceReconciler:  vsReconciler,
			DestinationRuleReconciler: drReconciler,
		}
		err = s.Sync(context.TODO(), &v1.TranslatorSnapshot{
			Meshes: map[string]v1.MeshList{
				"ignored-at-this-point": {{
					TargetMesh: &v1.TargetMesh{
						MeshType: v1.MeshType_ISTIO,
					},
					Routing: &v1.Routing{
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
