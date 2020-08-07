package trafficshift_test

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/testutils"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	v1alpha2sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2/sets"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators/trafficshift"
	mock_hostutils "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/hostutils/mocks"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"istio.io/api/networking/v1alpha3"
)

var _ = Describe("TrafficShiftDecorator", func() {
	var (
		ctrl                      *gomock.Controller
		mockClusterDomainRegistry *mock_hostutils.MockClusterDomainRegistry
		trafficShiftDecorator     decorators.TrafficPolicyVirtualServiceDecorator
		output                    *v1alpha3.HTTPRoute
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockClusterDomainRegistry = mock_hostutils.NewMockClusterDomainRegistry(ctrl)
		output = &v1alpha3.HTTPRoute{}
	})

	It("should decorate mirror with selected port", func() {
		meshServices := v1alpha2sets.NewMeshServiceSet(
			&discoveryv1alpha2.MeshService{
				Spec: discoveryv1alpha2.MeshServiceSpec{
					Type: &discoveryv1alpha2.MeshServiceSpec_KubeService_{
						KubeService: &discoveryv1alpha2.MeshServiceSpec_KubeService{
							Ref: &v1.ClusterObjectRef{
								Name:        "traffic-shift",
								Namespace:   "namespace",
								ClusterName: "cluster",
							},
							Ports: []*discoveryv1alpha2.MeshServiceSpec_KubeService_KubeServicePort{
								{
									Port:     9080,
									Name:     "http1",
									Protocol: "http",
								},
								{
									Port:     9081,
									Name:     "http2",
									Protocol: "http",
								},
							},
						},
					},
				},
			})
		trafficShiftDecorator = trafficshift.NewTrafficShiftDecorator(mockClusterDomainRegistry, meshServices)
		originalService := &discoveryv1alpha2.MeshService{
			Spec: discoveryv1alpha2.MeshServiceSpec{
				Type: &discoveryv1alpha2.MeshServiceSpec_KubeService_{
					KubeService: &discoveryv1alpha2.MeshServiceSpec_KubeService{
						Ref: &v1.ClusterObjectRef{
							ClusterName: "local-cluster",
						},
					},
				},
			},
		}
		registerField := func(fieldPtr, val interface{}) error {
			return nil
		}
		appliedPolicy := &discoveryv1alpha2.MeshServiceStatus_AppliedTrafficPolicy{
			Spec: &v1alpha2.TrafficPolicySpec{
				TrafficShift: &v1alpha2.TrafficPolicySpec_MultiDestination{
					Destinations: []*v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination{
						{
							DestinationType: &v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination_KubeService{
								KubeService: &v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination_KubeDestination{
									Name:      "traffic-shift",
									Namespace: "namespace",
									Cluster:   "cluster",
									Port:      9080,
								},
							},
							Weight: 50,
						},
					},
				},
			},
		}

		trafficShiftHostname := "name.namespace.svc.cluster.local"
		mockClusterDomainRegistry.
			EXPECT().
			GetDestinationServiceFQDN(originalService.Spec.GetKubeService().Ref.ClusterName,
				&v1.ClusterObjectRef{
					Name:        appliedPolicy.Spec.TrafficShift.Destinations[0].GetKubeService().Name,
					Namespace:   appliedPolicy.Spec.TrafficShift.Destinations[0].GetKubeService().Namespace,
					ClusterName: appliedPolicy.Spec.TrafficShift.Destinations[0].GetKubeService().Cluster,
				}).
			Return(trafficShiftHostname)

		expectedHTTPDestinations := []*v1alpha3.HTTPRouteDestination{
			{
				Destination: &v1alpha3.Destination{
					Host: trafficShiftHostname,
					Port: &v1alpha3.PortSelector{
						Number: 9080,
					},
				},
				Weight: 50,
			},
		}
		err := trafficShiftDecorator.ApplyTrafficPolicyToVirtualService(
			appliedPolicy,
			originalService,
			output,
			registerField,
		)

		Expect(err).ToNot(HaveOccurred())
		Expect(output.Route).To(Equal(expectedHTTPDestinations))
	})

	It("should throw error if traffic shift destination has multiple ports but traffic policy does not specify which port", func() {
		meshServices := v1alpha2sets.NewMeshServiceSet(
			&discoveryv1alpha2.MeshService{
				Spec: discoveryv1alpha2.MeshServiceSpec{
					Type: &discoveryv1alpha2.MeshServiceSpec_KubeService_{
						KubeService: &discoveryv1alpha2.MeshServiceSpec_KubeService{
							Ref: &v1.ClusterObjectRef{
								Name:        "traffic-shift",
								Namespace:   "namespace",
								ClusterName: "cluster",
							},
							Ports: []*discoveryv1alpha2.MeshServiceSpec_KubeService_KubeServicePort{
								{
									Port:     9080,
									Name:     "http1",
									Protocol: "http",
								},
								{
									Port:     9081,
									Name:     "http2",
									Protocol: "http",
								},
							},
						},
					},
				},
			})
		trafficShiftDecorator = trafficshift.NewTrafficShiftDecorator(mockClusterDomainRegistry, meshServices)
		originalService := &discoveryv1alpha2.MeshService{
			Spec: discoveryv1alpha2.MeshServiceSpec{
				Type: &discoveryv1alpha2.MeshServiceSpec_KubeService_{
					KubeService: &discoveryv1alpha2.MeshServiceSpec_KubeService{
						Ref: &v1.ClusterObjectRef{
							ClusterName: "local-cluster",
						},
					},
				},
			},
		}
		registerField := func(fieldPtr, val interface{}) error {
			return nil
		}
		appliedPolicyMissingPort := &discoveryv1alpha2.MeshServiceStatus_AppliedTrafficPolicy{
			Spec: &v1alpha2.TrafficPolicySpec{
				TrafficShift: &v1alpha2.TrafficPolicySpec_MultiDestination{
					Destinations: []*v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination{
						{
							DestinationType: &v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination_KubeService{
								KubeService: &v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination_KubeDestination{
									Name:      "traffic-shift",
									Namespace: "namespace",
									Cluster:   "cluster",
								},
							},
							Weight: 50,
						},
					},
				},
			},
		}
		appliedPolicyNonexistentPort := &discoveryv1alpha2.MeshServiceStatus_AppliedTrafficPolicy{
			Spec: &v1alpha2.TrafficPolicySpec{
				TrafficShift: &v1alpha2.TrafficPolicySpec_MultiDestination{
					Destinations: []*v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination{
						{
							DestinationType: &v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination_KubeService{
								KubeService: &v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination_KubeDestination{
									Name:      "traffic-shift",
									Namespace: "namespace",
									Cluster:   "cluster",
									Port:      1,
								},
							},
							Weight: 50,
						},
					},
				},
			},
		}

		trafficShiftHostname := "name.namespace.svc.cluster.local"
		mockClusterDomainRegistry.
			EXPECT().
			GetDestinationServiceFQDN(originalService.Spec.GetKubeService().Ref.ClusterName,
				&v1.ClusterObjectRef{
					Name:        appliedPolicyMissingPort.Spec.TrafficShift.Destinations[0].GetKubeService().Name,
					Namespace:   appliedPolicyMissingPort.Spec.TrafficShift.Destinations[0].GetKubeService().Namespace,
					ClusterName: appliedPolicyMissingPort.Spec.TrafficShift.Destinations[0].GetKubeService().Cluster,
				}).
			Return(trafficShiftHostname).Times(2)

		noPortError := trafficShiftDecorator.ApplyTrafficPolicyToVirtualService(
			appliedPolicyMissingPort,
			originalService,
			output,
			registerField,
		)
		Expect(noPortError.Error()).To(ContainSubstring("must provide port for traffic shift destination service"))

		nonexistentPort := trafficShiftDecorator.ApplyTrafficPolicyToVirtualService(
			appliedPolicyNonexistentPort,
			originalService,
			output,
			registerField,
		)
		Expect(nonexistentPort.Error()).To(ContainSubstring("does not exist for traffic shift destination service"))
	})

	It("should not decorate traffic shift if error during field registration", func() {
		meshServices := v1alpha2sets.NewMeshServiceSet(
			&discoveryv1alpha2.MeshService{
				Spec: discoveryv1alpha2.MeshServiceSpec{
					Type: &discoveryv1alpha2.MeshServiceSpec_KubeService_{
						KubeService: &discoveryv1alpha2.MeshServiceSpec_KubeService{
							Ref: &v1.ClusterObjectRef{
								Name:        "traffic-shift",
								Namespace:   "namespace",
								ClusterName: "cluster",
							},
							Ports: []*discoveryv1alpha2.MeshServiceSpec_KubeService_KubeServicePort{
								{
									Port:     9080,
									Name:     "http1",
									Protocol: "http",
								},
								{
									Port:     9081,
									Name:     "http2",
									Protocol: "http",
								},
							},
						},
					},
				},
			})
		trafficShiftDecorator = trafficshift.NewTrafficShiftDecorator(mockClusterDomainRegistry, meshServices)
		originalService := &discoveryv1alpha2.MeshService{
			Spec: discoveryv1alpha2.MeshServiceSpec{
				Type: &discoveryv1alpha2.MeshServiceSpec_KubeService_{
					KubeService: &discoveryv1alpha2.MeshServiceSpec_KubeService{
						Ref: &v1.ClusterObjectRef{
							ClusterName: "local-cluster",
						},
					},
				},
			},
		}

		testErr := eris.New("registration error")
		registerField := func(fieldPtr, val interface{}) error {
			return testErr
		}
		appliedPolicy := &discoveryv1alpha2.MeshServiceStatus_AppliedTrafficPolicy{
			Spec: &v1alpha2.TrafficPolicySpec{
				TrafficShift: &v1alpha2.TrafficPolicySpec_MultiDestination{
					Destinations: []*v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination{
						{
							DestinationType: &v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination_KubeService{
								KubeService: &v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination_KubeDestination{
									Name:      "traffic-shift",
									Namespace: "namespace",
									Cluster:   "cluster",
									Port:      9080,
								},
							},
							Weight: 50,
						},
					},
				},
			},
		}

		trafficShiftHostname := "name.namespace.svc.cluster.local"
		mockClusterDomainRegistry.
			EXPECT().
			GetDestinationServiceFQDN(originalService.Spec.GetKubeService().Ref.ClusterName,
				&v1.ClusterObjectRef{
					Name:        appliedPolicy.Spec.TrafficShift.Destinations[0].GetKubeService().Name,
					Namespace:   appliedPolicy.Spec.TrafficShift.Destinations[0].GetKubeService().Namespace,
					ClusterName: appliedPolicy.Spec.TrafficShift.Destinations[0].GetKubeService().Cluster,
				}).
			Return(trafficShiftHostname)

		err := trafficShiftDecorator.ApplyTrafficPolicyToVirtualService(
			appliedPolicy,
			originalService,
			output,
			registerField,
		)

		Expect(err).To(testutils.HaveInErrorChain(testErr))
	})
})
