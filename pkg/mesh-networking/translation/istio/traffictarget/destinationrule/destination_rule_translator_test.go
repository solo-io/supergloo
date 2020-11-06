package destinationrule_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	discoveryv1alpha2 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2"
	v1alpha2sets "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2/sets"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1alpha2"
	v1alpha2sets2 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1alpha2/sets"
	settingsv1alpha2 "github.com/solo-io/gloo-mesh/pkg/api/settings.mesh.gloo.solo.io/v1alpha2"
	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	mock_reporting "github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting/mocks"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators"
	mock_decorators "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators/mocks"
	mock_trafficpolicy "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators/mocks"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators/trafficshift"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/traffictarget/destinationrule"
	mock_hostutils "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/hostutils/mocks"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/metautils"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	networkingv1alpha3spec "istio.io/api/networking/v1alpha3"
	networkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("DestinationRuleTranslator", func() {
	var (
		ctrl                      *gomock.Controller
		mockClusterDomainRegistry *mock_hostutils.MockClusterDomainRegistry
		mockDecoratorFactory      *mock_decorators.MockFactory
		trafficTargets            v1alpha2sets.TrafficTargetSet
		failoverServices          v1alpha2sets2.FailoverServiceSet
		mockReporter              *mock_reporting.MockReporter
		mockDecorator             *mock_trafficpolicy.MockTrafficPolicyDestinationRuleDecorator
		destinationRuleTranslator destinationrule.Translator
		in                        input.Snapshot
		ctx                       = context.TODO()
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockClusterDomainRegistry = mock_hostutils.NewMockClusterDomainRegistry(ctrl)
		mockDecoratorFactory = mock_decorators.NewMockFactory(ctrl)
		trafficTargets = v1alpha2sets.NewTrafficTargetSet()
		failoverServices = v1alpha2sets2.NewFailoverServiceSet()
		mockReporter = mock_reporting.NewMockReporter(ctrl)
		mockDecorator = mock_trafficpolicy.NewMockTrafficPolicyDestinationRuleDecorator(ctrl)
		destinationRuleTranslator = destinationrule.NewTranslator(
			mockClusterDomainRegistry,
			mockDecoratorFactory,
			trafficTargets,
			failoverServices,
		)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should translate respecting default mTLS Settings", func() {
		in = input.NewInputSnapshotManualBuilder("").
			AddSettings(settingsv1alpha2.SettingsSlice{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      defaults.DefaultSettingsName,
						Namespace: defaults.DefaultPodNamespace,
					},
					Spec: settingsv1alpha2.SettingsSpec{
						Mtls: &v1alpha2.TrafficPolicySpec_MTLS{
							Istio: &v1alpha2.TrafficPolicySpec_MTLS_Istio{
								TlsMode: v1alpha2.TrafficPolicySpec_MTLS_Istio_ISTIO_MUTUAL,
							},
						},
					},
				},
			}).
			Build()

		trafficTarget := &discoveryv1alpha2.TrafficTarget{
			ObjectMeta: metav1.ObjectMeta{
				Name: "traffic-target",
			},
			Spec: discoveryv1alpha2.TrafficTargetSpec{
				Type: &discoveryv1alpha2.TrafficTargetSpec_KubeService_{
					KubeService: &discoveryv1alpha2.TrafficTargetSpec_KubeService{
						Ref: &v1.ClusterObjectRef{
							Name:        "traffic-target",
							Namespace:   "traffic-target-namespace",
							ClusterName: "traffic-target-cluster",
						},
					},
				},
			},
			Status: discoveryv1alpha2.TrafficTargetStatus{
				AppliedTrafficPolicies: []*discoveryv1alpha2.TrafficTargetStatus_AppliedTrafficPolicy{
					{
						Ref: &v1.ObjectRef{
							Name:      "tp-1",
							Namespace: "tp-namespace-1",
						},
						Spec: &v1alpha2.TrafficPolicySpec{},
					},
					{
						Ref: &v1.ObjectRef{
							Name:      "tp-2",
							Namespace: "tp-namespace-2",
						},
						Spec: &v1alpha2.TrafficPolicySpec{},
					},
				},
			},
		}

		trafficTargets.Insert(&discoveryv1alpha2.TrafficTarget{
			ObjectMeta: metav1.ObjectMeta{
				Name: "another-traffic-target",
			},
			Spec: discoveryv1alpha2.TrafficTargetSpec{
				Type: &discoveryv1alpha2.TrafficTargetSpec_KubeService_{
					KubeService: &discoveryv1alpha2.TrafficTargetSpec_KubeService{
						Ref: &v1.ClusterObjectRef{
							Name:        "another-traffic-target",
							Namespace:   "traffic-target-namespace",
							ClusterName: "traffic-target-cluster",
						},
					},
				},
			},
			Status: discoveryv1alpha2.TrafficTargetStatus{
				AppliedTrafficPolicies: []*discoveryv1alpha2.TrafficTargetStatus_AppliedTrafficPolicy{
					{
						Ref: &v1.ObjectRef{
							Name:      "another-tp",
							Namespace: "tp-namespace-1",
						},
						Spec: &v1alpha2.TrafficPolicySpec{
							TrafficShift: &v1alpha2.TrafficPolicySpec_MultiDestination{
								Destinations: []*v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination{
									{
										DestinationType: &v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination_KubeService{KubeService: &v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination_KubeDestination{
											// original service
											Name:        trafficTarget.Spec.GetKubeService().GetRef().Name,
											Namespace:   trafficTarget.Spec.GetKubeService().GetRef().Namespace,
											ClusterName: trafficTarget.Spec.GetKubeService().GetRef().ClusterName,

											Subset: map[string]string{"foo": "bar"},
										}},
									},
									{
										DestinationType: &v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination_FailoverService{
											FailoverService: &v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination_FailoverServiceDestination{
												// original service
												Name:      "fs1",
												Namespace: "fs1-ns",
												Subset:    map[string]string{"boo": "baz"},
											}},
									},
								},
							},
						},
					},
				},
			},
		})

		failoverServices.Insert(&v1alpha2.FailoverService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "fs1",
				Namespace: "fs1-ns",
			},
			Spec: v1alpha2.FailoverServiceSpec{
				BackingServices: []*v1alpha2.FailoverServiceSpec_BackingService{
					{
						BackingServiceType: &v1alpha2.FailoverServiceSpec_BackingService_KubeService{
							KubeService: &v1.ClusterObjectRef{
								Name:        "traffic-target",
								Namespace:   "traffic-target-namespace",
								ClusterName: "traffic-target-cluster",
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
			GetDestinationServiceFQDN(trafficTarget.Spec.GetKubeService().Ref.ClusterName, trafficTarget.Spec.GetKubeService().Ref).
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
						Mode: networkingv1alpha3spec.ClientTLSSettings_ISTIO_MUTUAL,
					},
				},
				Subsets: []*networkingv1alpha3spec.Subset{
					{
						Name:   "foo-bar",
						Labels: map[string]string{"foo": "bar"},
					},
					{
						Name:   "boo-baz",
						Labels: map[string]string{"boo": "baz"},
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
		mockDecorator.
			EXPECT().
			ApplyTrafficPolicyToDestinationRule(
				trafficTarget.Status.AppliedTrafficPolicies[1],
				trafficTarget,
				&initializedDestinatonRule.Spec,
				gomock.Any(),
			).
			Return(nil)

		destinationRule := destinationRuleTranslator.Translate(ctx, in, trafficTarget, nil, mockReporter)
		Expect(destinationRule).To(Equal(initializedDestinatonRule))
	})

	It("should not output DestinationRule when DestinationRule has no effect", func() {
		in = input.NewInputSnapshotManualBuilder("").
			AddSettings(settingsv1alpha2.SettingsSlice{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      defaults.DefaultSettingsName,
						Namespace: defaults.DefaultPodNamespace,
					},
					Spec: settingsv1alpha2.SettingsSpec{
						Mtls: &v1alpha2.TrafficPolicySpec_MTLS{
							Istio: &v1alpha2.TrafficPolicySpec_MTLS_Istio{
								TlsMode: v1alpha2.TrafficPolicySpec_MTLS_Istio_DISABLE,
							},
						},
					},
				},
			}).
			Build()

		trafficTarget := &discoveryv1alpha2.TrafficTarget{
			ObjectMeta: metav1.ObjectMeta{
				Name: "traffic-target",
			},
			Spec: discoveryv1alpha2.TrafficTargetSpec{
				Type: &discoveryv1alpha2.TrafficTargetSpec_KubeService_{
					KubeService: &discoveryv1alpha2.TrafficTargetSpec_KubeService{
						Ref: &v1.ClusterObjectRef{
							Name:        "traffic-target",
							Namespace:   "traffic-target-namespace",
							ClusterName: "traffic-target-cluster",
						},
					},
				},
			},
			Status: discoveryv1alpha2.TrafficTargetStatus{
				AppliedTrafficPolicies: []*discoveryv1alpha2.TrafficTargetStatus_AppliedTrafficPolicy{
					{
						Ref: &v1.ObjectRef{
							Name:      "tp-1",
							Namespace: "tp-namespace-1",
						},
						Spec: &v1alpha2.TrafficPolicySpec{
							SourceSelector: []*v1alpha2.WorkloadSelector{
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
							SourceSelector: []*v1alpha2.WorkloadSelector{
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
			GetDestinationServiceFQDN(trafficTarget.Spec.GetKubeService().Ref.ClusterName, trafficTarget.Spec.GetKubeService().Ref).
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

	It("should output DestinationRule for federated TrafficTarget", func() {
		trafficTargets = v1alpha2sets.NewTrafficTargetSet(
			&discoveryv1alpha2.TrafficTarget{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "target-1",
					Namespace: "ns",
				},
				Status: discoveryv1alpha2.TrafficTargetStatus{
					AppliedTrafficPolicies: []*discoveryv1alpha2.TrafficTargetStatus_AppliedTrafficPolicy{
						{
							Ref: &v1.ObjectRef{
								Name:      "tp-1",
								Namespace: "tp-namespace-1",
							},
							Spec: &v1alpha2.TrafficPolicySpec{
								TrafficShift: &v1alpha2.TrafficPolicySpec_MultiDestination{
									Destinations: []*v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination{
										{
											DestinationType: &v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination_KubeService{
												KubeService: &v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination_KubeDestination{
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
			&discoveryv1alpha2.TrafficTarget{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "target-2",
					Namespace: "ns",
				},
				Status: discoveryv1alpha2.TrafficTargetStatus{
					AppliedTrafficPolicies: []*discoveryv1alpha2.TrafficTargetStatus_AppliedTrafficPolicy{
						{
							Ref: &v1.ObjectRef{
								Name:      "tp-2",
								Namespace: "tp-namespace-2",
							},
							Spec: &v1alpha2.TrafficPolicySpec{
								TrafficShift: &v1alpha2.TrafficPolicySpec_MultiDestination{
									Destinations: []*v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination{
										{
											DestinationType: &v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination_KubeService{
												KubeService: &v1alpha2.TrafficPolicySpec_MultiDestination_WeightedDestination_KubeDestination{
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
		)

		destinationRuleTranslator = destinationrule.NewTranslator(
			mockClusterDomainRegistry,
			mockDecoratorFactory,
			trafficTargets,
			failoverServices,
		)

		sourceMeshInstallation := &discoveryv1alpha2.MeshSpec_MeshInstallation{
			Cluster: "source-cluster",
		}

		in = input.NewInputSnapshotManualBuilder("").
			AddSettings(settingsv1alpha2.SettingsSlice{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      defaults.DefaultSettingsName,
						Namespace: defaults.DefaultPodNamespace,
					},
					Spec: settingsv1alpha2.SettingsSpec{
						Mtls: &v1alpha2.TrafficPolicySpec_MTLS{
							Istio: &v1alpha2.TrafficPolicySpec_MTLS_Istio{
								TlsMode: v1alpha2.TrafficPolicySpec_MTLS_Istio_ISTIO_MUTUAL,
							},
						},
					},
				},
			}).
			Build()

		trafficTarget := &discoveryv1alpha2.TrafficTarget{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "traffic-target",
				Namespace: "service-mesh-hub",
			},
			Spec: discoveryv1alpha2.TrafficTargetSpec{
				Type: &discoveryv1alpha2.TrafficTargetSpec_KubeService_{
					KubeService: &discoveryv1alpha2.TrafficTargetSpec_KubeService{
						Ref: &v1.ClusterObjectRef{
							Name:        "traffic-target",
							Namespace:   "traffic-target-namespace",
							ClusterName: "traffic-target-clustername",
						},
					},
				},
			},
			Status: discoveryv1alpha2.TrafficTargetStatus{
				AppliedTrafficPolicies: []*discoveryv1alpha2.TrafficTargetStatus_AppliedTrafficPolicy{
					{
						Ref: &v1.ObjectRef{
							Name:      "tp-1",
							Namespace: "tp-namespace-1",
						},
						Spec: &v1alpha2.TrafficPolicySpec{
							SourceSelector: []*v1alpha2.WorkloadSelector{
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
			GetDestinationServiceFQDN(sourceMeshInstallation.Cluster, trafficTarget.Spec.GetKubeService().Ref).
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
})
