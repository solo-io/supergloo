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
	v1alpha2sets2 "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2/sets"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators/trafficshift"
	mock_hostutils "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/hostutils/mocks"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"istio.io/api/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		trafficTargets := v1alpha2sets.NewTrafficTargetSet(
			&discoveryv1alpha2.TrafficTarget{
				Spec: discoveryv1alpha2.TrafficTargetSpec{
					Type: &discoveryv1alpha2.TrafficTargetSpec_KubeService_{
						KubeService: &discoveryv1alpha2.TrafficTargetSpec_KubeService{
							Ref: &v1.ClusterObjectRef{
								Name:        "traffic-shift",
								Namespace:   "namespace",
								ClusterName: "cluster",
							},
							Ports: []*discoveryv1alpha2.TrafficTargetSpec_KubeService_KubeServicePort{
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
		trafficShiftDecorator = trafficshift.NewTrafficShiftDecorator(mockClusterDomainRegistry, trafficTargets, nil)
		originalService := &discoveryv1alpha2.TrafficTarget{
			Spec: discoveryv1alpha2.TrafficTargetSpec{
				Type: &discoveryv1alpha2.TrafficTargetSpec_KubeService_{
					KubeService: &discoveryv1alpha2.TrafficTargetSpec_KubeService{
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
		appliedPolicy := &discoveryv1alpha2.TrafficTargetStatus_AppliedTrafficPolicy{
			Spec: &v1alpha2.TrafficPolicySpec{
				TrafficShift: &v1alpha2.TrafficPolicySpec_MultiDestination{
					Destinations: []*v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination{
						{
							DestinationType: &v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination_KubeService{
								KubeService: &v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination_KubeDestination{
									Name:        "traffic-shift",
									Namespace:   "namespace",
									ClusterName: "cluster",
									Port:        9080,
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
					ClusterName: appliedPolicy.Spec.TrafficShift.Destinations[0].GetKubeService().ClusterName,
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
		trafficTargets := v1alpha2sets.NewTrafficTargetSet(
			&discoveryv1alpha2.TrafficTarget{
				Spec: discoveryv1alpha2.TrafficTargetSpec{
					Type: &discoveryv1alpha2.TrafficTargetSpec_KubeService_{
						KubeService: &discoveryv1alpha2.TrafficTargetSpec_KubeService{
							Ref: &v1.ClusterObjectRef{
								Name:        "traffic-shift",
								Namespace:   "namespace",
								ClusterName: "cluster",
							},
							Ports: []*discoveryv1alpha2.TrafficTargetSpec_KubeService_KubeServicePort{
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
		trafficShiftDecorator = trafficshift.NewTrafficShiftDecorator(mockClusterDomainRegistry, trafficTargets, nil)
		originalService := &discoveryv1alpha2.TrafficTarget{
			Spec: discoveryv1alpha2.TrafficTargetSpec{
				Type: &discoveryv1alpha2.TrafficTargetSpec_KubeService_{
					KubeService: &discoveryv1alpha2.TrafficTargetSpec_KubeService{
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
		appliedPolicyMissingPort := &discoveryv1alpha2.TrafficTargetStatus_AppliedTrafficPolicy{
			Spec: &v1alpha2.TrafficPolicySpec{
				TrafficShift: &v1alpha2.TrafficPolicySpec_MultiDestination{
					Destinations: []*v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination{
						{
							DestinationType: &v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination_KubeService{
								KubeService: &v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination_KubeDestination{
									Name:        "traffic-shift",
									Namespace:   "namespace",
									ClusterName: "cluster",
								},
							},
							Weight: 50,
						},
					},
				},
			},
		}
		appliedPolicyNonexistentPort := &discoveryv1alpha2.TrafficTargetStatus_AppliedTrafficPolicy{
			Spec: &v1alpha2.TrafficPolicySpec{
				TrafficShift: &v1alpha2.TrafficPolicySpec_MultiDestination{
					Destinations: []*v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination{
						{
							DestinationType: &v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination_KubeService{
								KubeService: &v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination_KubeDestination{
									Name:        "traffic-shift",
									Namespace:   "namespace",
									ClusterName: "cluster",
									Port:        1,
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
					ClusterName: appliedPolicyMissingPort.Spec.TrafficShift.Destinations[0].GetKubeService().ClusterName,
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
		trafficTargets := v1alpha2sets.NewTrafficTargetSet(
			&discoveryv1alpha2.TrafficTarget{
				Spec: discoveryv1alpha2.TrafficTargetSpec{
					Type: &discoveryv1alpha2.TrafficTargetSpec_KubeService_{
						KubeService: &discoveryv1alpha2.TrafficTargetSpec_KubeService{
							Ref: &v1.ClusterObjectRef{
								Name:        "traffic-shift",
								Namespace:   "namespace",
								ClusterName: "cluster",
							},
							Ports: []*discoveryv1alpha2.TrafficTargetSpec_KubeService_KubeServicePort{
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
		trafficShiftDecorator = trafficshift.NewTrafficShiftDecorator(mockClusterDomainRegistry, trafficTargets, nil)
		originalService := &discoveryv1alpha2.TrafficTarget{
			Spec: discoveryv1alpha2.TrafficTargetSpec{
				Type: &discoveryv1alpha2.TrafficTargetSpec_KubeService_{
					KubeService: &discoveryv1alpha2.TrafficTargetSpec_KubeService{
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
		appliedPolicy := &discoveryv1alpha2.TrafficTargetStatus_AppliedTrafficPolicy{
			Spec: &v1alpha2.TrafficPolicySpec{
				TrafficShift: &v1alpha2.TrafficPolicySpec_MultiDestination{
					Destinations: []*v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination{
						{
							DestinationType: &v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination_KubeService{
								KubeService: &v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination_KubeDestination{
									Name:        "traffic-shift",
									Namespace:   "namespace",
									ClusterName: "cluster",
									Port:        9080,
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
					ClusterName: appliedPolicy.Spec.TrafficShift.Destinations[0].GetKubeService().ClusterName,
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

	It("should decorate traffic shift targeting a FailoverService", func() {
		failoverServices := v1alpha2sets2.NewFailoverServiceSet(
			&v1alpha2.FailoverService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "fs-1",
					Namespace: "fs-ns-1",
				},
				Spec: v1alpha2.FailoverServiceSpec{
					Hostname: "failoverservice.foo.bar.global",
					Port: &v1alpha2.FailoverServiceSpec_Port{
						Number: 9080,
					},
				},
			})
		trafficShiftDecorator = trafficshift.NewTrafficShiftDecorator(mockClusterDomainRegistry, nil, failoverServices)
		appliedPolicy := &discoveryv1alpha2.TrafficTargetStatus_AppliedTrafficPolicy{
			Spec: &v1alpha2.TrafficPolicySpec{
				TrafficShift: &v1alpha2.TrafficPolicySpec_MultiDestination{
					Destinations: []*v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination{
						{
							DestinationType: &v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination_FailoverService{
								FailoverService: &v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination_FailoverServiceDestination{
									Name:      "fs-1",
									Namespace: "fs-ns-1",
									Subset: map[string]string{
										"version": "v1",
									},
								},
							},
							Weight: 50,
						},
					},
				},
			},
		}
		registerField := func(fieldPtr, val interface{}) error {
			return nil
		}
		err := trafficShiftDecorator.ApplyTrafficPolicyToVirtualService(
			appliedPolicy,
			nil,
			output,
			registerField,
		)
		expectedRoute := []*v1alpha3.HTTPRouteDestination{
			{
				Destination: &v1alpha3.Destination{
					Host: "failoverservice.foo.bar.global",
					Port: &v1alpha3.PortSelector{
						Number: 9080,
					},
					Subset: "version-v1",
				},
				Weight: 50,
			},
		}
		Expect(err).ToNot(HaveOccurred())
		Expect(output.Route).To(Equal(expectedRoute))
	})
})
