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
		s := &MeshRoutingSyncer{
			WriteSelector:             map[string]string{"creatd_by": "syncer"},
			WriteNamespace:            "gloo-system",
			VirtualServiceReconciler:  vsReconciler,
			DestinationRuleReconciler: drReconciler,
		}
		err = s.Sync(context.TODO(), &v1.TranslatorSnapshot{
			Meshes: map[string]v1.MeshList{
				"ignored-at-this-point": {{
					Metadata: core.Metadata{Name: "name", Namespace: "namespace"},
					MeshType: &v1.Mesh_Istio{
						Istio: &v1.Istio{
							WatchNamespaces: []string{"namespace"},
						},
					},
					Encryption: &v1.Encryption{TlsEnabled: true},
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
									Selector:         map[string]string{"app": "reviews"},
								},
							},
						},
					},
					{
						Metadata: core.Metadata{
							Name:      "default-reviews-9080-version-v2",
							Namespace: "gloo-system",
						},
						UpstreamSpec: &gloov1.UpstreamSpec{
							UpstreamType: &gloov1.UpstreamSpec_Kube{
								Kube: &kubernetes.UpstreamSpec{
									ServiceName:      "reviews",
									ServiceNamespace: "default",
									ServicePort:      9080,
									Selector:         map[string]string{"app": "reviews", "version": "v1"},
								},
							},
						},
					},
				},
			},
			Routingrules: map[string]v1.RoutingRuleList{
				"hi": {{
					Metadata:   core.Metadata{Name: "name", Namespace: "namespace"},
					TargetMesh: &core.ResourceRef{Name: "name", Namespace: "namespace"},
					FaultInjection: &v1alpha3.HTTPFaultInjection{
						Abort: &v1alpha3.HTTPFaultInjection_Abort{
							ErrorType: &v1alpha3.HTTPFaultInjection_Abort_HttpStatus{
								HttpStatus: 566,
							},
							Percentage: &v1alpha3.Percent{
								Value: 100,
							},
						},
					},
				}},
			},
		})
		Expect(err).NotTo(HaveOccurred())
	})
})
