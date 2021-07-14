package trafficshift_test

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	v1sets "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1/sets"
	enterprisenetworkingv1beta1 "github.com/solo-io/gloo-mesh/pkg/api/networking.enterprise.mesh.gloo.solo.io/v1beta1"
	v1beta1sets "github.com/solo-io/gloo-mesh/pkg/api/networking.enterprise.mesh.gloo.solo.io/v1beta1/sets"
	networkingv1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators/trafficshift"
	mock_hostutils "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/hostutils/mocks"
	"github.com/solo-io/go-utils/testutils"
	skv2corev1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
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

	It("should decorate traffic shift with selected port", func() {
		destinations := v1sets.NewDestinationSet(
			&discoveryv1.Destination{
				Spec: discoveryv1.DestinationSpec{
					Type: &discoveryv1.DestinationSpec_KubeService_{
						KubeService: &discoveryv1.DestinationSpec_KubeService{
							Ref: &skv2corev1.ClusterObjectRef{
								Name:        "traffic-shift",
								Namespace:   "namespace",
								ClusterName: "cluster",
							},
							Ports: []*discoveryv1.DestinationSpec_KubeService_KubeServicePort{
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
		trafficShiftDecorator = trafficshift.NewTrafficShiftDecorator(mockClusterDomainRegistry, destinations, nil)
		originalService := &discoveryv1.Destination{
			Spec: discoveryv1.DestinationSpec{
				Type: &discoveryv1.DestinationSpec_KubeService_{
					KubeService: &discoveryv1.DestinationSpec_KubeService{
						Ref: &skv2corev1.ClusterObjectRef{
							ClusterName: "local-cluster",
						},
					},
				},
			},
		}
		registerField := func(fieldPtr, val interface{}) error {
			return nil
		}
		appliedPolicy := &networkingv1.AppliedTrafficPolicy{
			Spec: &networkingv1.TrafficPolicySpec{
				Policy: &networkingv1.TrafficPolicySpec_Policy{
					TrafficShift: &networkingv1.TrafficPolicySpec_Policy_MultiDestination{
						Destinations: []*networkingv1.WeightedDestination{
							{
								DestinationType: &networkingv1.WeightedDestination_KubeService{
									KubeService: &networkingv1.WeightedDestination_KubeDestination{
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
			},
		}

		trafficShiftHostname := "name.namespace.svc.cluster.local"
		mockClusterDomainRegistry.
			EXPECT().
			GetDestinationFQDN(originalService.Spec.GetKubeService().Ref.ClusterName,
				&skv2corev1.ClusterObjectRef{
					Name:        appliedPolicy.Spec.GetPolicy().GetTrafficShift().Destinations[0].GetKubeService().Name,
					Namespace:   appliedPolicy.Spec.GetPolicy().GetTrafficShift().Destinations[0].GetKubeService().Namespace,
					ClusterName: appliedPolicy.Spec.GetPolicy().GetTrafficShift().Destinations[0].GetKubeService().ClusterName,
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
		err := trafficShiftDecorator.ApplyTrafficPolicyToVirtualService(appliedPolicy, originalService, nil, output, registerField)

		Expect(err).ToNot(HaveOccurred())
		Expect(output.Route).To(Equal(expectedHTTPDestinations))
	})

	It("should decorate traffic shift for VirtualDestination", func() {
		virtualDestinations := v1beta1sets.NewVirtualDestinationSet(&enterprisenetworkingv1beta1.VirtualDestination{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "virtual-destination",
				Namespace: "virtual-destination-ns",
			},
			Spec: enterprisenetworkingv1beta1.VirtualDestinationSpec{
				Hostname: "vdest.hostname",
				Port: &enterprisenetworkingv1beta1.VirtualDestinationSpec_Port{
					Number:   1000,
					Protocol: "TCP",
					TargetPort: &enterprisenetworkingv1beta1.VirtualDestinationSpec_Port_TargetNumber{
						TargetNumber: 1000,
					},
				},
			},
		})

		trafficShiftDecorator = trafficshift.NewTrafficShiftDecorator(mockClusterDomainRegistry, nil, virtualDestinations)

		originalService := &discoveryv1.Destination{
			Spec: discoveryv1.DestinationSpec{
				Type: &discoveryv1.DestinationSpec_KubeService_{
					KubeService: &discoveryv1.DestinationSpec_KubeService{
						Ref: &skv2corev1.ClusterObjectRef{
							ClusterName: "local-cluster",
						},
					},
				},
			},
		}
		registerField := func(fieldPtr, val interface{}) error {
			return nil
		}
		appliedPolicy := &networkingv1.AppliedTrafficPolicy{
			Spec: &networkingv1.TrafficPolicySpec{
				Policy: &networkingv1.TrafficPolicySpec_Policy{
					TrafficShift: &networkingv1.TrafficPolicySpec_Policy_MultiDestination{
						Destinations: []*networkingv1.WeightedDestination{
							{
								DestinationType: &networkingv1.WeightedDestination_VirtualDestination_{
									VirtualDestination: &networkingv1.WeightedDestination_VirtualDestination{
										Name:      "virtual-destination",
										Namespace: "virtual-destination-ns",
									},
								},
							},
						},
					},
				},
			},
		}

		expectedHTTPDestinations := []*v1alpha3.HTTPRouteDestination{
			{
				Destination: &v1alpha3.Destination{
					Host: "vdest.hostname",
					Port: &v1alpha3.PortSelector{
						Number: 1000,
					},
				},
			},
		}
		err := trafficShiftDecorator.ApplyTrafficPolicyToVirtualService(appliedPolicy, originalService, nil, output, registerField)

		Expect(err).ToNot(HaveOccurred())
		Expect(output.Route).To(Equal(expectedHTTPDestinations))
	})

	It("should decorate traffic shift for federated Destination with selected port", func() {
		destinations := v1sets.NewDestinationSet(
			&discoveryv1.Destination{
				Spec: discoveryv1.DestinationSpec{
					Type: &discoveryv1.DestinationSpec_KubeService_{
						KubeService: &discoveryv1.DestinationSpec_KubeService{
							Ref: &skv2corev1.ClusterObjectRef{
								Name:        "traffic-shift",
								Namespace:   "namespace",
								ClusterName: "cluster",
							},
							Ports: []*discoveryv1.DestinationSpec_KubeService_KubeServicePort{
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
		trafficShiftDecorator = trafficshift.NewTrafficShiftDecorator(mockClusterDomainRegistry, destinations, nil)
		originalService := &discoveryv1.Destination{
			Spec: discoveryv1.DestinationSpec{
				Type: &discoveryv1.DestinationSpec_KubeService_{
					KubeService: &discoveryv1.DestinationSpec_KubeService{
						Ref: &skv2corev1.ClusterObjectRef{
							ClusterName: "local-cluster",
						},
					},
				},
			},
		}
		registerField := func(fieldPtr, val interface{}) error {
			return nil
		}
		appliedPolicy := &networkingv1.AppliedTrafficPolicy{
			Spec: &networkingv1.TrafficPolicySpec{
				Policy: &networkingv1.TrafficPolicySpec_Policy{
					TrafficShift: &networkingv1.TrafficPolicySpec_Policy_MultiDestination{
						Destinations: []*networkingv1.WeightedDestination{
							{
								DestinationType: &networkingv1.WeightedDestination_KubeService{
									KubeService: &networkingv1.WeightedDestination_KubeDestination{
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
			},
		}

		sourceMeshInstallation := &discoveryv1.MeshInstallation{
			Cluster: "federated-cluster-name",
		}
		globalTrafficShiftHostname := "name.namespace.svc.local-cluster.global"
		mockClusterDomainRegistry.
			EXPECT().
			GetDestinationFQDN(
				sourceMeshInstallation.GetCluster(),
				&skv2corev1.ClusterObjectRef{
					Name:        appliedPolicy.Spec.GetPolicy().GetTrafficShift().Destinations[0].GetKubeService().Name,
					Namespace:   appliedPolicy.Spec.GetPolicy().GetTrafficShift().Destinations[0].GetKubeService().Namespace,
					ClusterName: appliedPolicy.Spec.GetPolicy().GetTrafficShift().Destinations[0].GetKubeService().ClusterName,
				}).
			Return(globalTrafficShiftHostname)

		expectedHTTPDestinations := []*v1alpha3.HTTPRouteDestination{
			{
				Destination: &v1alpha3.Destination{
					Host: globalTrafficShiftHostname,
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
			sourceMeshInstallation,
			output,
			registerField,
		)

		Expect(err).ToNot(HaveOccurred())
		Expect(output.Route).To(Equal(expectedHTTPDestinations))
	})

	It("should throw error if traffic shift destination has multiple ports but TrafficPolicy does not specify which port", func() {
		destinations := v1sets.NewDestinationSet(
			&discoveryv1.Destination{
				Spec: discoveryv1.DestinationSpec{
					Type: &discoveryv1.DestinationSpec_KubeService_{
						KubeService: &discoveryv1.DestinationSpec_KubeService{
							Ref: &skv2corev1.ClusterObjectRef{
								Name:        "traffic-shift",
								Namespace:   "namespace",
								ClusterName: "cluster",
							},
							Ports: []*discoveryv1.DestinationSpec_KubeService_KubeServicePort{
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
		trafficShiftDecorator = trafficshift.NewTrafficShiftDecorator(mockClusterDomainRegistry, destinations, nil)
		originalService := &discoveryv1.Destination{
			Spec: discoveryv1.DestinationSpec{
				Type: &discoveryv1.DestinationSpec_KubeService_{
					KubeService: &discoveryv1.DestinationSpec_KubeService{
						Ref: &skv2corev1.ClusterObjectRef{
							ClusterName: "local-cluster",
						},
					},
				},
			},
		}
		registerField := func(fieldPtr, val interface{}) error {
			return nil
		}
		appliedPolicyMissingPort := &networkingv1.AppliedTrafficPolicy{
			Spec: &networkingv1.TrafficPolicySpec{
				Policy: &networkingv1.TrafficPolicySpec_Policy{
					TrafficShift: &networkingv1.TrafficPolicySpec_Policy_MultiDestination{
						Destinations: []*networkingv1.WeightedDestination{
							{
								DestinationType: &networkingv1.WeightedDestination_KubeService{
									KubeService: &networkingv1.WeightedDestination_KubeDestination{
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
			},
		}
		appliedPolicyNonexistentPort := &networkingv1.AppliedTrafficPolicy{
			Spec: &networkingv1.TrafficPolicySpec{
				Policy: &networkingv1.TrafficPolicySpec_Policy{
					TrafficShift: &networkingv1.TrafficPolicySpec_Policy_MultiDestination{
						Destinations: []*networkingv1.WeightedDestination{
							{
								DestinationType: &networkingv1.WeightedDestination_KubeService{
									KubeService: &networkingv1.WeightedDestination_KubeDestination{
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
			},
		}

		trafficShiftHostname := "name.namespace.svc.cluster.local"
		mockClusterDomainRegistry.
			EXPECT().
			GetDestinationFQDN(originalService.Spec.GetKubeService().Ref.ClusterName,
				&skv2corev1.ClusterObjectRef{
					Name:        appliedPolicyMissingPort.Spec.GetPolicy().GetTrafficShift().Destinations[0].GetKubeService().Name,
					Namespace:   appliedPolicyMissingPort.Spec.GetPolicy().GetTrafficShift().Destinations[0].GetKubeService().Namespace,
					ClusterName: appliedPolicyMissingPort.Spec.GetPolicy().GetTrafficShift().Destinations[0].GetKubeService().ClusterName,
				}).
			Return(trafficShiftHostname).Times(2)

		noPortError := trafficShiftDecorator.ApplyTrafficPolicyToVirtualService(appliedPolicyMissingPort, originalService, nil, output, registerField)
		Expect(noPortError.Error()).To(ContainSubstring("must provide port for traffic shift destination service"))

		nonexistentPort := trafficShiftDecorator.ApplyTrafficPolicyToVirtualService(appliedPolicyNonexistentPort, originalService, nil, output, registerField)
		Expect(nonexistentPort.Error()).To(ContainSubstring("does not exist for traffic shift destination service"))
	})

	It("should not decorate traffic shift if error during field registration", func() {
		destinations := v1sets.NewDestinationSet(
			&discoveryv1.Destination{
				Spec: discoveryv1.DestinationSpec{
					Type: &discoveryv1.DestinationSpec_KubeService_{
						KubeService: &discoveryv1.DestinationSpec_KubeService{
							Ref: &skv2corev1.ClusterObjectRef{
								Name:        "traffic-shift",
								Namespace:   "namespace",
								ClusterName: "cluster",
							},
							Ports: []*discoveryv1.DestinationSpec_KubeService_KubeServicePort{
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
		trafficShiftDecorator = trafficshift.NewTrafficShiftDecorator(mockClusterDomainRegistry, destinations, nil)
		originalService := &discoveryv1.Destination{
			Spec: discoveryv1.DestinationSpec{
				Type: &discoveryv1.DestinationSpec_KubeService_{
					KubeService: &discoveryv1.DestinationSpec_KubeService{
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
		appliedPolicy := &networkingv1.AppliedTrafficPolicy{
			Spec: &networkingv1.TrafficPolicySpec{
				Policy: &networkingv1.TrafficPolicySpec_Policy{
					TrafficShift: &networkingv1.TrafficPolicySpec_Policy_MultiDestination{
						Destinations: []*networkingv1.WeightedDestination{
							{
								DestinationType: &networkingv1.WeightedDestination_KubeService{
									KubeService: &networkingv1.WeightedDestination_KubeDestination{
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
			},
		}

		trafficShiftHostname := "name.namespace.svc.cluster.local"
		mockClusterDomainRegistry.
			EXPECT().
			GetDestinationFQDN(originalService.Spec.GetKubeService().Ref.ClusterName,
				&skv2corev1.ClusterObjectRef{
					Name:        appliedPolicy.Spec.GetPolicy().GetTrafficShift().Destinations[0].GetKubeService().Name,
					Namespace:   appliedPolicy.Spec.GetPolicy().GetTrafficShift().Destinations[0].GetKubeService().Namespace,
					ClusterName: appliedPolicy.Spec.GetPolicy().GetTrafficShift().Destinations[0].GetKubeService().ClusterName,
				}).
			Return(trafficShiftHostname)

		err := trafficShiftDecorator.ApplyTrafficPolicyToVirtualService(appliedPolicy, originalService, nil, output, registerField)

		Expect(err).To(testutils.HaveInErrorChain(testErr))
	})
})
