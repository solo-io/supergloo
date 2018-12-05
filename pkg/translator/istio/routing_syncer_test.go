package istio_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/memory"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	gloov1 "github.com/solo-io/supergloo/pkg/api/external/gloo/v1"
	"github.com/solo-io/supergloo/pkg/api/external/gloo/v1/plugins/kubernetes"
	"github.com/solo-io/supergloo/pkg/api/external/istio/networking/v1alpha3"
	"github.com/solo-io/supergloo/pkg/api/v1"
	. "github.com/solo-io/supergloo/pkg/translator/istio"
)

var _ = Describe("RoutingSyncer", func() {
	namespace := "test"
	It("works", func() {
		memory := &factory.MemoryResourceClientFactory{
			Cache: memory.NewInMemoryResourceCache(),
		}
		vsClient, err := v1alpha3.NewVirtualServiceClient(memory)
		Expect(err).NotTo(HaveOccurred())
		err = vsClient.Register()
		Expect(err).NotTo(HaveOccurred())
		vsReconciler := v1alpha3.NewVirtualServiceReconciler(vsClient)
		drClient, err := v1alpha3.NewDestinationRuleClient(memory)
		Expect(err).NotTo(HaveOccurred())
		err = drClient.Register()
		Expect(err).NotTo(HaveOccurred())
		drReconciler := v1alpha3.NewDestinationRuleReconciler(drClient)
		s := NewMeshRoutingSyncer([]string{namespace},
			nil,
			drReconciler,
			vsReconciler,
			nil,
		)
		err = s.Sync(context.TODO(), &v1.TranslatorSnapshot{
			Meshes: map[string]v1.MeshList{
				"ignored-at-this-point": {{
					Metadata: core.Metadata{Name: "name", Namespace: namespace},
					MeshType: &v1.Mesh_Istio{
						Istio: &v1.Istio{
							WatchNamespaces: []string{namespace},
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
							Namespace: namespace,
						},
						UpstreamSpec: &gloov1.UpstreamSpec{
							UpstreamType: &gloov1.UpstreamSpec_Kube{
								Kube: &kubernetes.UpstreamSpec{
									ServiceName:      "reviews",
									ServiceNamespace: "default",
									ServicePort:      9080,
									Selector:         map[string]string{"app": "reviews", "version": "v2"},
								},
							},
						},
					},
				},
			},
			Routingrules: map[string]v1.RoutingRuleList{
				"": {
					{
						Metadata:   core.Metadata{Name: "fault", Namespace: namespace},
						TargetMesh: &core.ResourceRef{Name: "name", Namespace: namespace},
						FaultInjection: &v1alpha3.HTTPFaultInjection{
							Abort: &v1alpha3.HTTPFaultInjection_Abort{
								ErrorType: &v1alpha3.HTTPFaultInjection_Abort_HttpStatus{
									HttpStatus: 566,
								},
								Percent: 100,
							},
						},
					},
					{
						Metadata:   core.Metadata{Name: "trafficshifting", Namespace: namespace},
						TargetMesh: &core.ResourceRef{Name: "name", Namespace: namespace},
						TrafficShifting: &v1.TrafficShifting{
							Destinations: []*v1.WeightedDestination{
								{
									Upstream: &core.ResourceRef{
										Name:      "default-reviews-9080-version-v2",
										Namespace: namespace,
									},
									Weight: 100,
								},
							},
						},
					},
				},
			},
		})
		Expect(err).NotTo(HaveOccurred())

		dr, err := drClient.List(namespace, clients.ListOpts{})
		Expect(err).NotTo(HaveOccurred())
		Expect(dr).To(HaveLen(1))
		Expect(dr[0]).To(Equal(&v1alpha3.DestinationRule{
			Metadata: core.Metadata{
				Name:            "name-reviews-default-svc-cluster-local",
				Namespace:       "test",
				ResourceVersion: "2",
				Labels: map[string]string{
					"reconciler.solo.io": "supergloo.istio.routing",
				},
				Annotations: map[string]string{
					"created_by": "supergloo",
				},
			},
			Host: "reviews.default.svc.cluster.local",
			TrafficPolicy: &v1alpha3.TrafficPolicy{
				Tls: &v1alpha3.TLSSettings{
					Mode: v1alpha3.TLSSettings_ISTIO_MUTUAL,
				},
			},
			Subsets: []*v1alpha3.Subset{
				{
					Name:   "app-reviews",
					Labels: map[string]string{"app": "reviews"},
				},
				{
					Name:   "app-reviews-version-v2",
					Labels: map[string]string{"app": "reviews", "version": "v2"},
				},
			},
		}))

		Expect(err).NotTo(HaveOccurred())
		vs, err := vsClient.List(namespace, clients.ListOpts{})
		Expect(err).NotTo(HaveOccurred())
		Expect(vs).To(HaveLen(1))
		Expect(vs[0]).To(Equal(&v1alpha3.VirtualService{
			Metadata: core.Metadata{
				Name:            "name-reviews-default-svc-cluster-local",
				Namespace:       "test",
				ResourceVersion: "2",
				Labels: map[string]string{
					"reconciler.solo.io": "supergloo.istio.routing",
				},
				Annotations: map[string]string{
					"created_by": "supergloo",
				},
			},
			Hosts: []string{
				"reviews.default.svc.cluster.local",
			},
			Gateways: []string{"mesh"},
			Http: []*v1alpha3.HTTPRoute{
				{
					Match: []*v1alpha3.HTTPMatchRequest{
						{
							Uri: &v1alpha3.StringMatch{MatchType: &v1alpha3.StringMatch_Prefix{Prefix: "/"}},
						},
					},
					Route: []*v1alpha3.DestinationWeight{
						{Destination: &v1alpha3.Destination{
							Host:   "reviews.default.svc.cluster.local",
							Subset: "app-reviews-version-v2",
							Port:   &v1alpha3.PortSelector{Port: &v1alpha3.PortSelector_Number{Number: 9080}},
						}},
					},
					Fault: &v1alpha3.HTTPFaultInjection{Abort: &v1alpha3.HTTPFaultInjection_Abort{
						Percent:   100,
						ErrorType: &v1alpha3.HTTPFaultInjection_Abort_HttpStatus{HttpStatus: 566}}},
				},
			},
		}))

	})
})
