package destinationrule_test

import (
	"context"
	"sort"

	"github.com/gogo/protobuf/types"
	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/proto"
	duration "github.com/golang/protobuf/ptypes/duration"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	v1alpha3sets "github.com/solo-io/external-apis/pkg/api/istio/networking.istio.io/v1alpha3/sets"
	commonv1 "github.com/solo-io/gloo-mesh/pkg/api/common.mesh.gloo.solo.io/v1"
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	discoveryv1sets "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1/sets"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	v1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	settingsv1 "github.com/solo-io/gloo-mesh/pkg/api/settings.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	mock_reporting "github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting/mocks"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators"
	mock_decorators "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators/mocks"
	mock_trafficpolicy "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators/mocks"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/destination/destinationrule"
	mock_hostutils "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/hostutils/mocks"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/metautils"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/routeutils"
	"github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	skv2corev1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/solo-io/skv2/pkg/ezkube"
	networkingv1alpha3spec "istio.io/api/networking/v1alpha3"
	networkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("DestinationRuleTranslator", func() {
	var (
		ctrl                      *gomock.Controller
		mockClusterDomainRegistry *mock_hostutils.MockClusterDomainRegistry
		mockDecoratorFactory      *mock_decorators.MockFactory
		destinations              discoveryv1sets.DestinationSet
		mockReporter              *mock_reporting.MockReporter
		mockDecorator             *mock_trafficpolicy.MockTrafficPolicyDestinationRuleDecorator
		destinationRuleTranslator destinationrule.Translator
		in                        input.LocalSnapshot
		ctx                       = context.TODO()
		settings                  = &settingsv1.Settings{
			ObjectMeta: metav1.ObjectMeta{
				Name:      defaults.DefaultSettingsName,
				Namespace: defaults.DefaultPodNamespace,
			},
		}
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockClusterDomainRegistry = mock_hostutils.NewMockClusterDomainRegistry(ctrl)
		mockDecoratorFactory = mock_decorators.NewMockFactory(ctrl)
		destinations = discoveryv1sets.NewDestinationSet()
		mockReporter = mock_reporting.NewMockReporter(ctrl)
		mockDecorator = mock_trafficpolicy.NewMockTrafficPolicyDestinationRuleDecorator(ctrl)
		destinationRuleTranslator = destinationrule.NewTranslator(
			settings,
			nil,
			mockClusterDomainRegistry,
			mockDecoratorFactory,
			destinations,
		)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should translate respecting default mTLS Settings", func() {
		settings.Spec = settingsv1.SettingsSpec{
			Mtls: &v1.TrafficPolicySpec_Policy_MTLS{
				Istio: &v1.TrafficPolicySpec_Policy_MTLS_Istio{
					TlsMode: v1.TrafficPolicySpec_Policy_MTLS_Istio_ISTIO_MUTUAL,
				},
			},
		}

		destination := &discoveryv1.Destination{
			ObjectMeta: metav1.ObjectMeta{
				Name: "traffic-target",
			},
			Spec: discoveryv1.DestinationSpec{
				Type: &discoveryv1.DestinationSpec_KubeService_{
					KubeService: &discoveryv1.DestinationSpec_KubeService{
						Ref: &skv2corev1.ClusterObjectRef{
							Name:        "traffic-target",
							Namespace:   "traffic-target-namespace",
							ClusterName: "traffic-target-cluster",
						},
					},
				},
			},
			Status: discoveryv1.DestinationStatus{
				AppliedTrafficPolicies: []*v1.AppliedTrafficPolicy{
					{
						Ref: &skv2corev1.ObjectRef{
							Name:      "tp-1",
							Namespace: "tp-namespace-1",
						},
						Spec: &v1.TrafficPolicySpec{},
					},
				},
				RequiredSubsets: []*discoveryv1.RequiredSubsets{
					{
						TrafficPolicyRef: &skv2corev1.ObjectRef{
							Name:      "another-tp",
							Namespace: "tp-namespace-1",
						},
						TrafficShift: &v1.TrafficPolicySpec_Policy_MultiDestination{
							Destinations: []*v1.WeightedDestination{
								{
									DestinationType: &v1.WeightedDestination_KubeService{
										KubeService: &v1.WeightedDestination_KubeDestination{
											Subset: map[string]string{"foo": "bar", "version": "v1"},
										},
									},
								},
							},
						},
					},
				},
			},
		}

		destinations.Insert(&discoveryv1.Destination{
			ObjectMeta: metav1.ObjectMeta{
				Name: "another-traffic-target",
			},
			Spec: discoveryv1.DestinationSpec{
				Type: &discoveryv1.DestinationSpec_KubeService_{
					KubeService: &discoveryv1.DestinationSpec_KubeService{
						Ref: &skv2corev1.ClusterObjectRef{
							Name:        "another-traffic-target",
							Namespace:   "traffic-target-namespace",
							ClusterName: "traffic-target-cluster",
						},
					},
				},
			},
			Status: discoveryv1.DestinationStatus{
				AppliedTrafficPolicies: []*v1.AppliedTrafficPolicy{
					{
						Ref: &skv2corev1.ObjectRef{
							Name:      "another-tp",
							Namespace: "tp-namespace-1",
						},
						Spec: &v1.TrafficPolicySpec{
							Policy: &v1.TrafficPolicySpec_Policy{
								TrafficShift: &v1.TrafficPolicySpec_Policy_MultiDestination{
									Destinations: []*v1.WeightedDestination{
										{
											DestinationType: &v1.WeightedDestination_KubeService{
												KubeService: &v1.WeightedDestination_KubeDestination{
													// original service
													Name:        destination.Spec.GetKubeService().GetRef().Name,
													Namespace:   destination.Spec.GetKubeService().GetRef().Namespace,
													ClusterName: destination.Spec.GetKubeService().GetRef().ClusterName,

													Subset: map[string]string{"foo": "bar", "version": "v1"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		})

		mockDecoratorFactory.
			EXPECT().
			MakeDecorators(decorators.Parameters{
				ClusterDomains: mockClusterDomainRegistry,
				Snapshot:       in,
			}).
			Return([]decorators.Decorator{mockDecorator})

		mockClusterDomainRegistry.
			EXPECT().
			GetDestinationFQDN(destination.Spec.GetKubeService().Ref.ClusterName, destination.Spec.GetKubeService().Ref).
			Return("local-hostname")

		initializedDestinatonRule := &networkingv1alpha3.DestinationRule{
			ObjectMeta: metautils.TranslatedObjectMeta(
				destination.Spec.GetKubeService().Ref,
				destination.Annotations,
			),
			Spec: networkingv1alpha3spec.DestinationRule{
				Host: "local-hostname",
				TrafficPolicy: &networkingv1alpha3spec.TrafficPolicy{
					Tls: &networkingv1alpha3spec.ClientTLSSettings{
						Mode: networkingv1alpha3spec.ClientTLSSettings_ISTIO_MUTUAL,
					},
				},
				Subsets: []*networkingv1alpha3spec.Subset{
					{
						Name:   "foo-bar-version-v1",
						Labels: map[string]string{"foo": "bar", "version": "v1"},
					},
				},
			},
		}

		mockDecorator.
			EXPECT().
			ApplyTrafficPolicyToDestinationRule(
				destination.Status.AppliedTrafficPolicies[0],
				destination,
				&initializedDestinatonRule.Spec,
				gomock.Any(),
			).
			Return(nil)

		destinationRule := destinationRuleTranslator.Translate(ctx, in, destination, nil, mockReporter)
		Expect(destinationRule).To(Equal(initializedDestinatonRule))
	})

	It("should output DestinationRule for federated Destination", func() {
		destinationRuleTranslator = destinationrule.NewTranslator(
			settings,
			nil,
			mockClusterDomainRegistry,
			mockDecoratorFactory,
			destinations,
		)

		sourceMeshInstallation := &discoveryv1.MeshInstallation{
			Cluster: "source-cluster",
		}

		settings.Spec = settingsv1.SettingsSpec{
			Mtls: &v1.TrafficPolicySpec_Policy_MTLS{
				Istio: &v1.TrafficPolicySpec_Policy_MTLS_Istio{
					TlsMode: v1.TrafficPolicySpec_Policy_MTLS_Istio_ISTIO_MUTUAL,
				},
			},
		}

		destination := &discoveryv1.Destination{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "traffic-target",
				Namespace: "gloo-mesh",
			},
			Spec: discoveryv1.DestinationSpec{
				Type: &discoveryv1.DestinationSpec_KubeService_{
					KubeService: &discoveryv1.DestinationSpec_KubeService{
						Ref: &skv2corev1.ClusterObjectRef{
							Name:        "traffic-target",
							Namespace:   "traffic-target-namespace",
							ClusterName: "traffic-target-clustername",
						},
					},
				},
			},
			Status: discoveryv1.DestinationStatus{
				AppliedTrafficPolicies: []*v1.AppliedTrafficPolicy{
					{
						Ref: &skv2corev1.ObjectRef{
							Name:      "tp-1",
							Namespace: "tp-namespace-1",
						},
						Spec: &v1.TrafficPolicySpec{
							SourceSelector: []*commonv1.WorkloadSelector{
								{
									KubeWorkloadMatcher: &commonv1.WorkloadSelector_KubeWorkloadMatcher{
										Clusters: []string{sourceMeshInstallation.Cluster},
									},
								},
							},
						},
					},
				},
				RequiredSubsets: []*discoveryv1.RequiredSubsets{
					{
						TrafficPolicyRef: &skv2corev1.ObjectRef{
							Name:      "tp-1",
							Namespace: "tp-namespace-1",
						},
						TrafficShift: &v1.TrafficPolicySpec_Policy_MultiDestination{
							Destinations: []*v1.WeightedDestination{
								{
									DestinationType: &v1.WeightedDestination_KubeService{
										KubeService: &v1.WeightedDestination_KubeDestination{
											Name:        "traffic-target",
											Namespace:   "traffic-target-namespace",
											ClusterName: "traffic-target-clustername",
											Subset:      map[string]string{"k1": "v1"},
											Port:        9080,
										},
									},
								},
							},
						},
					},
					{
						TrafficPolicyRef: &skv2corev1.ObjectRef{
							Name:      "tp-2",
							Namespace: "tp-namespace-2",
						},
						TrafficShift: &v1.TrafficPolicySpec_Policy_MultiDestination{
							Destinations: []*v1.WeightedDestination{
								{
									DestinationType: &v1.WeightedDestination_KubeService{
										KubeService: &v1.WeightedDestination_KubeDestination{
											Name:        "traffic-target",
											Namespace:   "traffic-target-namespace",
											ClusterName: "traffic-target-clustername",
											Subset:      map[string]string{"k2": "v2"},
											Port:        9080,
										},
									},
								},
							},
						},
					},
				},
			},
		}

		mockDecoratorFactory.
			EXPECT().
			MakeDecorators(decorators.Parameters{
				ClusterDomains: mockClusterDomainRegistry,
				Snapshot:       in,
			}).
			Return([]decorators.Decorator{mockDecorator})

		mockClusterDomainRegistry.
			EXPECT().
			GetDestinationFQDN(sourceMeshInstallation.Cluster, destination.Spec.GetKubeService().Ref).
			Return("global-hostname")

		expectedDestinatonRule := &networkingv1alpha3.DestinationRule{
			ObjectMeta: metautils.FederatedObjectMeta(
				destination.Spec.GetKubeService().Ref,
				sourceMeshInstallation,
				destination.Annotations,
			),
			Spec: networkingv1alpha3spec.DestinationRule{
				Host: "global-hostname",
				TrafficPolicy: &networkingv1alpha3spec.TrafficPolicy{
					Tls: &networkingv1alpha3spec.ClientTLSSettings{
						Mode: networkingv1alpha3spec.ClientTLSSettings_ISTIO_MUTUAL,
					},
				},
				Subsets: routeutils.MakeDestinationRuleSubsets(destination.Status.RequiredSubsets),
			},
		}

		mockDecorator.
			EXPECT().
			ApplyTrafficPolicyToDestinationRule(
				destination.Status.AppliedTrafficPolicies[0],
				destination,
				gomock.Any(),
				gomock.Any(),
			).
			DoAndReturn(func(
				_ *v1.AppliedTrafficPolicy,
				service *discoveryv1.Destination,
				destinationRuleSpec *networkingv1alpha3spec.DestinationRule,
				_ decorators.RegisterField,
			) error {
				// sort subsets for deterministic comparison
				sort.Slice(destinationRuleSpec.Subsets, func(i, j int) bool {
					return destinationRuleSpec.Subsets[i].Name < destinationRuleSpec.Subsets[j].Name
				})

				Expect(proto.Equal(destinationRuleSpec, &expectedDestinatonRule.Spec)).To(BeTrue())
				return nil
			})

		destinationRule := destinationRuleTranslator.Translate(ctx, in, destination, sourceMeshInstallation, mockReporter)

		// sort subsets for deterministic comparison
		sort.Slice(destinationRule.Spec.Subsets, func(i, j int) bool {
			return destinationRule.Spec.Subsets[i].Name < destinationRule.Spec.Subsets[j].Name
		})

		Expect(destinationRule).To(Equal(expectedDestinatonRule))
	})

	It("should output DestinationRule with keepalive for federated Destination with appropriate inputs", func() {
		destinationRuleTranslator = destinationrule.NewTranslator(
			settings,
			nil,
			mockClusterDomainRegistry,
			mockDecoratorFactory,
			destinations,
		)

		sourceMeshInstallation := &discoveryv1.MeshInstallation{
			Cluster: "source-cluster",
		}

		settings.Spec = settingsv1.SettingsSpec{
			Mtls: &v1.TrafficPolicySpec_Policy_MTLS{
				Istio: &v1.TrafficPolicySpec_Policy_MTLS_Istio{
					TlsMode: v1.TrafficPolicySpec_Policy_MTLS_Istio_ISTIO_MUTUAL,
				},
			},
		}

		destination := &discoveryv1.Destination{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "traffic-target",
				Namespace: "gloo-mesh",
			},
			Spec: discoveryv1.DestinationSpec{
				Type: &discoveryv1.DestinationSpec_KubeService_{
					KubeService: &discoveryv1.DestinationSpec_KubeService{
						Ref: &skv2corev1.ClusterObjectRef{
							Name:        "traffic-target",
							Namespace:   "traffic-target-namespace",
							ClusterName: "traffic-target-clustername",
						},
					},
				},
			},
			Status: discoveryv1.DestinationStatus{
				AppliedFederation: &discoveryv1.DestinationStatus_AppliedFederation{
					TcpKeepalive: &commonv1.TCPKeepalive{
						Probes: 1,
						Time: &duration.Duration{
							Seconds: 2,
							Nanos:   3,
						},
						Interval: &duration.Duration{
							Seconds: 4,
							Nanos:   5,
						},
					},
				},
				AppliedTrafficPolicies: []*v1.AppliedTrafficPolicy{
					{
						Ref: &skv2corev1.ObjectRef{
							Name:      "tp-1",
							Namespace: "tp-namespace-1",
						},
						Spec: &v1.TrafficPolicySpec{
							SourceSelector: []*commonv1.WorkloadSelector{
								{
									KubeWorkloadMatcher: &commonv1.WorkloadSelector_KubeWorkloadMatcher{
										Clusters: []string{sourceMeshInstallation.Cluster},
									},
								},
							},
						},
					},
				},
				RequiredSubsets: []*discoveryv1.RequiredSubsets{
					{
						TrafficPolicyRef: &skv2corev1.ObjectRef{
							Name:      "tp-1",
							Namespace: "tp-namespace-1",
						},
						TrafficShift: &v1.TrafficPolicySpec_Policy_MultiDestination{
							Destinations: []*v1.WeightedDestination{
								{
									DestinationType: &v1.WeightedDestination_KubeService{
										KubeService: &v1.WeightedDestination_KubeDestination{
											Name:        "traffic-target",
											Namespace:   "traffic-target-namespace",
											ClusterName: "traffic-target-clustername",
											Subset:      map[string]string{"k1": "v1"},
											Port:        9080,
										},
									},
								},
							},
						},
					},
					{
						TrafficPolicyRef: &skv2corev1.ObjectRef{
							Name:      "tp-2",
							Namespace: "tp-namespace-2",
						},
						TrafficShift: &v1.TrafficPolicySpec_Policy_MultiDestination{
							Destinations: []*v1.WeightedDestination{
								{
									DestinationType: &v1.WeightedDestination_KubeService{
										KubeService: &v1.WeightedDestination_KubeDestination{
											Name:        "traffic-target",
											Namespace:   "traffic-target-namespace",
											ClusterName: "traffic-target-clustername",
											Subset:      map[string]string{"k2": "v2"},
											Port:        9080,
										},
									},
								},
							},
						},
					},
				},
			},
		}

		mockDecoratorFactory.
			EXPECT().
			MakeDecorators(decorators.Parameters{
				ClusterDomains: mockClusterDomainRegistry,
				Snapshot:       in,
			}).
			Return([]decorators.Decorator{mockDecorator})

		mockClusterDomainRegistry.
			EXPECT().
			GetDestinationFQDN(sourceMeshInstallation.Cluster, destination.Spec.GetKubeService().Ref).
			Return("global-hostname")

		expectedDestinatonRule := &networkingv1alpha3.DestinationRule{
			ObjectMeta: metautils.FederatedObjectMeta(
				destination.Spec.GetKubeService().Ref,
				sourceMeshInstallation,
				destination.Annotations,
			),
			Spec: networkingv1alpha3spec.DestinationRule{
				Host: "global-hostname",
				TrafficPolicy: &networkingv1alpha3spec.TrafficPolicy{
					Tls: &networkingv1alpha3spec.ClientTLSSettings{
						Mode: networkingv1alpha3spec.ClientTLSSettings_ISTIO_MUTUAL,
					},
					ConnectionPool: &networkingv1alpha3spec.ConnectionPoolSettings{
						Tcp: &networkingv1alpha3spec.ConnectionPoolSettings_TCPSettings{
							TcpKeepalive: &networkingv1alpha3spec.ConnectionPoolSettings_TCPSettings_TcpKeepalive{
								Probes: 1,
								Time: &types.Duration{
									Seconds: 2,
									Nanos:   3,
								},
								Interval: &types.Duration{
									Seconds: 4,
									Nanos:   5,
								},
							},
						},
					},
				},

				Subsets: routeutils.MakeDestinationRuleSubsets(destination.Status.RequiredSubsets),
			},
		}

		mockDecorator.
			EXPECT().
			ApplyTrafficPolicyToDestinationRule(
				destination.Status.AppliedTrafficPolicies[0],
				destination,
				gomock.Any(),
				gomock.Any(),
			).
			DoAndReturn(func(
				_ *v1.AppliedTrafficPolicy,
				service *discoveryv1.Destination,
				destinationRuleSpec *networkingv1alpha3spec.DestinationRule,
				_ decorators.RegisterField,
			) error {
				return nil
			})

		destinationRule := destinationRuleTranslator.Translate(ctx, in, destination, sourceMeshInstallation, mockReporter)

		// sort subsets for deterministic comparison
		sort.Slice(destinationRule.Spec.Subsets, func(i, j int) bool {
			return destinationRule.Spec.Subsets[i].Name < destinationRule.Spec.Subsets[j].Name
		})

		Expect(destinationRule).To(Equal(expectedDestinatonRule))
	})

	It("should report error if translated DestinationRule applies to host already configured by existing DestinationRule", func() {
		settings.Spec = settingsv1.SettingsSpec{}

		existingDestinationRules := v1alpha3sets.NewDestinationRuleSet(
			// user-supplied, should yield conflict error
			&networkingv1alpha3.DestinationRule{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "user-provided-dr",
					Namespace:   "foo",
					ClusterName: "traffic-target-cluster",
				},
				Spec: networkingv1alpha3spec.DestinationRule{
					Host: "*-hostname",
				},
			},
		)

		destination := &discoveryv1.Destination{
			ObjectMeta: metav1.ObjectMeta{
				Name: "traffic-target",
			},
			Spec: discoveryv1.DestinationSpec{
				Type: &discoveryv1.DestinationSpec_KubeService_{
					KubeService: &discoveryv1.DestinationSpec_KubeService{
						Ref: &skv2corev1.ClusterObjectRef{
							Name:        "traffic-target",
							Namespace:   "traffic-target-namespace",
							ClusterName: "traffic-target-cluster",
						},
					},
				},
			},
			Status: discoveryv1.DestinationStatus{
				AppliedTrafficPolicies: []*v1.AppliedTrafficPolicy{
					{
						Ref: &skv2corev1.ObjectRef{
							Name:      "tp-1",
							Namespace: "tp-namespace-1",
						},
						Spec: &v1.TrafficPolicySpec{
							SourceSelector: []*commonv1.WorkloadSelector{
								{
									KubeWorkloadMatcher: &commonv1.WorkloadSelector_KubeWorkloadMatcher{
										Clusters: []string{"traffic-target-cluster"},
									},
								},
							},
							Policy: &v1.TrafficPolicySpec_Policy{
								OutlierDetection: &v1.TrafficPolicySpec_Policy_OutlierDetection{
									ConsecutiveErrors: 5,
								},
							},
						},
					},
				},
			},
		}

		mockDecoratorFactory.
			EXPECT().
			MakeDecorators(decorators.Parameters{
				ClusterDomains: mockClusterDomainRegistry,
				Snapshot:       in,
			}).
			Return([]decorators.Decorator{mockDecorator})

		mockClusterDomainRegistry.
			EXPECT().
			GetDestinationFQDN(destination.Spec.GetKubeService().Ref.ClusterName, destination.Spec.GetKubeService().Ref).
			Return("local-hostname")

		mockDecorator.
			EXPECT().
			ApplyTrafficPolicyToDestinationRule(
				destination.Status.AppliedTrafficPolicies[0],
				destination,
				gomock.Any(),
				gomock.Any(),
			).
			DoAndReturn(func(
				appliedPolicy *v1.AppliedTrafficPolicy,
				service *discoveryv1.Destination,
				output *networkingv1alpha3spec.DestinationRule,
				registerField decorators.RegisterField,
			) error {
				output.TrafficPolicy.OutlierDetection = &networkingv1alpha3spec.OutlierDetection{
					Consecutive_5XxErrors: &types.UInt32Value{Value: 5},
				}
				return nil
			})

		mockReporter.
			EXPECT().
			ReportTrafficPolicyToDestination(
				destination,
				destination.Status.AppliedTrafficPolicies[0].Ref,
				gomock.Any()).
			DoAndReturn(func(destination *discoveryv1.Destination, trafficPolicy ezkube.ResourceId, err error) {
				Expect(err).To(testutils.HaveInErrorChain(
					eris.Errorf("Unable to translate AppliedTrafficPolicies to DestinationRule, applies to host %s that is already configured by the existing DestinationRule %s",
						"local-hostname",
						sets.Key(existingDestinationRules.List()[0])),
				),
				)
			})

		destinationRuleTranslator = destinationrule.NewTranslator(
			settings,
			existingDestinationRules,
			mockClusterDomainRegistry,
			mockDecoratorFactory,
			destinations,
		)

		_ = destinationRuleTranslator.Translate(ctx, in, destination, nil, mockReporter)
	})
})
