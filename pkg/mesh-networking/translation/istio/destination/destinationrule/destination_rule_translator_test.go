package destinationrule_test

import (
	"context"

	"github.com/gogo/protobuf/types"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	v1alpha3sets "github.com/solo-io/external-apis/pkg/api/istio/networking.istio.io/v1alpha3/sets"
	commonv1alpha2 "github.com/solo-io/gloo-mesh/pkg/api/common.mesh.gloo.solo.io/v1alpha2"
	discoveryv1alpha2 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2"
	v1alpha2sets "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2/sets"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1alpha2"
	settingsv1alpha2 "github.com/solo-io/gloo-mesh/pkg/api/settings.mesh.gloo.solo.io/v1alpha2"
	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	mock_reporting "github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting/mocks"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators"
	mock_decorators "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators/mocks"
	mock_trafficpolicy "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators/mocks"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators/trafficshift"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/destination/destinationrule"
	mock_hostutils "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/hostutils/mocks"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/metautils"
	"github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
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
		destinations              v1alpha2sets.DestinationSet
		mockReporter              *mock_reporting.MockReporter
		mockDecorator             *mock_trafficpolicy.MockTrafficPolicyDestinationRuleDecorator
		destinationRuleTranslator destinationrule.Translator
		in                        input.LocalSnapshot
		ctx                       = context.TODO()
		settings                  = &settingsv1alpha2.Settings{
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
		destinations = v1alpha2sets.NewDestinationSet()
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
		settings.Spec = settingsv1alpha2.SettingsSpec{
			Mtls: &v1alpha2.TrafficPolicySpec_Policy_MTLS{
				Istio: &v1alpha2.TrafficPolicySpec_Policy_MTLS_Istio{
					TlsMode: v1alpha2.TrafficPolicySpec_Policy_MTLS_Istio_ISTIO_MUTUAL,
				},
			},
		}

		destination := &discoveryv1alpha2.Destination{
			ObjectMeta: metav1.ObjectMeta{
				Name: "traffic-target",
			},
			Spec: discoveryv1alpha2.DestinationSpec{
				Type: &discoveryv1alpha2.DestinationSpec_KubeService_{
					KubeService: &discoveryv1alpha2.DestinationSpec_KubeService{
						Ref: &v1.ClusterObjectRef{
							Name:        "traffic-target",
							Namespace:   "traffic-target-namespace",
							ClusterName: "traffic-target-cluster",
						},
					},
				},
			},
			Status: discoveryv1alpha2.DestinationStatus{
				AppliedTrafficPolicies: []*discoveryv1alpha2.DestinationStatus_AppliedTrafficPolicy{
					{
						Ref: &v1.ObjectRef{
							Name:      "tp-1",
							Namespace: "tp-namespace-1",
						},
						Spec: &v1alpha2.TrafficPolicySpec{},
					},
				},
			},
		}

		destinations.Insert(&discoveryv1alpha2.Destination{
			ObjectMeta: metav1.ObjectMeta{
				Name: "another-traffic-target",
			},
			Spec: discoveryv1alpha2.DestinationSpec{
				Type: &discoveryv1alpha2.DestinationSpec_KubeService_{
					KubeService: &discoveryv1alpha2.DestinationSpec_KubeService{
						Ref: &v1.ClusterObjectRef{
							Name:        "another-traffic-target",
							Namespace:   "traffic-target-namespace",
							ClusterName: "traffic-target-cluster",
						},
					},
				},
			},
			Status: discoveryv1alpha2.DestinationStatus{
				AppliedTrafficPolicies: []*discoveryv1alpha2.DestinationStatus_AppliedTrafficPolicy{
					{
						Ref: &v1.ObjectRef{
							Name:      "another-tp",
							Namespace: "tp-namespace-1",
						},
						Spec: &v1alpha2.TrafficPolicySpec{
							Policy: &v1alpha2.TrafficPolicySpec_Policy{
								TrafficShift: &v1alpha2.TrafficPolicySpec_Policy_MultiDestination{
									Destinations: []*v1alpha2.TrafficPolicySpec_Policy_MultiDestination_WeightedDestination{
										{
											DestinationType: &v1alpha2.TrafficPolicySpec_Policy_MultiDestination_WeightedDestination_KubeService{
												KubeService: &v1alpha2.TrafficPolicySpec_Policy_MultiDestination_WeightedDestination_KubeDestination{
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

	It("should not output DestinationRule when DestinationRule has no effect", func() {
		settings.Spec = settingsv1alpha2.SettingsSpec{
			Mtls: &v1alpha2.TrafficPolicySpec_Policy_MTLS{
				Istio: &v1alpha2.TrafficPolicySpec_Policy_MTLS_Istio{
					TlsMode: v1alpha2.TrafficPolicySpec_Policy_MTLS_Istio_DISABLE,
				},
			},
		}

		trafficTarget := &discoveryv1alpha2.Destination{
			ObjectMeta: metav1.ObjectMeta{
				Name: "traffic-target",
			},
			Spec: discoveryv1alpha2.DestinationSpec{
				Type: &discoveryv1alpha2.DestinationSpec_KubeService_{
					KubeService: &discoveryv1alpha2.DestinationSpec_KubeService{
						Ref: &v1.ClusterObjectRef{
							Name:        "traffic-target",
							Namespace:   "traffic-target-namespace",
							ClusterName: "traffic-target-cluster",
						},
					},
				},
			},
			Status: discoveryv1alpha2.DestinationStatus{
				AppliedTrafficPolicies: []*discoveryv1alpha2.DestinationStatus_AppliedTrafficPolicy{
					{
						Ref: &v1.ObjectRef{
							Name:      "tp-1",
							Namespace: "tp-namespace-1",
						},
						Spec: &v1alpha2.TrafficPolicySpec{
							SourceSelector: []*commonv1alpha2.WorkloadSelector{
								{
									Clusters: []string{"traffic-target-cluster"},
								},
							},
						},
					},
					{
						Ref: &v1.ObjectRef{
							Name:      "tp-1",
							Namespace: "tp-namespace-1",
						},
						Spec: &v1alpha2.TrafficPolicySpec{
							SourceSelector: []*commonv1alpha2.WorkloadSelector{
								{
									Clusters: []string{"different-cluster"},
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
			GetDestinationFQDN(trafficTarget.Spec.GetKubeService().Ref.ClusterName, trafficTarget.Spec.GetKubeService().Ref).
			Return("local-hostname")

		initializedDestinatonRule := &networkingv1alpha3.DestinationRule{
			ObjectMeta: metautils.TranslatedObjectMeta(
				trafficTarget.Spec.GetKubeService().Ref,
				trafficTarget.Annotations,
			),
			Spec: networkingv1alpha3spec.DestinationRule{
				Host: "local-hostname",
				TrafficPolicy: &networkingv1alpha3spec.TrafficPolicy{
					Tls: &networkingv1alpha3spec.ClientTLSSettings{
						Mode: networkingv1alpha3spec.ClientTLSSettings_DISABLE,
					},
				},
			},
		}

		mockDecorator.
			EXPECT().
			ApplyTrafficPolicyToDestinationRule(
				trafficTarget.Status.AppliedTrafficPolicies[0],
				trafficTarget,
				&initializedDestinatonRule.Spec,
				gomock.Any(),
			).
			Return(nil)

		destinationRule := destinationRuleTranslator.Translate(ctx, in, trafficTarget, nil, mockReporter)
		Expect(destinationRule).To(BeNil())
	})

	It("should output DestinationRule for federated Destination", func() {
		destinations = v1alpha2sets.NewDestinationSet(
			&discoveryv1alpha2.Destination{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "target-1",
					Namespace: "ns",
				},
				Status: discoveryv1alpha2.DestinationStatus{
					AppliedTrafficPolicies: []*discoveryv1alpha2.DestinationStatus_AppliedTrafficPolicy{
						{
							Ref: &v1.ObjectRef{
								Name:      "tp-1",
								Namespace: "tp-namespace-1",
							},
							Spec: &v1alpha2.TrafficPolicySpec{
								Policy: &v1alpha2.TrafficPolicySpec_Policy{
									TrafficShift: &v1alpha2.TrafficPolicySpec_Policy_MultiDestination{
										Destinations: []*v1alpha2.TrafficPolicySpec_Policy_MultiDestination_WeightedDestination{
											{
												DestinationType: &v1alpha2.TrafficPolicySpec_Policy_MultiDestination_WeightedDestination_KubeService{
													KubeService: &v1alpha2.TrafficPolicySpec_Policy_MultiDestination_WeightedDestination_KubeDestination{
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
							},
						},
					},
				},
			},
			&discoveryv1alpha2.Destination{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "target-2",
					Namespace: "ns",
				},
				Status: discoveryv1alpha2.DestinationStatus{
					AppliedTrafficPolicies: []*discoveryv1alpha2.DestinationStatus_AppliedTrafficPolicy{
						{
							Ref: &v1.ObjectRef{
								Name:      "tp-2",
								Namespace: "tp-namespace-2",
							},
							Spec: &v1alpha2.TrafficPolicySpec{
								Policy: &v1alpha2.TrafficPolicySpec_Policy{
									TrafficShift: &v1alpha2.TrafficPolicySpec_Policy_MultiDestination{
										Destinations: []*v1alpha2.TrafficPolicySpec_Policy_MultiDestination_WeightedDestination{
											{
												DestinationType: &v1alpha2.TrafficPolicySpec_Policy_MultiDestination_WeightedDestination_KubeService{
													KubeService: &v1alpha2.TrafficPolicySpec_Policy_MultiDestination_WeightedDestination_KubeDestination{
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
					},
				},
			},
		)

		destinationRuleTranslator = destinationrule.NewTranslator(
			settings,
			nil,
			mockClusterDomainRegistry,
			mockDecoratorFactory,
			destinations,
		)

		sourceMeshInstallation := &discoveryv1alpha2.MeshSpec_MeshInstallation{
			Cluster: "source-cluster",
		}

		settings.Spec = settingsv1alpha2.SettingsSpec{
			Mtls: &v1alpha2.TrafficPolicySpec_Policy_MTLS{
				Istio: &v1alpha2.TrafficPolicySpec_Policy_MTLS_Istio{
					TlsMode: v1alpha2.TrafficPolicySpec_Policy_MTLS_Istio_ISTIO_MUTUAL,
				},
			},
		}

		trafficTarget := &discoveryv1alpha2.Destination{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "traffic-target",
				Namespace: "gloo-mesh",
			},
			Spec: discoveryv1alpha2.DestinationSpec{
				Type: &discoveryv1alpha2.DestinationSpec_KubeService_{
					KubeService: &discoveryv1alpha2.DestinationSpec_KubeService{
						Ref: &v1.ClusterObjectRef{
							Name:        "traffic-target",
							Namespace:   "traffic-target-namespace",
							ClusterName: "traffic-target-clustername",
						},
					},
				},
			},
			Status: discoveryv1alpha2.DestinationStatus{
				AppliedTrafficPolicies: []*discoveryv1alpha2.DestinationStatus_AppliedTrafficPolicy{
					{
						Ref: &v1.ObjectRef{
							Name:      "tp-1",
							Namespace: "tp-namespace-1",
						},
						Spec: &v1alpha2.TrafficPolicySpec{
							SourceSelector: []*commonv1alpha2.WorkloadSelector{
								{
									Clusters: []string{sourceMeshInstallation.Cluster},
								},
							},
						},
					},
				},
			},
		}

		federatedClusterLabels := trafficshift.MakeFederatedSubsetLabel(trafficTarget.Spec.GetKubeService().Ref.ClusterName)

		mockDecoratorFactory.
			EXPECT().
			MakeDecorators(decorators.Parameters{
				ClusterDomains: mockClusterDomainRegistry,
				Snapshot:       in,
			}).
			Return([]decorators.Decorator{mockDecorator})

		mockClusterDomainRegistry.
			EXPECT().
			GetDestinationFQDN(sourceMeshInstallation.Cluster, trafficTarget.Spec.GetKubeService().Ref).
			Return("global-hostname")

		expectedDestinatonRule := &networkingv1alpha3.DestinationRule{
			ObjectMeta: metautils.FederatedObjectMeta(
				trafficTarget.Spec.GetKubeService().Ref,
				sourceMeshInstallation,
				trafficTarget.Annotations,
			),
			Spec: networkingv1alpha3spec.DestinationRule{
				Host: "global-hostname",
				TrafficPolicy: &networkingv1alpha3spec.TrafficPolicy{
					Tls: &networkingv1alpha3spec.ClientTLSSettings{
						Mode: networkingv1alpha3spec.ClientTLSSettings_ISTIO_MUTUAL,
					},
				},
				Subsets: []*networkingv1alpha3spec.Subset{
					{
						Name:   "k1-v1",
						Labels: federatedClusterLabels,
					},
					{
						Name:   "k2-v2",
						Labels: federatedClusterLabels,
					},
				},
			},
		}

		mockDecorator.
			EXPECT().
			ApplyTrafficPolicyToDestinationRule(
				trafficTarget.Status.AppliedTrafficPolicies[0],
				trafficTarget,
				&expectedDestinatonRule.Spec,
				gomock.Any(),
			).
			Return(nil)

		destinationRule := destinationRuleTranslator.Translate(ctx, in, trafficTarget, sourceMeshInstallation, mockReporter)
		Expect(destinationRule).To(Equal(expectedDestinatonRule))
	})

	It("should report error if translated DestinationRule applies to host already configured by existing DestinationRule", func() {
		settings.Spec = settingsv1alpha2.SettingsSpec{}

		existingDestinationRules := v1alpha3sets.NewDestinationRuleSet(
			// user-supplied, should yield conflict error
			&networkingv1alpha3.DestinationRule{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "user-provided-dr",
					Namespace: "foo",
				},
				Spec: networkingv1alpha3spec.DestinationRule{
					Host: "*-hostname",
				},
			},
		)

		destination := &discoveryv1alpha2.Destination{
			ObjectMeta: metav1.ObjectMeta{
				Name: "traffic-target",
			},
			Spec: discoveryv1alpha2.DestinationSpec{
				Type: &discoveryv1alpha2.DestinationSpec_KubeService_{
					KubeService: &discoveryv1alpha2.DestinationSpec_KubeService{
						Ref: &v1.ClusterObjectRef{
							Name:        "traffic-target",
							Namespace:   "traffic-target-namespace",
							ClusterName: "traffic-target-cluster",
						},
					},
				},
			},
			Status: discoveryv1alpha2.DestinationStatus{
				AppliedTrafficPolicies: []*discoveryv1alpha2.DestinationStatus_AppliedTrafficPolicy{
					{
						Ref: &v1.ObjectRef{
							Name:      "tp-1",
							Namespace: "tp-namespace-1",
						},
						Spec: &v1alpha2.TrafficPolicySpec{
							SourceSelector: []*commonv1alpha2.WorkloadSelector{
								{
									Clusters: []string{"traffic-target-cluster"},
								},
							},
							Policy: &v1alpha2.TrafficPolicySpec_Policy{
								OutlierDetection: &v1alpha2.TrafficPolicySpec_Policy_OutlierDetection{
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
				appliedPolicy *discoveryv1alpha2.DestinationStatus_AppliedTrafficPolicy,
				service *discoveryv1alpha2.Destination,
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
			DoAndReturn(func(trafficTarget *discoveryv1alpha2.Destination, trafficPolicy ezkube.ResourceId, err error) {
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
